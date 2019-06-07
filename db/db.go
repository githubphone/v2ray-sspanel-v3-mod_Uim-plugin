package db

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/imroc/req"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"regexp"
	"strconv"
	"strings"
)

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

type NodeinfoResponse struct {
	Ret  uint            `json:"ret"`
	Data *model.NodeInfo `json:"data"`
}
type PostResponse struct {
	Ret  uint   `json:"ret"`
	Data string `json:"data"`
}
type UsersResponse struct {
	Ret  uint              `json:"ret"`
	Data []model.UserModel `json:"data"`
}
type AllUsers struct {
	Ret  uint
	Data map[string]model.UserModel
}

type DisNodenfoResponse struct {
	Ret  uint                 `json:"ret"`
	Data []*model.DisNodeInfo `json:"data"`
}
type AuthResponse struct {
	Ret   uint   `json:"ret"`
	Token string `json:"token"`
}

var id2string = map[uint]string{
	0: "server_address",
	1: "port",
	2: "alterid",
	3: "protocol",
	4: "protocol_param",
	5: "path",
	6: "host",
	7: "inside_port",
	8: "server",
}
var maps = map[string]interface{}{
	"server_address": "",
	"port":           "",
	"alterid":        "2",
	"protocol":       "tcp",
	"protocol_param": "",
	"path":           "",
	"host":           "",
	"inside_port":    "",
	"server":         "",
}

func getMD5(data string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(data))
	cipherStr := md5Ctx.Sum(nil)
	current_md5 := hex.EncodeToString(cipherStr)
	return current_md5
}
func get_mu_host(id uint, md5 string, MU_REGEX string, MU_SUFFIX string) string {
	regex_text := MU_REGEX
	regex_text = strings.Replace(regex_text, "%id", fmt.Sprintf("%d", id), -1)
	regex_text = strings.Replace(regex_text, "%suffix", MU_SUFFIX, -1)
	regex := regexp.MustCompile(`%-?[1-9]\d*m`)
	for _, item := range regex.FindAllString(regex_text, -1) {
		regex_num := strings.Replace(item, "%", "", -1)
		regex_num = strings.Replace(regex_num, "m", "", -1)
		md5_length, _ := strconv.ParseInt(regex_num, 10, 0)
		if md5_length < 0 {
			regex_text = strings.Replace(regex_text, item, md5[32+md5_length:], -1)
		} else {
			regex_text = strings.Replace(regex_text, item, md5[:md5_length], -1)
		}
	}
	return regex_text
}

type Db interface {
	GetApi(url string, params map[string]interface{}) (*req.Resp, error)

	GetNodeInfo(nodeid uint) (*NodeinfoResponse, error)

	GetDisNodeInfo(nodeid uint) (*DisNodenfoResponse, error)

	GetALLUsers(info *model.NodeInfo) (*AllUsers, error)

	Post(url string, params map[string]interface{}, data map[string]interface{}) (*req.Resp, error)

	UploadSystemLoad(nodeid uint) bool

	UpLoadUserTraffics(nodeid uint, trafficLog []model.UserTrafficLog) bool
	UploadSpeedTest(nodeid uint, speedresult []speedtest.Speedresult) bool
	UpLoadOnlineIps(nodeid uint, onlineIPS []model.UserOnLineIP) bool
}
