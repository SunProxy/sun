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
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sunproxy/sun/sun/command"
	"github.com/sunproxy/sun/sun/event"
	"github.com/sunproxy/sun/sun/logger"
	"github.com/sunproxy/sun/sun/ray"
	"go.uber.org/atomic"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
)

var emptychunk = make([]byte, 257)

type Sun struct {
	Listener *minecraft.Listener
	Rays     map[string]*ray.Ray
	Hub      string
	//Planets    map[uuid.UUID]*planet.Planet
	//PListener  net.Listener
	Status        StatusProvider
	Key           string
	LoadBalancer  LoadBalancer
	StatusCommand bool
	Logger        logger.Logger
	handler       Handler
	handlerMu     sync.RWMutex
	CmdProcessor  command.Processor
}

type StatusProvider struct {
	ogs     minecraft.ServerStatus
	playerc *atomic.Int64
}

func (s StatusProvider) ServerStatus(_ int, _ int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  s.ogs.ServerName,
		PlayerCount: int(s.playerc.Load()),
		MaxPlayers:  s.ogs.MaxPlayers,
	}
}

func NewSunW(config Config) (*Sun, error) {
	var status minecraft.ServerStatusProvider
	status = StatusProvider{config.Status, atomic.NewInt64(0)}
	sun := &Sun{
		Rays: make(map[string]*ray.Ray,
			config.Status.MaxPlayers),
		Hub:           config.Hub,
		StatusCommand: config.Proxy.StatusCommand,
		Logger:        logger.New(config.Proxy.Logger.File, config.Proxy.Logger.Debug),
		handler:       NopHandler{},
	}
	sun.CmdProcessor = command.NewProcessor(sun.Logger, func(cmd command.Command) {})
	if config.Proxy.MOTDForward {
		tmpStatus, err := sun.MotdForward()
		if err != nil {
			_ = sun.Logger.Warn("Unable to MOTDForward to any LoadBalancer or the hub, Falling back to the normal status.")
		} else {
			status = tmpStatus
		}
	}
	listener, err := minecraft.ListenConfig{
		AuthenticationDisabled: !config.Proxy.XboxAuthentication,
		StatusProvider:         status,
		ResourcePacks:          LoadResourcePacks("./resource_packs/"),
	}.Listen("raknet", fmt.Sprint(":", config.Proxy.Port))
	sun.Status = StatusProvider{config.Status, atomic.NewInt64(0)}
	if err != nil {
		return sun, err
	}
	sun.Listener = listener
	lb := LoadBalancer{Enabled: false}
	if config.Proxy.LoadBalancer.Enabled {
		lb.Servers = config.Proxy.LoadBalancer.Balancers
		lb.Overflow = NewOverflowBalancer(lb.Servers)
		lb.Enabled = true
	}
	sun.LoadBalancer = lb
	if config.Proxy.Pprof.Enabled {
		go func() {
			addr := fmt.Sprint("127.0.0.1:", config.Proxy.Pprof.Port)
			_ = sun.Logger.Debugf("Ppof webserver starting on %s!", addr)
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				_ = sun.Logger.Warnf("Failed to start Ppof webserver on %s!", addr)
				return
			}
		}()
	}
	return sun, nil
}

func (s *Sun) MotdForward() (*minecraft.ForeignStatusProvider, error) {
	status, err := minecraft.NewForeignStatusProvider(s.Hub)
	if err != nil {
		if s.LoadBalancer.Enabled {
			for i := 0; i < len(s.LoadBalancer.Servers); i++ {
				status, err := minecraft.NewForeignStatusProvider(s.LoadBalancer.Balance(nil))
				if err == nil {
					_ = s.Logger.Warnf("Hub Server and LoadBalancers %+v are all down rays are "+
						"now being connected to LoadBalancer %v", s.LoadBalancer.Servers[:i], i)
					return status, err
				}
			}
			return nil, err
		}
	}
	return status, nil
}

// NewSun
func NewSun() (*Sun, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return NewSunW(cfg)
}

func (s *Sun) main() {
	s.CmdProcessor.RegisterDefaults()
	s.CmdProcessor.StartProcessing(os.Stdin)
	defer s.Listener.Close()
	for {
		//Listener won't be closed unless it is manually done
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		r := ray.New(conn.(*minecraft.Conn))
		rconn, err := s.ConnectToHub(r)
		if err != nil {
			_ = s.Logger.Errorf("No Active LoadBalancers or Hub Could accept Ray: %s!",
				r.Conn().IdentityData().DisplayName)
			_ = r.Conn().WritePacket(&packet.Disconnect{
				Message:                 text.Colourf("<red>You Have been Disconnected!</red>"),
				HideDisconnectionScreen: false})
			_ = r.Conn().Close()
			continue
		}

		r.SetRemote(rconn)
		s.MakeRay(r)
	}
}

