package v2ray_sspanel_v3_mod_Uim_plugin

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/client"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/config"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/db"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"google.golang.org/grpc"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/infra/conf"
	"v2ray.com/core/transport/internet"
)

type UserModelData struct {
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	Passwd  string `json:"password"`
	Method  string `json:"method"`
	Port    string `json:"port"`
	Speed   string `json:"speed"`
	AlterId uint32
	Muhost  string
}
type UserTraffic struct {
	Email    string `json:"email"`
	Upload   string `json:"upload"`
	Download string `json:"download"`
}

func (u *UserModelData) Build() model.UserModel {
	port, _ := strconv.ParseUint(u.Port, 10, 0)
	speed, _ := strconv.ParseUint(u.Speed, 10, 10)
	return model.UserModel{
		Uuid:       u.Uuid,
		Email:      u.Email,
		Passwd:     u.Passwd,
		Method:     u.Method,
		Port:       uint16(port),
		Rate:       uint32(speed * 62),
		PrefixedId: u.Email,
		AlterId:    2,
	}
}

type tlsSettings struct {
	ServerName string `json:"serverName"`
}
type wsSettings struct {
	Path    string                 `json:"path"`
	Headers map[string]interface{} `json:"headers"`
}
type kcpSettings struct {
	Header map[string]interface{}
}
type streamSetting struct {
	Network     string      `json:"network"`
	TlsSettings tlsSettings `json:"tlsSettings"`
	Wssettings  wsSettings  `json:"wsSettings"`
	Kcpsettings kcpSettings `json:"kcpSettings"`
}
type InboundConfig struct {
	GrpcApiPort    uint16        `json:"grpcApiPort"`
	Port           uint16        `json:"port"`
	Protocol       string        `json:"protocol"`
	Tag            string        `json:"tag"`
	StreamSettings streamSetting `json:"streamSettings"`
}
type userAdd struct {
	Users []UserModelData `json:"users"`
}
type userOld struct {
	Email string `json:"email"`
}

type userUpdate struct {
	UsersNew []UserModelData `json:"usersNew"`
	UsersOld []userOld       `json:"userOld"`
}
type userRemove struct {
	Users []userOld `json:"users"`
}
type GoPanel struct {
	sync.Mutex
	Conn                 *websocket.Conn
	speedtestClient      speedtest.Client
	downwithpanel        int
	Key                  string
	HandlerServiceClient *client.HandlerServiceClient
	StatsServiceClient   *client.StatsServiceClient
	RuleServiceClient    *client.RuleServerClient
	CheckRate            int
	SpeedTestCheckRate   int
	Users                map[string]model.UserModel
	Host                 string
	Handshake            bool
	Server               *Serverdata
}

type Serverdata struct {
	Element   string          `json:"element"`
	Operation string          `json:"operation"`
	Users     []UserModelData `json:"users"`
	Value     string          `json:"value"`
	Inbound   InboundConfig   `json:"inboundConfig"`
	NewUsers  []UserModelData `json:"usersNew"`
	OldUsers  []userOld       `json:"usersOld"`
	Key       string
}

