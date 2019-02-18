package speedtest

import (
	"time"
)

type Opts struct {
	SpeedInBytes bool
	Quiet        bool
	List         bool
	Server       ServerID
	Interface    string
	Timeout      time.Duration
	Secure       bool
	Help         bool
	Version      bool
}

func NewOpts() *Opts {

	return &Opts{
		SpeedInBytes: false,
		Quiet:        true,
		List:         false,
		Server:       0,
		Interface:    "",
		Timeout:      10 * time.Second,
		Secure:       true,
		Help:         false,
		Version:      false,
	}
}
