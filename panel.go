package v2ray_sspanel_v3_mod_Uim_plugin

import (
	"fmt"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/Manager"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/client"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/config"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/webapi"
	"github.com/robfig/cron"
	"google.golang.org/grpc"
	"reflect"
	"runtime"
)

type Panel struct {
	db              *webapi.Webapi
	manager         *Manager.Manager
	speedtestClient speedtest.Client
}

func NewPanel(gRPCConn *grpc.ClientConn, db *webapi.Webapi, cfg *config.Config) (*Panel, error) {
	opts := speedtest.NewOpts()
	speedtestClient := speedtest.NewClient(opts)
	var newpanel = Panel{
		speedtestClient: speedtestClient,
		db:              db,
		manager: &Manager.Manager{
			HandlerServiceClient: client.NewHandlerServiceClient(gRPCConn, "MAIN_INBOUND"),
			StatsServiceClient:   client.NewStatsServiceClient(gRPCConn),
			NodeID:               cfg.NodeID,
			CheckRate:            cfg.CheckRate,
			SpeedTestCheckRate:   cfg.SpeedTestCheckRate,
			CurrentNodeInfo:      &model.NodeInfo{},
			NextNodeInfo:         &model.NodeInfo{},
			Users:                map[string]model.UserModel{},
			UserToBeMoved:        map[string]model.UserModel{},
			UserToBeAdd:          map[string]model.UserModel{},
		},
	}
	return &newpanel, nil
}

func (p *Panel) Start() {
	doFunc := func() {
		if err := p.do(); err != nil {
			newError("panel#do").Base(err).AtError().WriteToLog()
		}
		// Explicitly triggering GC to remove garbage
		runtime.GC()
	}
	doFunc()

	speedTestFunc := func() {
		result, err := speedtest.GetSpeedtest(p.speedtestClient)
		if err != nil {
			newError("panel#speedtest").Base(err).AtError().WriteToLog()
		}
		newError(result).AtInfo().WriteToLog()
		if p.db.UploadSpeedTest(p.manager.NodeID, result) {
			newError("succesfully upload speedtest result").AtInfo().WriteToLog()
		} else {
			newError("failed to upload speedtest result").AtInfo().WriteToLog()
		}
		// Explicitly triggering GC to remove garbage
		runtime.GC()
	}
	c := cron.New()
	err := c.AddFunc(fmt.Sprintf("@every %ds", p.manager.CheckRate), doFunc)
	if err != nil {
		fatal(err)
	}
	if p.manager.SpeedTestCheckRate > 0 {
		newErrorf("@every %ds", p.manager.SpeedTestCheckRate).AtInfo().WriteToLog()
		err = c.AddFunc(fmt.Sprintf("@every %dh", p.manager.SpeedTestCheckRate), speedTestFunc)
		if err != nil {
			newError("Can't add speed test into cron").AtWarning().WriteToLog()
		}
	}
	c.Start()
	c.Run()
}

func (p *Panel) do() error {
	p.updateManager()
	p.updateThroughout()
	return nil
}
func (p *Panel) initial() {
	newError("initial system").AtWarning().WriteToLog()
	p.manager.RemoveInbound()
	p.manager.CopyUsers()
	p.manager.UpdataUsers()
	p.manager.CurrentNodeInfo = &model.NodeInfo{}
	p.manager.NextNodeInfo = &model.NodeInfo{}
	p.manager.UserToBeAdd = map[string]model.UserModel{}
	p.manager.UserToBeMoved = map[string]model.UserModel{}
	p.manager.Users = map[string]model.UserModel{}

}