// Start starts the main loop for the given listener.
func (s *Sun) Start() {
	s.main()
}

// ConnectToHub redirects a ray to the hub server.
func (s *Sun) ConnectToHub(ray *ray.Ray) (*minecraft.Conn, error) {
	rconn, err := s.Dial(ray, s.Hub)
	if err != nil {
		if s.LoadBalancer.Enabled {
			for i := 0; i < len(s.LoadBalancer.Servers); i++ {
				conn, err := s.Dial(ray, s.LoadBalancer.Balance(ray))
				if err == nil {
					_ = s.Logger.Warnf("Hub Server and LoadBalancers %+v are all down rays are "+
						"now being connected to LoadBalancer %v", s.LoadBalancer.Servers[:i], i)
					return conn, err
				}
			}
			return nil, err
		}
	}
	return rconn, err
}

// Dial dials a connection to a server.
func (s *Sun) Dial(ray *ray.Ray, addr string) (*minecraft.Conn, error) {
	return minecraft.Dialer{
		ClientData:   ray.Conn().ClientData(),
		IdentityData: ray.Conn().IdentityData()}.Dial("raknet", addr)
}

// MakeRay creates a new ray and makes its listener and receiver.
func (s *Sun) MakeRay(ray *ray.Ray) {
	//start the player up
	var g sync.WaitGroup
	g.Add(2)
	var Gerr error
	go func() {
		if err := ray.Conn().StartGame(ray.Remote().GameData()); err != nil {
			_ = s.Logger.Errorf("Start Game Timeout on ray: %s", ray.Conn().IdentityData().DisplayName)
			Gerr = err
		}
		g.Done()
	}()
	go func() {
		if err := ray.Remote().DoSpawn(); err != nil {
			_ = s.Logger.Errorf("Spawn Timeout on remote: %s", ray.Remote())
			Gerr = err
		}
		g.Done()
	}()
	if Gerr != nil {
		return
	}
	g.Wait()
	//Run through the join handler
	ctx := event.C()
	ctx.Continue(func() {
		//start translator
		ray.InitTranslators(ray.Conn().GameData())
		//Add to player count
		s.Status.playerc.Inc()
		//add to player list
		s.Rays[ray.Conn().IdentityData().Identity] = ray
		//Start the two listener functions
		s.handleRay(ray)
	})
	s.Handler().HandleRayJoin(ctx, ray)
}

// BreakRay disconnects a ray from the proxy in its entirety.
func (s *Sun) BreakRay(ray *ray.Ray) {
	_ = s.Listener.Disconnect(ray.Conn(), text.Colourf("<red>You Have been Disconnected!</red>"))
	_ = ray.Remote().Close()
	s.Status.playerc.Dec()
	delete(s.Rays, ray.Conn().IdentityData().Identity)
}

// SendMessageToServers sends a message to a list of connected servers
func (s *Sun) SendMessageToServers(Message string, Servers []string) {
	for _, server := range Servers {
		for _, r := range s.Rays {
			if r.Remote().RemoteAddr().String() == server {
				err := r.Conn().WritePacket(&packet.Text{Message: Message, TextType: packet.TextTypeRaw})
				if err != nil {
					s.BreakRay(r)
				}
			}
		}
	}
}

/*
SendMessage is used for sending a Sun wide message to all the connected clients
*/
func (s *Sun) SendMessage(Message string) {
	for _, r := range s.Rays {
		//Send raw chat to each player as client will accept it
		err := r.Conn().WritePacket(&packet.Text{Message: Message, TextType: packet.TextTypeRaw})
		if err != nil {
			s.BreakRay(r)
		}
	}
}

func (s *Sun) Handler() Handler {
	if s == nil {
		return NopHandler{}
	}
	s.handlerMu.RLock()
	handler := s.handler
	s.handlerMu.RUnlock()
	return handler
}

func (s *Sun) Handle(handler Handler) {
	s.handlerMu.Lock()
	defer s.handlerMu.Unlock()
	if handler == nil {
		handler = NopHandler{}
	}
	s.handler = handler
}
