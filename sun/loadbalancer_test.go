package sun

import (
	"fmt"
	"testing"
)

func TestOverflowBalancer_Balance(t *testing.T) {
	b := NewOverflowBalancer([]IpAddr{{Address: "127.0.0.1", Port: 19133}, {Address: "127.0.0.1", Port: 19134}})
	for i := 0; i < 4; i++ {
		fmt.Println(b.Balance(&Ray{}))
	}
}
