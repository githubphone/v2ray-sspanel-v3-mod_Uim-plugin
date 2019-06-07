package db

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/imroc/req"
	"github.com/jinzhu/gorm"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/utility"
	"math"
	"strconv"
	"strings"
	"time"
)

type SSpanel struct {
	Db        *gorm.DB
	MU_REGEX  string
	MU_SUFFIX string
}

func (api *SSpanel) GetApi(url string, params map[string]interface{}) (*req.Resp, error) {
	return nil, nil
}

func (api *SSpanel) GetNodeInfo(nodeid uint) (*NodeinfoResponse, error) {
	var response = NodeinfoResponse{Data: &model.NodeInfo{}}
	var nodeinfotable SSNode
	err := api.Db.First(&nodeinfotable, "id = ?", nodeid).Error
	if err != nil {
		return nil, newError("nodeinfo not find,please check your nodeid or mysql setting").Base(err)
	}
	response.Ret = 1
	response.Data.Server_raw = nodeinfotable.Server
	response.Data.Sort = nodeinfotable.Sort
	response.Data.NodeSpeedlimit = uint(nodeinfotable.NodeSpeedlimit)
	response.Data.TrafficRate = nodeinfotable.TrafficRate
	response.Data.NodeID = nodeinfotable.Id
	if response.Data.Server_raw != "" {
		response.Data.Server_raw = strings.ToLower(response.Data.Server_raw)
		data := strings.Split(response.Data.Server_raw, ";")
		var count uint
		count = 0
		for v := range data {
			if len(data[v]) > 0 {
				maps[id2string[count]] = data[v]
			}
			count += 1
		}
		var extraArgues []string
		if len(data) == 6 {
			extraArgues = append(extraArgues, strings.Split(data[5], "|")...)
			for item := range extraArgues {
				data = strings.Split(extraArgues[item], "=")
				if len(data) > 0 {
					if len(data[1]) > 0 {
						maps[data[0]] = data[1]
					}

				}
			}
		}

		if maps["protocol"] == "tls" {
			temp := maps["protocol_param"]
			maps["protocol"] = temp
			maps["protocol_param"] = "tls"
		}
		response.Data.Server = maps
	}
	response.Data.NodeID = nodeid
	return &response, nil
}

func (api *SSpanel) GetDisNodeInfo(nodeid uint) (*DisNodenfoResponse, error) {
	var response = DisNodenfoResponse{}
	var rules []Relay

	err := api.Db.Find(&rules, "source_node_id =?", nodeid).Error
	if err != nil {
		return nil, newError("no relay rule ,please check your nodeid or mysql setting").Base(err)
	}
	response.Ret = 1
	for index := range rules {
		rule := rules[index]
		if rule.DistNodeId == -1 {
			continue
		}
		var dis SSNode
		err = api.Db.First(&dis, "id = ?", rule.DistNodeId).Error
		if err != nil {
			continue
		} else {
			disnode := model.DisNodeInfo{
				Server_raw: dis.Server,
				Sort:       dis.Sort,
				Port:       uint16(rule.Port),
				UserId:     rule.Userid}
			response.Data = append(response.Data, &disnode)
		}
	}
	if len(response.Data) > 0 {
		for _, relayrule := range response.Data {
			relayrule.Server_raw = strings.ToLower(relayrule.Server_raw)
			data := strings.Split(relayrule.Server_raw, ";")
			var count uint
			count = 0
			for v := range data {
				if len(data[v]) > 0 {
					maps[id2string[count]] = data[v]
				}
				count += 1
			}
			var extraArgues []string
			if len(data) == 6 {
				extraArgues = append(extraArgues, strings.Split(data[5], "|")...)
				for item := range extraArgues {
					data = strings.Split(extraArgues[item], "=")
					if len(data) > 0 {
						if len(data[1]) > 0 {
							maps[data[0]] = data[1]
						}

					}
				}
			}

			if maps["protocol"] == "tls" {
				temp := maps["protocol_param"]
				maps["protocol"] = temp
				maps["protocol_param"] = "tls"
			}
			relayrule.Server = maps
		}
	}
	return &response, nil
}