func NewGoPanel(gRPCConn *grpc.ClientConn, db db.Db, cfg *config.Config) (*GoPanel, error) {
	opts := speedtest.NewOpts()
	speedtestClient := speedtest.NewClient(opts)
	var newpanel = GoPanel{
		speedtestClient:      speedtestClient,
		downwithpanel:        cfg.DownWithPanel,
		HandlerServiceClient: client.NewHandlerServiceClient(gRPCConn, "MAIN_INBOUND"),
		StatsServiceClient:   client.NewStatsServiceClient(gRPCConn),
		RuleServiceClient:    client.NewRuleServerClient(gRPCConn),
		CheckRate:            cfg.CheckRate,
		SpeedTestCheckRate:   cfg.SpeedTestCheckRate,
		Users:                map[string]model.UserModel{},
		Key:                  cfg.GoPanelKey,
		Host:                 cfg.GoPanelHost,
	}
	return &newpanel, nil
}
func (p *GoPanel) do() error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})
	go func() {
		for {
			if p.Handshake {
				p.updateThroughout()
				time.Sleep(time.Second * time.Duration(p.CheckRate))
			}

		}
	}()
	go func() {
		for {
			if p.Handshake {
				result, err := speedtest.GetSpeedtest(p.speedtestClient)
				if err != nil {
					newError("panel#speedtest").Base(err).AtError().WriteToLog()
				}
				newError(result).AtInfo().WriteToLog()
				uploadresult := map[string]interface{}{"element": "speed", "operators": []map[string]interface{}{
					{"brand": "ctcc",
						"ping":    strings.Split(result[0].CTPing, " ")[0],
						"downloa": strings.Split(result[0].CTDLSpeed, " ")[0],
						"upload":  strings.Split(result[0].CTUpSpeed, " ")[0],
					}, {"brand": "cmcc",
						"ping":    strings.Split(result[0].CMPing, " ")[0],
						"downloa": strings.Split(result[0].CMDLSpeed, " ")[0],
						"upload":  strings.Split(result[0].CMUpSpeed, " ")[0],
					}, {"brand": "cucc",
						"ping":    strings.Split(result[0].CUPing, " ")[0],
						"downloa": strings.Split(result[0].CUDLSpeed, " ")[0],
						"upload":  strings.Split(result[0].CUUpSpeed, " ")[0],
					}}}
				if p.Conn.WriteJSON(uploadresult) == nil {
					newError("succesfully upload speedtest result").AtInfo().WriteToLog()
				} else {
					newError("failed to upload speedtest result").AtInfo().WriteToLog()
				}
			}
			time.Sleep(time.Hour * time.Duration(p.SpeedTestCheckRate))

		}
	}()
	go func() {
		defer close(done)
		for {
			_, message, err := p.Conn.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				return
			}
			var dat Serverdata
			if err := json.Unmarshal(message, &dat); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(dat)
			}

			if !p.Handshake {
				b, _ := base64.StdEncoding.DecodeString(dat.Value)
				result, _ := RsaDecrypt(b)
				dat.Value = string(result)
				dat.Key = ""
				p.Conn.WriteJSON(dat)
				p.Handshake = true
			}
			if p.Handshake {
				p.updatenode(dat)
			}

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return newError("done")
		case t := <-ticker.C:
			err := p.Conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				return err
			}
		case <-interrupt:

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := p.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				fmt.Println("write close:", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return newError("interrupt")
		}
	}
}
func (p *GoPanel) updatenode(data Serverdata) {
	defer p.Unlock()
	p.Lock()
	switch data.Element {
	case "v2ray":
		{
			switch data.Operation {
			case "config":
				{
					if p.Server != nil {
						p.RemoveInbound(p.Server.Inbound.Tag)
						switch p.Server.Inbound.Protocol {
						case "ss":
							{
								for _, user := range p.Users {
									p.RemoveInbound(user.Email)
								}
							}
						}
						p.Server = &data
						p.HandlerServiceClient.InboundTag = p.Server.Inbound.Tag
						switch p.Server.Inbound.Protocol {
						case "ss":
							{
								p.AddDokodemoInbound()
								var streamsetting *internet.StreamConfig
								for key, user := range p.Users {
									newErrorf("ADD WS+SS %s  %s", key, user.Muhost).AtInfo().WriteToLog()
									streamsetting = client.GetDomainsocketStreamConfig(fmt.Sprintf("/etc/v2ray/%s.sock", user.Muhost))
									p.RuleServiceClient.AddUserAttrMachter("out_"+user.Muhost, fmt.Sprintf("attrs['host'] == '%s'", user.Muhost))
									p.HandlerServiceClient.AddFreedomOutbound("out_"+user.Muhost, streamsetting)
									p.HandlerServiceClient.AddDokodemoUser(user)
									p.AddSSuser(user)
								}
							}
						case "vmess":
							{
								p.AddVmessInbound()
								for _, user := range p.Users {
									p.AddVmessuser(user)
								}
							}

						}
					} else {
						p.Server = &data
						p.HandlerServiceClient.InboundTag = p.Server.Inbound.Tag
						switch p.Server.Inbound.Protocol {
						case "ss":
							{
							}
						case "vmess":
							{
								p.AddVmessInbound()
							}

						}
					}

				}
			case "userAdd":
				{
					switch p.Server.Inbound.Protocol {
					case "ss":
						{
							for _, user := range data.Users {
								usermodel := user.Build()
								p.Users[usermodel.Email] = usermodel
								p.AddSSuser(usermodel)
							}
						}
					case "vmess":
						{
							for _, user := range data.Users {
								usermodel := user.Build()
								p.Users[usermodel.Email] = usermodel
								p.AddVmessuser(usermodel)
							}
						}

					}
				}
			case "userUpdate":
				{
					switch p.Server.Inbound.Protocol {
					case "ss":
						{
							for _, email := range data.OldUsers {
								p.RemoveInbound(email.Email)
								delete(p.Users, email.Email)
								value := p.Users[email.Email]
								newErrorf("Remove AttrMachter %s", value.Muhost).AtInfo().WriteToLog()
								p.HandlerServiceClient.RemoveOutbound("out_" + value.Muhost)
								p.RuleServiceClient.RemveUserAttrMachter("out_" + value.Muhost)
								p.HandlerServiceClient.DelUser(value.Muhost)
							}
							for _, user := range data.NewUsers {
								usermodel := user.Build()
								p.Users[usermodel.Email] = usermodel
								p.AddSSuser(usermodel)
							}
						}
					case "vmess":
						{
							for _, email := range data.OldUsers {
								p.Removeuser(email.Email)
								delete(p.Users, email.Email)
							}
							for _, user := range data.NewUsers {
								usermodel := user.Build()
								p.Users[usermodel.Email] = usermodel
								p.AddVmessuser(usermodel)
							}
						}

					}
				}
			case "userRemove":
				{
					switch p.Server.Inbound.Protocol {
					case "ss":
						{
							for _, user := range data.Users {
								p.RemoveInbound(user.Email)
								delete(p.Users, user.Email)
							}
						}
					case "vmess":
						{
							for _, user := range data.Users {
								p.Removeuser(user.Email)
								delete(p.Users, user.Email)
							}
						}

					}
				}

			}

		}

	case "shadowsocks":
		{
			switch data.Operation {
			case "userAdd":
				{
					for _, user := range data.Users {
						usermodel := user.Build()
						p.Users[usermodel.Email] = usermodel
						p.AddSSuser(usermodel)
					}

				}
			case "userUpdate":
				{
					for _, email := range data.OldUsers {
						p.RemoveInbound(email.Email)
						delete(p.Users, email.Email)
					}
					for _, user := range data.NewUsers {
						usermodel := user.Build()
						p.Users[usermodel.Email] = usermodel
						p.AddSSuser(usermodel)
					}
				}
			case "userRemove":
				{
					for _, user := range data.Users {
						p.RemoveInbound(user.Email)
						delete(p.Users, user.Email)
					}
				}
			}

		}
	case "key":
		{
			p.Key = data.Value
			cmd := exec.Command("sh", "-c", fmt.Sprintf(`sed -i "s/\"bbbbaaaaaaa\"/\"%s\"/g" "/etc/v2ray/config.json"`, data.Value))
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			cmd.Run()
			newError(out.String()).AtInfo().WriteToLog()
			newError(stderr.String()).AtInfo().WriteToLog()

		}

	}

}
func (p *GoPanel) Job() error {
	defer p.Conn.Close()
	var err error
	u := url.URL{Scheme: "ws", Host: "167.160.180.162:8080", Path: "/node"}
	newErrorf("connecting to %s\n", u.String()).AtInfo().WriteToLog()
	p.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	p.Handshake = false
	if err != nil {
		return err
	}
	err = p.do()
	return err
}
func (p *GoPanel) Start() {
	var err error
	for {
		for err == nil {
			err = p.Job()
		}
		err = nil
	}
}
func (p *GoPanel) AddSSuser(user model.UserModel) error {
	var streamsetting *internet.StreamConfig
	var tm *serial.TypedMessage
	var address string
	streamsetting = &internet.StreamConfig{}
	address = "0.0.0.0"
	if p.Server != nil {
		if p.Server.Inbound.Protocol == "ss" {

			if p.Server.Inbound.StreamSettings.Network == "ws" {
				host := "www.bing.com"
				path := "/"
				address = "127.0.0.1"
				if p.Server.Inbound.StreamSettings.Wssettings.Path != "" {
					path = p.Server.Inbound.StreamSettings.Wssettings.Path
				}
				if p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"] != "" {
					host = p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"].(string)
				}
				if p.Server.Inbound.StreamSettings.TlsSettings.ServerName != "" {
					tm, _ = p.AddCert(p.Server.Inbound.StreamSettings.TlsSettings.ServerName)
				}
				streamsetting = client.GetWebSocketStreamConfig(path, host, tm)
			}
		}
	}
	return p.HandlerServiceClient.AddSSInbound(user, address, streamsetting)
}
func (p *GoPanel) Removeuser(email string) error {
	return p.HandlerServiceClient.DelUser(email)
}
func (p *GoPanel) AddVmessuser(user model.UserModel) error {
	return p.HandlerServiceClient.AddVmessUser(user)
}
func (p *GoPanel) AddVmessInbound() error {
	var streamsetting *internet.StreamConfig
	var tm *serial.TypedMessage
	//var err error
	streamsetting = &internet.StreamConfig{}

	if p.Server.Inbound.StreamSettings.Network == "ws" {
		host := "www.bing.com"
		path := "/"
		if p.Server.Inbound.StreamSettings.Wssettings.Path != "" {
			path = p.Server.Inbound.StreamSettings.Wssettings.Path
		}
		if p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"] != "" {
			host = p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"].(string)
		}
		if p.Server.Inbound.StreamSettings.TlsSettings.ServerName != "" {
			tm, _ = p.AddCert(p.Server.Inbound.StreamSettings.TlsSettings.ServerName)
		}

		streamsetting = client.GetWebSocketStreamConfig(path, host, tm)
	} else if p.Server.Inbound.StreamSettings.Network == "kcp" || p.Server.Inbound.StreamSettings.Network == "mkcp" {
		header_key := "noop"
		if p.Server.Inbound.StreamSettings.Kcpsettings.Header["type"] != "" {
			header_key = p.Server.Inbound.StreamSettings.Kcpsettings.Header["type"].(string)
		}
		streamsetting = client.GetKcpStreamConfig(header_key)
	}
	if err := p.HandlerServiceClient.AddVmessInbound(p.Server.Inbound.Port, "0.0.0.0", streamsetting); err != nil {
		return err
	} else {
		newErrorf("Successfully add MAIN INBOUND %s port %d", "0.0.0.0", p.Server.Inbound.Port).AtInfo().WriteToLog()
	}
	return nil
}
func (p *GoPanel) AddDokodemoInbound() error {
	var streamsetting *internet.StreamConfig
	var tm *serial.TypedMessage
	if p.Server.Inbound.StreamSettings.Network == "ws" {
		host := "www.bing.com"
		path := "/"
		if p.Server.Inbound.StreamSettings.Wssettings.Path != "" {
			path = p.Server.Inbound.StreamSettings.Wssettings.Path
		}
		if p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"] != "" {
			host = p.Server.Inbound.StreamSettings.Wssettings.Headers["Host"].(string)
		}
		if p.Server.Inbound.StreamSettings.TlsSettings.ServerName != "" {
			tm, _ = p.AddCert(p.Server.Inbound.StreamSettings.TlsSettings.ServerName)
		}

		streamsetting = client.GetWebSocketStreamConfig(path, host, tm)
	}
	if err := p.HandlerServiceClient.AddDokodemoInbound(p.Server.Inbound.Port, "0.0.0.0", streamsetting); err != nil {
		return err
	} else {
		newErrorf("Successfully add MAIN DokodemoInbound %s port %d", "0.0.0.0", p.Server.Inbound.Port).AtInfo().WriteToLog()
	}
}
func (p *GoPanel) RemoveInbound(tag string) error {
	return p.HandlerServiceClient.RemoveInbound(tag)
}
func (m *GoPanel) AddCert(server string) (*serial.TypedMessage, error) {
	var tlsconfig *conf.TLSConfig
	newError("Starting Issuing Tls Cert, please make sure 80 is free").AtInfo().WriteToLog()
	//cmd := exec.Command(fmt.Sprintf("command: %s %s %s %s", fmt.Sprintf("%s/.acme.sh/acme.sh", homeDir()), "--issue", fmt.Sprintf("-d %s", server), "--standalone"))
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s/.acme.sh/acme.sh --issue -d %s --standalone", homeDir(), server))
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Run()
	newError(out.String()).AtInfo().WriteToLog()
	newError(stderr.String()).AtInfo().WriteToLog()
	tlsconfig = &conf.TLSConfig{
		Certs: []*conf.TLSCertConfig{&conf.TLSCertConfig{
			CertFile: fmt.Sprintf("%s/.acme.sh/%s/fullchain.cer", homeDir(), server),
			KeyFile:  fmt.Sprintf("%[1]s/.acme.sh/%[2]s/%[2]s.key", homeDir(), server),
		}},
		InsecureCiphers: true,
	}
	cert, err := tlsconfig.Build()
	if err != nil {
		return nil, newError("Failed to build TLS config.").Base(err)
	}
	tm := serial.ToTypedMessage(cert)
	return tm, nil

}
func (m *GoPanel) StopCert(server string) error {
	newErrorf("Starting to remove %s from renew list", server).AtInfo().WriteToLog()
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s/.acme.sh/acme.sh --remove -d %s", homeDir(), server))
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Run()
	newError(out.String()).AtInfo().WriteToLog()
	newError(stderr.String()).AtInfo().WriteToLog()
	//newErrorf("Starting Remove  %s certs", server).AtInfo().WriteToLog()
	//cmd = exec.Command("rm", "-rf", fmt.Sprintf("%s/.acme.sh/%s/fullchain.cer", homeDir(), server), fmt.Sprintf("%[1]s/.acme.sh/%[2]s/%[2]s.key", homeDir(), server))
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr
	//cmd.Run()
	//newError(out.String()).AtInfo().WriteToLog()
	//newError(stderr.String()).AtInfo().WriteToLog()
	return nil
}
func (p *GoPanel) updateThroughout() {
	current_user := p.Users
	usertraffic := []UserTraffic{}
	userIPS := []model.UserOnLineIP{}
	for _, value := range current_user {
		current_upload, err := p.StatsServiceClient.GetUserUplink(value.Email)
		current_download, err := p.StatsServiceClient.GetUserDownlink(value.Email)
		if err != nil {
			newError(err).AtDebug().WriteToLog()
		}
		if current_upload+current_download > 0 {

			newErrorf("USER %s has use %d", value.Email, current_upload+current_download).AtDebug().WriteToLog()
			usertraffic = append(usertraffic, UserTraffic{Email: value.Email,
				Download: string(current_download),
				Upload:   string(current_upload)})
			current_user_ips, err := p.StatsServiceClient.GetUserIPs(value.Email)
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
	data := map[string]interface{}{"element": "traffic", "users": usertraffic}
	err := p.Conn.WriteJSON(data)
	if err == nil {
		newErrorf("Successfully upload %d users traffic", len(usertraffic)).AtInfo().WriteToLog()
	}

}
func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	//获取私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, newError("private key error!")
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 解密
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwOJK1RJBUwRu/5aCyktTaietXFMOAAkElhSq1M6BocVWs7yD
y592CX30Bl0Ul4faWM9EZSlhak8Ay1CdMNis+lBZanKmAO2bPmSIIYBDQE2BzLIo
MoJWi/Cd5PevioKSRPytqVB/S4+xz1IOD8Y407SZM3LfZ5XMfqC+VHpcnAycQ8iT
FK0s3yjImathFNF3U7fiEzU4G7PJRn8e9ntubDd1pXYABqrVF/REcd/3Rs/qrlhG
v3b7tAXZb2lkSLdCq3Md+BMksxUCoH3rZijSphbZSCdIrzofg+IG0y5WtdsBz6uw
Ol2QX/EUoEdO+xhLgaOFykUoWz037ZzkLEhKkQIDAQABAoIBAB+1lAPPSnnxYqYW
Ak5rb70l5LQm20haMyzRHPx7Loh/vq8xsKELCAardDCPoNEAfn7XJDFVSjSF5GWI
TS84j8de6jQ7wNqqNTleoZqQUX4Cv/H83+rdzoiW9/4qUet9Z7p7p7kMCMFNUDf7
D2C8f58eM4lnux52W/X9SwzsSMlGaGHcAKPz4vXUFWyt3naVtANhdkHjgKxA0Ev4
W7yRgpbOKruPKzBNTRXAqq+yHZj/pONtXl8do+plwhHU8CW0BPyvkU4DH52lxWza
mM71ow8UJC30FXF/NZ+wthFnRZO3/dhaeuNYgX7yAs3DhNn7Q8nzU4ujd8ug2OGf
iJ9C8YECgYEA32KthV7VTQRq3VuXMoVrYjjGf4+z6BVNpTsJAa4kF+vtTXTLgb4i
jpUrq6zPWZkQ/nR7+CuEQRUKbky4SSHTnrQ4yIWZTCPDAveXbLwzvNA1xD4w4nOc
JgG/WYiDtAf05TwC8p/BslX20Ox8ZAXUq6pkAeb1t8M2s7uDpZNtBMkCgYEA3QuU
vrrqYuD9bQGl20qJI6svH875uDLYFcUEu/vA/7gDycVRChrmVe4wU5HFErLNFkHi
uifiHo75mgBzwYKyiLgO5ik8E5BJCgEyA9SfEgRHjozIpnHfGbTtpfh4MQf2hFsy
DJbeeRFzQs4EW2gS964FK53zsEtnr7bphtvfY4kCgYEAgf6wr95iDnG1pp94O2Q8
+2nCydTcgwBysObL9Phb9LfM3rhK/XOiNItGYJ8uAxv6MbmjsuXQDvepnEp1K8nN
lpuWN8rXTOG6yG1A53wWN5iK0WrHk+BnTA7URcwVqJzAvO3RYVPqqlcwTKByOtrR
yhxcGmdHMusdWDaVA7PpS1ECgYATCGs/XQLPjsXje+/XCPzz+Epvd7fi12XpwfQd
Z5j/q82PsxC+SQCqR38bwwJwELs9/mBSXRrIPNFbJEzTTbinswl9YfGNUbAoT2AK
GmWz/HBY4uBoDIgEQ6Lu1o0q05+zV9LgaKExVYJSL0EKydRQRUimr8wK0wNTivFi
rk322QKBgHD3aEN39rlUesTPX8OAbPD77PcKxoATwpPVrlH8YV7TxRQjs5yxLrxL
S21UkPRxuDS5CMXZ+7gA3HqEQTXanNKJuQlsCIWsvipLn03DK40nYj54OjEKYo/F
UgBgrck6Zhxbps5leuf9dhiBrFUPjC/gcfyHd/PYxoypHuQ3JUsJ
-----END RSA PRIVATE KEY-----
`)

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		os.Exit(1)
	}
	return usr.HomeDir
}
