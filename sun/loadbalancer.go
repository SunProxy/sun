package sun

import (
	roundrobin "github.com/hlts2/round-robin"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type Balancer interface {
	Balance(ray *Ray) IpAddr
}

type LoadBalancer struct {
	Servers  []IpAddr
	Overflow *OverflowBalancer
	Enabled  bool
}

func (l LoadBalancer) Balance(ray *Ray) IpAddr {
	return l.Overflow.Balance(ray)
}

type OverflowBalancer struct {
	rr roundrobin.RoundRobin
}

func (r *OverflowBalancer) Balance(ray *Ray) IpAddr {
	ul := r.rr.Next()
	arr := strings.Split(ul.Host, ":")
	port, _ := strconv.Atoi(arr[1])
	return IpAddr{Address: arr[0], Port: uint16(port)}
}

func NewOverflowBalancer(servers []IpAddr) *OverflowBalancer {
	urls := make([]*url.URL, 0)
	for _, server := range servers {
		urls = append(urls, &url.URL{Host: server.ToString()})
	}
	rr, err := roundrobin.New(urls)
	if err != nil {
		log.Fatal(err)
	}
	return &OverflowBalancer{rr: rr}
}