func (api *SSpanel) GetALLUsers(info *model.NodeInfo) (*AllUsers, error) {
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
	var node SSNode
	err := api.Db.First(&node, "id = ?", info.NodeID).Error
	if err != nil {
		return nil, err
	}
	node.NodeHeartbeat = time.Now().Unix()
	api.Db.Save(&node)

	var users []User
	time1 := time.Now()
	location, _ := time.LoadLocation("Asia/Shanghai")
	t_in := time1.In(location).Format("2006-01-02 15:04:05")
	if node.NodeGroup != 0 {
		err := api.Db.Where("class >= ? AND node_group = ?", node.NodeClass, node.NodeGroup).Or("is_admin = 1").Where("enable =1 AND expire_in > ?", t_in).Find(&users).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := api.Db.Where("class >= ?", node.NodeClass).Or("is_admin = 1").Where("enable =1 AND expire_in > ?", t_in).Find(&users).Error
		if err != nil {
			return nil, err
		}
	}
	response.Ret = 1

	var filterd_user []model.UserModel
	for index := range users {
		user := users[index]
		if user.TransferEnable > user.Download+user.Upload {
			filterd_user = append(filterd_user, model.UserModel{
				UserID:         user.ID,
				Uuid:           uuid.NewV3(uuid.NamespaceDNS, fmt.Sprintf("%d|%s", user.ID, user.Passwd)).String(),
				Email:          user.Email,
				Passwd:         user.Passwd,
				Method:         user.Method,
				Port:           user.Port,
				NodeSpeedlimit: uint(user.NodeSpeedlimit),
				Obfs:           user.Obfs,
				Protocol:       user.Protocol,
			})
		}
	}
	if node.NodeBrandwithLimit != 0 {
		if node.NodeBrandwithLimit < node.NodeBrandwith {
			response.Data = []model.UserModel{}
		} else {
			response.Data = filterd_user
		}
	} else {
		response.Data = filterd_user
	}
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

func (api *SSpanel) Post(url string, params map[string]interface{}, data map[string]interface{}) (*req.Resp, error) {

	return nil, nil
}

func (api *SSpanel) UploadSystemLoad(nodeid uint) bool {
	uptime, _ := strconv.ParseFloat(utility.GetSystemUptime(), 64)
	err := api.Db.Create(&SSNodeInfo{
		NodeId:  nodeid,
		Load:    utility.GetSystemLoad(),
		Uptime:  uptime,
		Logtime: time.Now().Unix(),
	}).Error
	if err != nil {
		return false
	}
	return true
}

func (api *SSpanel) flowAutoShow(value float64) string {
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

func (api *SSpanel) UpLoadUserTraffics(nodeid uint, trafficLog []model.UserTrafficLog) bool {
	var ssnode SSNode
	err := api.Db.First(&ssnode, "id = ?", nodeid).Error
	if err != nil {
		return false
	}
	var this_time_total_bandwidth uint64 = 0
	if len(trafficLog) > 0 {
		for index := range trafficLog {
			traffic := trafficLog[index]

			var user User
			err := api.Db.First(&user, "id = ?", traffic.UserID).Error

			if err != nil {
				continue
			}
			user.Time = time.Now().Unix()
			user.Upload += int64(float64(traffic.Uplink) * ssnode.TrafficRate)
			user.Download += int64(float64(traffic.Downlink) * ssnode.TrafficRate)
			this_time_total_bandwidth += traffic.Uplink + traffic.Downlink
			api.Db.Save(&user)

			api.Db.Save(&UserTrafficLog{
				Userid:   traffic.UserID,
				Upload:   traffic.Uplink,
				Download: traffic.Downlink,
				Nodeid:   nodeid,
				Rate:     ssnode.TrafficRate,
				Logtime:  time.Now().Unix(),
				Traffic:  api.flowAutoShow(float64(traffic.Uplink+traffic.Downlink) * ssnode.TrafficRate),
			})
		}
	}
	ssnode.NodeBrandwith += this_time_total_bandwidth
	api.Db.Save(&ssnode)
	api.Db.Create(&SsNodeOnlineLog{
		NodeId:     nodeid,
		OnlineUser: uint(len(trafficLog)),
		Logtime:    time.Now().Unix(),
	})
	return true
}
func (api *SSpanel) UploadSpeedTest(nodeid uint, speedresult []speedtest.Speedresult) bool {
	var ssnode SSNode
	err := api.Db.First(&ssnode, "id = ?", nodeid).Error
	if err != nil {
		return false
	}

	if len(speedresult) > 0 {
		for index := range speedresult {
			speed := speedresult[index]
			api.Db.Create(&Speedtest{
				Nodeid:           nodeid,
				Datetime:         time.Now().Unix(),
				Telecomping:      speed.CTPing,
				Telecomeupload:   speed.CTUpSpeed,
				Telecomedownload: speed.CTDLSpeed,
				Unicomping:       speed.CUPing,
				Unicomupload:     speed.CUUpSpeed,
				Unicomdownload:   speed.CUDLSpeed,
				Cmccping:         speed.CMPing,
				Cmccupload:       speed.CMUpSpeed,
				Cmccdownload:     speed.CMDLSpeed,
			})
		}
	}
	return true
}
func (api *SSpanel) UpLoadOnlineIps(nodeid uint, onlineIPS []model.UserOnLineIP) bool {
	var ssnode SSNode
	err := api.Db.First(&ssnode, "id = ?", nodeid).Error
	if err != nil {
		return false
	}
	if len(onlineIPS) > 0 {
		for index := range onlineIPS {
			ip := onlineIPS[index]
			api.Db.Create(&AliveIP{
				Nodeid:   nodeid,
				Userid:   ip.UserId,
				Ip:       ip.Ip,
				Datetime: time.Now().Unix(),
			})
		}
	}
	return true
}

func (api *SSpanel) CheckAuth(url string, params map[string]interface{}) (*AuthResponse, error) {
	var response = AuthResponse{}
	parm := req.Param{}
	for k, v := range params {
		parm[k] = v
	}
	r, err := req.Get(url, parm)
	if err != nil {
		return nil, err
	} else {
		err = r.ToJSON(&response)
		if err != nil {
			return &response, err
		} else if response.Ret != 1 {
			return nil, err
		}
	}
	return &response, nil
}
