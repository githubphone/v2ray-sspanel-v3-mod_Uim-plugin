package v2ray_sspanel_v3_mod_Uim_plugin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/imroc/req"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/client"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/config"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/db"
	"google.golang.org/grpc/status"
	"os"
	"runtime"
	"strings"
	"time"
	"v2ray.com/core/common/errors"
)

func init1() {
	go func() {
		err := run()
		if err != nil {
			fatal(err)
		}
	}()
}
func CheckAuth(url string, params map[string]interface{}) (*db.AuthResponse, error) {
	var response = db.AuthResponse{}
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
func checkAuth(panelurl string) (bool, error) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(panelurl))
	cipherStr := md5Ctx.Sum(nil)
	current_md5 := hex.EncodeToString(cipherStr)
	re, err := CheckAuth("https://auth.rico93.com", map[string]interface{}{"md5": current_md5})
	if re != nil {
		if re.Token != "" {
			auth := config.AESDecodeStr(re.Token, config.Key)
			return current_md5 == auth, nil
		} else {
			return false, newErrorf("Auth failed, current url: %s  current md5 %s", panelurl, current_md5)
		}
	} else {
		return false, newErrorf("Can't get data from server or the  data is not as expected current url: %s  current md5 %s", panelurl, current_md5).Base(err)
	}

}
func run() error {

	err := config.CommandLine.Parse(os.Args[1:])

	cfg, err := config.GetConfig()
	if err != nil || *config.Test || cfg == nil {
		return err
	}

	// wait v2ray
	time.Sleep(3 * time.Second)

	go func() {
		var ok bool
		var err error
		if cfg.Usemysql == 0 {
			if strings.HasPrefix(cfg.PanelUrl, "https://") || strings.HasPrefix(cfg.PanelUrl, "http://") {
				ok, err = checkAuth(cfg.PanelUrl)
			} else {
				ok, err = checkAuth(cfg.PanelUrl)
				cfg.PanelUrl = "https://" + cfg.PanelUrl
			}

		} else {
			if cfg.MySQL != nil {
				ok, err = checkAuth(cfg.MySQL.Host)
			} else {
				fatal("Please Add Mysql setting")
			}
		}
		if ok && err == nil {
			apiInbound := config.GetInboundConfigByTag(cfg.V2rayConfig.Api.Tag, cfg.V2rayConfig.InboundConfigs)
			gRPCAddr := fmt.Sprintf("%s:%d", apiInbound.ListenOn.String(), apiInbound.PortRange.From)
			gRPCConn, err := client.ConnectGRPC(gRPCAddr, 10*time.Second)
			if err != nil {
				if s, ok := status.FromError(err); ok {
					err = errors.New(s.Message())
				}
				fatal(fmt.Sprintf("connect to gRPC server \"%s\" err: ", gRPCAddr), err)
			}
			newErrorf("Connected gRPC server \"%s\" ", gRPCAddr).AtWarning().WriteToLog()
			var database db.Db
			if cfg.Paneltype == 0 {
				newError("Using SSpanel").AtInfo().WriteToLog()
				if cfg.Usemysql == 1 {
					mysql, err := db.NewMySQLConn(cfg.MySQL)
					if err != nil {
						fmt.Println(err)
					}
					database = &db.SSpanel{Db: mysql, MU_SUFFIX: cfg.MU_SUFFIX, MU_REGEX: cfg.MU_REGEX}
					newError("Using Mysql Now").AtInfo().WriteToLog()
				} else {
					database = &db.Webapi{
						WebToken:   cfg.PanelKey,
						WebBaseURl: cfg.PanelUrl,
						MU_SUFFIX:  cfg.MU_SUFFIX,
						MU_REGEX:   cfg.MU_REGEX,
					}
					newError("Using Webapi Now").AtInfo().WriteToLog()
				}
			} else {
				newError("Using SSRpanel").AtInfo().WriteToLog()
				if cfg.Usemysql == 1 {
					mysql, err := db.NewMySQLConn(cfg.MySQL)
					if err != nil {
						fmt.Println(err)
					}
					database = &db.SSRpanel{Db: mysql, MU_SUFFIX: cfg.MU_SUFFIX, MU_REGEX: cfg.MU_REGEX}
					newError("Using Mysql Now").AtInfo().WriteToLog()
				} else {
					fatal("No databese config for ssrpanel")
				}
			}

			p, err := NewPanel(gRPCConn, database, cfg)
			if err != nil {
				fatal("new panel error", err)
			}
			p.Start()
		} else {
			if err != nil {
				fatal(err)
			}
		}

	}()

	// Explicitly triggering GC to remove garbage
	runtime.GC()

	return nil
}

func newErrorf(format string, a ...interface{}) *errors.Error {
	return newError(fmt.Sprintf(format, a...))
}

func newError(values ...interface{}) *errors.Error {
	values = append([]interface{}{"SSPanelPlugin: "}, values...)
	return errors.New(values...)
}

func fatal(values ...interface{}) {
	newError(values...).AtError().WriteToLog()
	// Wait log
	time.Sleep(1 * time.Second)
	os.Exit(-2)
}
