package db

import (
	"github.com/jinzhu/gorm"
	"time"
)

type SSRUserModel struct {
	ID                uint
	VmessID           string
	Passwd            string
	Method            string
	Port              int
	SpeedLimitPerUser int
	Email             string `gorm:"column:username"`
	Time              int64  `gorm:"column:t"`
	Upload            int64  `gorm:"column:u"`
	Download          int64  `gorm:"column:d"`
	Obfs              string `gorm:"column:obfs"`
	Protocol          string `gorm:"column:protocol"`
}

func (*SSRUserModel) TableName() string {
	return "user"
}

type SSRUserTrafficLog struct {
	ID       uint `gorm:"primary_key"`
	UserID   uint
	Uplink   uint64 `gorm:"column:u"`
	Downlink uint64 `gorm:"column:d"`
	NodeID   uint
	Rate     float64
	Traffic  string
	LogTime  int64
}

func (l *SSRUserTrafficLog) BeforeCreate(scope *gorm.Scope) error {
	l.LogTime = time.Now().Unix()
	return nil
}

type SSRNodeOnlineLog struct {
	ID         uint `gorm:"primary_key"`
	NodeID     uint
	OnlineUser int
	LogTime    int64
}

func (*SSRNodeOnlineLog) TableName() string {
	return "ss_node_online_log"
}

func (l *SSRNodeOnlineLog) BeforeCreate(scope *gorm.Scope) error {
	l.LogTime = time.Now().Unix()
	return nil
}

type SSRNodeInfo struct {
	ID      uint `gorm:"primary_key"`
	NodeID  uint
	Uptime  float64
	Load    string
	LogTime int64
}

func (*SSRNodeInfo) TableName() string {
	return "ss_node_info"
}

func (l *SSRNodeInfo) BeforeCreate(scope *gorm.Scope) error {
	l.LogTime = time.Now().Unix()
	return nil
}

type SSRNode struct {
	ID               uint `gorm:"primary_key"`
	TrafficRate      float64
	Type             uint   `gorm:"column:type"`
	Method           string `gorm:"column:method"`
	V2AlterId        uint
	V2Port           uint
	V2Net            string
	V2Type           string
	V2Host           string
	V2Path           string
	V2Tls            uint
	Bandwidth        uint
	Ip               string
	V2rayInsiderPort string
	Server           string
}

func (*SSRNode) TableName() string {
	return "ss_node"
}

type SSRNodeIP struct {
	ID        uint `gorm:"primary_key"`
	NodeID    uint
	UserID    uint
	Ip        string
	CreatedAt int64
}

func (*SSRNodeIP) TableName() string {
	return "ss_node_ip"
}
