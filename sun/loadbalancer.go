/**
      ___           ___           ___
     /  /\         /__/\         /__/\
    /  /:/_        \  \:\        \  \:\
   /  /:/ /\        \  \:\        \  \:\
  /  /:/ /::\   ___  \  \:\   _____\__\:\
 /__/:/ /:/\:\ /__/\  \__\:\ /__/::::::::\
 \  \:\/:/~/:/ \  \:\ /  /:/ \  \:\~~\~~\/
  \  \::/ /:/   \  \:\  /:/   \  \:\  ~~~
   \__\/ /:/     \  \:\/:/     \  \:\
     /__/:/       \  \::/       \  \:\
     \__\/         \__\/         \__\/

MIT License

Copyright (c) 2020 Jviguy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package sun

import (
	"fmt"
	roundrobin "github.com/hlts2/round-robin"
	"github.com/sunproxy/sun/sun/ray"
	"log"
	"net/url"
)

type Balancer interface {
	Balance(ray *ray.Ray) string
}

type LoadBalancer struct {
	Servers  []string
	Overflow *OverflowBalancer
	Enabled  bool
}

func (l LoadBalancer) Balance(ray *ray.Ray) string {
	return l.Overflow.Balance(ray)
}

type OverflowBalancer struct {
	rr roundrobin.RoundRobin
}

func (r *OverflowBalancer) Balance(ray *ray.Ray) string {
	ul := r.rr.Next()
	return fmt.Sprintf(ul.Host)
}

func NewOverflowBalancer(servers []string) *OverflowBalancer {
	urls := make([]*url.URL, 0)
	for _, server := range servers {
		urls = append(urls, &url.URL{Host: server})
	}
	rr, err := roundrobin.New(urls)
	if err != nil {
		log.Fatal(err)
	}
	return &OverflowBalancer{rr: rr}
}
