package speedtest

// the go speedtest-cli code is from https://github.com/surol/speedtest-cli
import (
	"fmt"
	"os"
	"strings"
	"time"
	"v2ray.com/core/common/errors"
)

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

type Speedresult struct {
	CTPing    string `json:"telecomping"`
	CTUpSpeed string `json:"telecomeupload"`
	CTDLSpeed string `json:"telecomedownload"`
	CUPing    string `json:"unicomping"`
	CUUpSpeed string `json:"unicomupload"`
	CUDLSpeed string `json:"unicomdownload"`
	CMPing    string `json:"cmccping"`
	CMUpSpeed string `json:"cmccupload"`
	CMDLSpeed string `json:"cmccdownload"`
}

func GetSpeedtest() ([]Speedresult, error) {
	opts := ParseOpts()
	client := NewClient(opts)

	config, err := client.Config()
	if err != nil {
		return nil, newError(err)
	}

	newErrorf("Testing from %s (%s)...\n", config.Client.ISP, config.Client.IP).AtInfo().WriteToLog()
	final_result := []Speedresult{}
	result := Speedresult{}
	server := selectServer("Telecom", client)
	if server != nil {
		result.CTPing = fmt.Sprintf("%f ms", server.Latency.Seconds()*1e3)
		result.CTDLSpeed = fmt.Sprintf("%.2f MiB/s", server.DownloadSpeed())
		result.CTUpSpeed = fmt.Sprintf("%.2f MiB/s", server.UploadSpeed())
	}
	server = selectServer("Mobile", client)
	if server != nil {
		result.CMPing = fmt.Sprintf("%f ms", server.Latency.Seconds()*1e3)
		result.CMDLSpeed = fmt.Sprintf("%.2f MiB/s", server.DownloadSpeed())
		result.CMUpSpeed = fmt.Sprintf("%.2f MiB/s", server.UploadSpeed())
	}

	server = selectServer("Unicom", client)
	if server != nil {
		result.CUPing = fmt.Sprintf("%f ms", server.Latency.Seconds()*1e3)
		result.CUDLSpeed = fmt.Sprintf("%.2f MiB/s", server.DownloadSpeed())
		result.CUUpSpeed = fmt.Sprintf("%.2f MiB/s", server.UploadSpeed())
	}
	return append(final_result, result), nil
}

func selectServer(sponsor string, client Client) (selected *Server) {
	servers, err := client.AllServers()
	if err != nil {
		newError("Failed to load server list: %v", err).AtWarning().WriteToLog()
		return nil
	}
	sponsor_servers := new(Servers)
	for _, server := range servers.List {
		if (server.Country == "China" || server.Country == "CN") && strings.Contains(server.Sponsor, sponsor) {
			sponsor_servers.List = append(sponsor_servers.List, server)
		}
	}
	if len(sponsor_servers.List) > 0 {

		selected = sponsor_servers.MeasureLatencies(
			DefaultLatencyMeasureTimes,
			DefaultErrorLatency).First()
		fmt.Println(selected.Sponsor)
		return selected
	}
	return nil
}
