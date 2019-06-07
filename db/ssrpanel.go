package db

import (
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/jinzhu/gorm"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/utility"
	"math"
	"strconv"
	"time"
)

type SSRpanel struct {
	Db        *gorm.DB
	MU_REGEX  string
	MU_SUFFIX string
}

func (api *SSRpanel) GetApi(url string, params map[string]interface{}) (*req.Resp, error) {
	return nil, nil
}

func (api *SSRpanel) GetNodeInfo(nodeid uint) (*NodeinfoResponse, error) {
	var response = NodeinfoResponse{Data: &model.NodeInfo{}}

	node := SSRNode{}
	err := api.Db.First(&node, nodeid).Error
	if err != nil {
		return nil, newError("nodeinfo not find,please check your nodeid or mysql setting").Base(err)
	}

	response.Ret = 1
	sever_raw, _ := json.Marshal(node)
	response.Data.Server_raw = string(sever_raw)
	if node.Type == 1 {
		response.Data.Sort = 0
	} else if node.Type == 2 {
		response.Data.Sort = 11
	} else {
		return nil, newError("not implement type ")
	}
	response.Data.NodeSpeedlimit = node.Bandwidth
	response.Data.TrafficRate = node.TrafficRate
	response.Data.NodeID = node.ID
	response.Data.Server = map[string]interface{}{
		"server_address": node.Server,
		"port":           fmt.Sprintf("%d", node.V2Port),
		"alterid":        fmt.Sprintf("%d", node.V2AlterId),
		"protocol":       node.V2Type,
		"protocol_param": node.V2Net,
		"path":           node.V2Path,
		"host":           node.V2Host,
		"inside_port":    node.V2rayInsiderPort,
		"server":         node.Server,
	}
	if node.V2Tls == 1 {
		response.Data.Server["protocol_param"] = "tls"
	}
	return &response, nil
}

func (api *SSRpanel) GetDisNodeInfo(nodeid uint) (*DisNodenfoResponse, error) {
	var response = DisNodenfoResponse{}
	return &response, nil
}

func (api *SSRpanel) GetALLUsers(info *model.NodeInfo) (*AllUsers, error) {
	sort := info.Sort
	var prifix string
	var allusers = AllUsers{
		Data: map[string]model.UserModel{},
	}
	if sort == 0 {
		prifix = "SS_"
	} else {
		prifix = "Vmess_"
		if info.Server["protocol"] == "tcp" {
			prifix += "tcp_"
		} else if info.Server["protocol"] == "ws" {
			if info.Server["protocol_param"] != "" {
				prifix += "ws_" + info.Server["protocol_param"].(string) + "_"
			} else {
				prifix += "ws_" + "none" + "_"
			}
		} else if info.Server["protocol"] == "kcp" {
			if info.Server["protocol_param"] != "" {
				prifix += "kcp_" + info.Server["protocol_param"].(string) + "_"
			} else {
				prifix += "kcp_" + "none" + "_"
			}
		}
	}
	var response = UsersResponse{}
	users := make([]SSRUserModel, 0)
	err := api.Db.Select("id, vmess_id, username").Where("enable = 1 AND u + d < transfer_enable").Find(&users).Error
	if err != nil {
		return nil, err
	}

	response.Ret = 1

	var filterd_user []model.UserModel
	for index := range users {
		user := users[index]
		filterd_user = append(filterd_user, model.UserModel{
			UserID:         user.ID,
			Uuid:           user.VmessID,
			Email:          user.Email,
			Passwd:         user.Passwd,
			Method:         user.Method,
			Port:           uint16(user.Port),
			NodeSpeedlimit: uint(user.SpeedLimitPerUser / 125000),
			Obfs:           user.Obfs,
			Protocol:       user.Protocol,
		})

	}

	response.Data = filterd_user

	for index := range response.Data {
		// 按照node 限速来调整用户限速
		if info.NodeSpeedlimit != 0 {
			if info.NodeSpeedlimit > response.Data[index].NodeSpeedlimit && response.Data[index].NodeSpeedlimit != 0 {
			} else if response.Data[index].NodeSpeedlimit == 0 || response.Data[index].NodeSpeedlimit > info.NodeSpeedlimit {
				response.Data[index].NodeSpeedlimit = info.NodeSpeedlimit
			}
		}
		// 接受到的是 Mbps， 然后我们的一个buffer 是2048byte， 差不多61个
		response.Data[index].Rate = uint32(response.Data[index].NodeSpeedlimit * 62)

		if info.Server["alterid"].(string) == "" {
			response.Data[index].AlterId = 2
		} else {
			alterid, err := strconv.ParseUint(info.Server["alterid"].(string), 10, 0)
			if err == nil {
				response.Data[index].AlterId = uint32(alterid)
			}
		}
		user := response.Data[index]
		response.Data[index].Muhost = get_mu_host(user.UserID, getMD5(fmt.Sprintf("%d%s%s%s%s", user.UserID, user.Passwd, user.Method, user.Obfs, user.Protocol)), api.MU_REGEX, api.MU_SUFFIX)
		key := prifix + response.Data[index].Email + fmt.Sprintf("Rate_%d_AlterID_%d_Method_%s_Passwd_%s_Port_%d_Obfs_%s_Protocol_%s", response.Data[index].Rate,
			response.Data[index].AlterId, response.Data[index].Method, response.Data[index].Passwd, response.Data[index].Port, response.Data[index].Obfs, response.Data[index].Protocol,
		)
		response.Data[index].PrefixedId = key
		allusers.Data[key] = response.Data[index]
	}
	return &allusers, nil
}