func (p *Panel) updateManager() {
	newNodeinfo, err := p.db.GetNodeInfo(p.manager.NodeID)
	if err != nil {
		newError(err).AtWarning().WriteToLog()
		p.initial()
		return
	}
	if newNodeinfo.Ret != 1 {
		newError(newNodeinfo.Data).AtWarning().WriteToLog()
		p.initial()
		return
	}
	newErrorf("old node info %s ", p.manager.NextNodeInfo.Server_raw).AtInfo().WriteToLog()
	newErrorf("new node info %s", newNodeinfo.Data.Server_raw).AtInfo().WriteToLog()
	if p.manager.NextNodeInfo.Server_raw != newNodeinfo.Data.Server_raw {
		p.manager.NextNodeInfo = newNodeinfo.Data
		if err = p.manager.UpdateServer(); err != nil {
			newError(err).AtWarning().WriteToLog()
		}
		p.manager.UserChanged = true
	}
	users, err := p.db.GetALLUsers(p.manager.NextNodeInfo)
	if err != nil {
		newError(err).AtDebug().WriteToLog()
	}
	newError("now begin to check users").AtInfo().WriteToLog()
	current_user := p.manager.GetUsers()
	// remove user by prefixed_id
	for key, _ := range current_user {
		_, ok := users.Data[key]
		if !ok {
			p.manager.Remove(key)
			newErrorf("need to remove client: %s.", key).AtInfo().WriteToLog()
		}
	}
	// add users
	for key, value := range users.Data {
		current, ok := current_user[key]
		if !ok {
			p.manager.Add(value)
			newErrorf("need to add user email %s", key).AtInfo().WriteToLog()
		} else {
			if !reflect.DeepEqual(value, current) {
				p.manager.Remove(key)
				p.manager.Add(value)
				newErrorf("need to add user email %s due to method or password changed", key).AtInfo().WriteToLog()
			}
		}

	}

	if p.manager.UserChanged {
		p.manager.UserChanged = false
		newErrorf("Before Update, Current Users %d need to be add %d need to be romved %d", len(p.manager.Users),
			len(p.manager.UserToBeAdd), len(p.manager.UserToBeMoved)).AtWarning().WriteToLog()
		p.manager.UpdataUsers()
		newErrorf("After Update, Current Users %d need to be add %d need to be romved %d", len(p.manager.Users),
			len(p.manager.UserToBeAdd), len(p.manager.UserToBeMoved)).AtWarning().WriteToLog()
		p.manager.CurrentNodeInfo = p.manager.NextNodeInfo
	} else {
		newError("check ports finished. No need to update ").AtInfo().WriteToLog()
	}

}
func (p *Panel) updateThroughout() {
	current_user := p.manager.GetUsers()
	usertraffic := []model.UserTrafficLog{}
	userIPS := []model.UserOnLineIP{}
	for _, value := range current_user {
		current_upload, err := p.manager.StatsServiceClient.GetUserUplink(value.Email)
		current_download, err := p.manager.StatsServiceClient.GetUserDownlink(value.Email)
		if err != nil {
			newError(err).AtDebug().WriteToLog()
		}
		if current_upload+current_download > 0 {

			newErrorf("USER %s has use %d", value.Email, current_upload+current_download).AtDebug().WriteToLog()
			usertraffic = append(usertraffic, model.UserTrafficLog{
				UserID:   value.UserID,
				Downlink: current_download,
				Uplink:   current_upload,
			})
			current_user_ips, err := p.manager.StatsServiceClient.GetUserIPs(value.Email)
			if current_upload+current_download > 1024 {
				if err != nil {
					newError(err).AtDebug().WriteToLog()
				}
				for index := range current_user_ips {
					userIPS = append(userIPS, model.UserOnLineIP{
						UserId: value.UserID,
						Ip:     current_user_ips[index],
					})
				}

			}
		}
	}
	if p.db.UpLoadUserTraffics(p.manager.NodeID, usertraffic) {
		newErrorf("Successfully upload %d users traffics", len(usertraffic)).AtInfo().WriteToLog()
	} else {
		newError("update trafic failed").AtDebug().WriteToLog()
	}
	if p.db.UpLoadOnlineIps(p.manager.NodeID, userIPS) {
		newErrorf("Successfully upload %d ips", len(userIPS)).AtInfo().WriteToLog()
	} else {
		newError("update trafic failed").AtDebug().WriteToLog()
	}
	if p.db.UploadSystemLoad(p.manager.NodeID) {
		newError("Uploaded systemLoad successfully").AtInfo().WriteToLog()
	} else {
		newError("Failed to uploaded systemLoad ").AtDebug().WriteToLog()
	}
}
