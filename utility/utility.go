package utility

import (
	"fmt"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

func InStr(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func GetSystemLoad() string {
	stat, err := load.Avg()
	if err != nil {
		return "0.00 0.00 0.00"
	}

	return fmt.Sprintf("%.2f %.2f %.2f", stat.Load1, stat.Load5, stat.Load15)
}
func GetSystemUptime() string {
	time, err := host.Uptime()
	if err != nil {
		return ""
	}
	return fmt.Sprint(time)

}
