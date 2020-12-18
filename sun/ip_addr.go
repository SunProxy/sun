package sun

import (
	"fmt"
)

type IpAddr struct {
	Ip   string
	Port uint16
}

func (ip *IpAddr) ToString() string {
	return fmt.Sprint(ip.Ip, ":", ip.Port)
}