func (api *SSRpanel) Post(url string, params map[string]interface{}, data map[string]interface{}) (*req.Resp, error) {

	return nil, nil
}

func (api *SSRpanel) UploadSystemLoad(nodeid uint) bool {
	uptime, _ := strconv.ParseFloat(utility.GetSystemUptime(), 64)
	err := api.Db.Create(&SSRNodeInfo{
		NodeID: nodeid,
		Load:   utility.GetSystemLoad(),
		Uptime: uptime,
	}).Error
	if err != nil {
		return false
	}
	return true
}

func (api *SSRpanel) flowAutoShow(value float64) string {
	value_float := value
	kb := 1024.0
	mb := 1048576.0
	gb := 1073741824.0
	tb := gb * 1024.0
	pb := tb * 1024.0
	if math.Abs(value_float) > pb {
		return fmt.Sprintf("%.2fPB", math.Round(value_float/pb))
	} else if math.Abs(value_float) > tb {
		return fmt.Sprintf("%.2fTB", math.Round(value_float/tb))
	} else if math.Abs(value_float) > gb {
		return fmt.Sprintf("%.2fGB", math.Round(value_float/gb))
	} else if math.Abs(value_float) > mb {
		return fmt.Sprintf("%.2fMB", math.Round(value_float/mb))
	} else if math.Abs(value_float) > kb {
		return fmt.Sprintf("%.2fKB", math.Round(value_float/kb))
	} else {
		return fmt.Sprintf("%.2fB", math.Round(value_float))
	}
}

func (api *SSRpanel) UpLoadUserTraffics(nodeid uint, trafficLog []model.UserTrafficLog) bool {
	var ssrnode SSRNode
	err := api.Db.First(&ssrnode, "id = ?", nodeid).Error
	if err != nil {
		return false
	}
	var this_time_total_bandwidth uint64 = 0
	if len(trafficLog) > 0 {
		for index := range trafficLog {
			traffic := trafficLog[index]

			var user SSRUserModel
			err := api.Db.First(&user, "id = ?", traffic.UserID).Error

			if err != nil {
				continue
			}
			user.Time = time.Now().Unix()
			user.Upload += int64(float64(traffic.Uplink) * ssrnode.TrafficRate)
			user.Download += int64(float64(traffic.Downlink) * ssrnode.TrafficRate)
			this_time_total_bandwidth += traffic.Uplink + traffic.Downlink
			api.Db.Save(&user)

			api.Db.Save(&SSRUserTrafficLog{
				UserID:   traffic.UserID,
				Uplink:   traffic.Uplink,
				Downlink: traffic.Downlink,
				NodeID:   nodeid,
				Rate:     ssrnode.TrafficRate,
				LogTime:  time.Now().Unix(),
				Traffic:  api.flowAutoShow(float64(traffic.Uplink+traffic.Downlink) * ssrnode.TrafficRate),
			})
		}
	}

	api.Db.Create(&SSRNodeOnlineLog{
		NodeID:     nodeid,
		OnlineUser: len(trafficLog),
		LogTime:    time.Now().Unix(),
	})
	return true
}
func (api *SSRpanel) UploadSpeedTest(nodeid uint, speedresult []speedtest.Speedresult) bool {
	return true
}
func (api *SSRpanel) UpLoadOnlineIps(nodeid uint, onlineIPS []model.UserOnLineIP) bool {
	var ssrnode SSRNode
	err := api.Db.First(&ssrnode, "id = ?", nodeid).Error
	if err != nil {
		return false
	}
	if len(onlineIPS) > 0 {
		for index := range onlineIPS {
			ip := onlineIPS[index]
			api.Db.Create(&SSRNodeIP{
				NodeID:    nodeid,
				UserID:    ip.UserId,
				CreatedAt: time.Now().Unix(),
			})
		}
	}
	return true
}
