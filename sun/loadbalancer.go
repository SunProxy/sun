package sun

import (
	roundrobin "github.com/hlts2/round-robin"
	"github.com/sunproxy/sun/sun/ip_addr"
	"github.com/sunproxy/sun/sun/ray"
	"log"
	"net/url"
	"strconv"
)

type Balancer interface {
	Balance(ray *ray.Ray) ip_addr.IpAddr
}

type LoadBalancer struct {
	Servers  []ip_addr.IpAddr
	Overflow *OverflowBalancer
	Enabled  bool
}

func (l LoadBalancer) Balance(ray *ray.Ray) ip_addr.IpAddr {
	return l.Overflow.Balance(ray)
}

type OverflowBalancer struct {
	rr roundrobin.RoundRobin
}

func (r *OverflowBalancer) Balance(ray *ray.Ray) ip_addr.IpAddr {
	ul := r.rr.Next()
	port, _ := strconv.Atoi(ul.Port())
	return ip_addr.IpAddr{Address: ul.Hostname(), Port: uint16(port)}
}

func NewOverflowBalancer(servers []ip_addr.IpAddr) *OverflowBalancer {
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
