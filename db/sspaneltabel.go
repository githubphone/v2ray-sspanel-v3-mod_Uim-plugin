package db

import (
	"time"
)

type AliveIP struct {
	Nodeid   uint   `gorm:"column:nodeid"`
	Userid   uint   `gorm:"column:userid"`
	Ip       string `gorm:"column:ip"`
	Datetime int64  `gorm:"column:datetime"`
}

func (*AliveIP) TableName() string {
	return "alive_ip"
}

type Speedtest struct {
	Nodeid           uint   `gorm:"column:nodeid"`
	Datetime         int64  `gorm:"column:datetime"`
	Telecomping      string `gorm:"column:telecomping"`
	Telecomeupload   string `gorm:"column:telecomeupload"`
	Telecomedownload string `gorm:"column:telecomedownload"`
	Unicomping       string `gorm:"column:unicomping"`
	Unicomupload     string `gorm:"column:unicomupload"`
	Unicomdownload   string `gorm:"column:unicomdownload"`
	Cmccping         string `gorm:"column:cmccping"`
	Cmccupload       string `gorm:"column:cmccupload"`
	Cmccdownload     string `gorm:"column:cmccdownload"`
}

func (*Speedtest) TableName() string {
	return "speedtest"
}

type SSNode struct {
	Id                    uint    `gorm:"column:id"`
	Name                  string  `gorm:"column:name"`
	Type                  uint    `gorm:"column:type"`
	Server                string  `gorm:"column:server"`
	Method                string  `gorm:"column:method"`
	Info                  string  `gorm:"column:info"`
	Status                string  `gorm:"column:status"`
	Sort                  uint    `gorm:"column:sort"`
	CustomMethod          uint    `gorm:"column:custom_method"`
	TrafficRate           float64 `gorm:"column:traffic_rate"`
	NodeClass             uint    `gorm:"column:node_class"`
	NodeSpeedlimit        float64 `gorm:"column:node_speedlimit"`
	NodeConnector         uint    `gorm:"column:node_connector"`
	NodeBrandwith         uint64  `gorm:"column:node_bandwidth"`
	NodeBrandwithLimit    uint64  `gorm:"column:node_bandwidth_limit"`
	BrandwithlimitRestday uint    `gorm:"column:bandwidthlimit_resetday"`
	NodeHeartbeat         int64   `gorm:"column:node_heartbeat"`
	NodeIp                string  `gorm:"column:node_ip"`
	NodeGroup             uint    `gorm:"column:node_group"`
}

func (*SSNode) TableName() string {
	return "ss_node"
}

type SSNodeInfo struct {
	NodeId  uint    `gorm:"column:node_id"`
	Uptime  float64 `gorm:"column:uptime"`
	Load    string  `gorm:"column:load"`
	Logtime int64   `gorm:"column:log_time"`
}

func (*SSNodeInfo) TableName() string {
	return "ss_node_info"
}

type SsNodeOnlineLog struct {
	NodeId     uint  `gorm:"column:node_id"`
	OnlineUser uint  `gorm:"column:online_user"`
	Logtime    int64 `gorm:"column:log_time"`
}

func (*SsNodeOnlineLog) TableName() string {
	return "ss_node_online_log"
}

type User struct {
	ID             uint      `gorm:"column:id"`
	Email          string    `gorm:"column:email"`
	Passwd         string    `gorm:"column:passwd"`
	Method         string    `gorm:"column:method"`
	Port           uint16    `gorm:"column:port"`
	NodeSpeedlimit float64   `gorm:"column:node_speedlimit"`
	Upload         int64     `gorm:"column:u"`
	Download       int64     `gorm:"column:d"`
	Enable         uint      `gorm:"column:enable"`
	ExpireTime     uint      `gorm:"column:expire_time"`
	IsAdmin        uint      `gorm:"column:is_admin"`
	Class          uint      `gorm:"column:class"`
	ExpireIN       time.Time `gorm:"colum:expire_in"`
	TransferEnable int64     `gorm:"column:transfer_enable"`
	Time           int64     `gorm:"column:t"`
	Obfs           string    `gorm:"column:obfs"`
	Protocol       string    `gorm:"column:protocol"`
}

func (*User) TableName() string {
	return "user"
}

type UserTrafficLog struct {
	Userid   uint    `gorm:"column:user_id"`
	Upload   uint64  `gorm:"column:u"`
	Download uint64  `gorm:"column:d"`
	Nodeid   uint    `gorm:"column:node_id"`
	Rate     float64 `gorm:"column:rate"`
	Logtime  int64   `gorm:"column:log_time"`
	Traffic  string  `gorm:"column:traffic"`
}

func (*UserTrafficLog) TableName() string {
	return "user_traffic_log"
}

type Relay struct {
	ID           uint   `gorm:"column:id"`
	Userid       uint   `gorm:"column:user_id"`
	SourceNodeId uint   `gorm:"column:source_node_id"`
	DistNodeId   int    `gorm:"column:dist_node_id"`
	DistIp       string `gorm:"column:dist_ip"`
	Port         uint   `gorm:"column:port"`
	Priority     uint   `gorm:"column:priority"`
}

func (*Relay) TableName() string {
	return "relay"
}
