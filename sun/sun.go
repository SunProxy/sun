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
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/sunproxy/sun/sun/event"
	"github.com/sunproxy/sun/sun/ip_addr"
	"github.com/sunproxy/sun/sun/logger"
	sunpacket "github.com/sunproxy/sun/sun/packet"
	"github.com/sunproxy/sun/sun/planet"
	"github.com/sunproxy/sun/sun/plugin"
	"github.com/sunproxy/sun/sun/ray"
	"github.com/sunproxy/sun/sun/remote"
	"go.uber.org/atomic"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

var emptychunk = make([]byte, 257)

type Sun struct {
	Listener        *minecraft.Listener
	Rays            map[string]*ray.Ray
	Hub             ip_addr.IpAddr
	Planets         map[uuid.UUID]*planet.Planet
	PListener       net.Listener
	Status          StatusProvider
	Key             string
	PWarnings       map[string]int
	PCooldowns      map[string]time.Time
	Servers         map[string]ip_addr.IpAddr
	TransferCommand bool
	LoadBalancer    LoadBalancer
	StatusCommand   bool
	Logger          logger.Logger
	PluginManager   *plugin.Manager
	handler         Handler
	handlerMu       sync.RWMutex
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
		ShowVersion: s.ogs.ShowVersion,
	}
}

/*
Returns a new sun with config the specified config hence W
*/
func NewSunW(config Config) (*Sun, error) {
	var status minecraft.ServerStatusProvider
	status = StatusProvider{config.Status, atomic.NewInt64(0)}
	sun := &Sun{
		Rays: make(map[string]*ray.Ray,
			config.Status.MaxPlayers),
		Hub: config.Hub, Planets: make(map[uuid.UUID]*planet.Planet),
		TransferCommand: config.Proxy.TransferCommand.Enabled,
		Servers:         config.Proxy.TransferCommand.Servers,
		StatusCommand:   config.Proxy.StatusCommand,
		Logger:          logger.New(config.Proxy.Logger.File, config.Proxy.Logger.Debug),
		handler:         NopHandler{},
	}
	sun.PluginManager = plugin.NewManager(sun.Logger)
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
	if config.Proxy.Ppof.Enabled {
		go func() {
			addr := fmt.Sprint("127.0.0.1:", config.Proxy.Ppof.Port)
			_ = sun.Logger.Debugf("Ppof webserver starting on %s!", addr)
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				_ = sun.Logger.Warnf("Failed to start Ppof webserver on %s!", addr)
				return
			}
		}()
	}
	if config.Tcp.Enabled {
		plistener, err := net.Listen("tcp", ":42069")
		if err != nil {
			return nil, err
		}
		sun.PListener = plistener
		sun.PCooldowns = make(map[string]time.Time)
		sun.PWarnings = make(map[string]int)
	}
	registerPackets()
	return sun, nil
}

func registerPackets() {
	packet.Register(sunpacket.IDRayTransfer, func() packet.Packet { return &sunpacket.Transfer{} })
	packet.Register(sunpacket.IDRayText, func() packet.Packet { return &sunpacket.Text{} })
}

func (s *Sun) MotdForward() (*minecraft.ForeignStatusProvider, error) {
	status, err := minecraft.NewForeignStatusProvider(s.Hub.ToString())
	if err != nil {
		if s.LoadBalancer.Enabled {
			for i := 0; i < len(s.LoadBalancer.Servers); i++ {
				status, err := minecraft.NewForeignStatusProvider(s.LoadBalancer.Balance(nil).ToString())
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

/*
Returns a new sun with a auto detected config
*/
func NewSun() (*Sun, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return NewSunW(cfg)
}

func (s *Sun) main() {
	_ = s.PluginManager.VM.Set("sun", s)
	s.PluginManager.LoadPluginDir()
	defer s.Listener.Close()
	if s.PListener != nil {
		go func() {
			for {
				conn, err := s.PListener.Accept()
				if err != nil {
					log.Println(err)
					continue
				}
				pl := planet.NewPlanet(conn)
				if tl, ok := s.PCooldowns[pl.Conn().RemoteAddr().String()]; ok {
					if time.Now().Before(tl) {
						_ = pl.WritePacket(&sunpacket.PlanetDisconnect{Message: fmt.Sprintf("You are on cooldown for %v seconds!",
							time.Now().Sub(s.PCooldowns[pl.Conn().RemoteAddr().String()]).
								Seconds())})
						_ = pl.Conn().Close()
						continue
					}
					delete(s.PCooldowns, conn.RemoteAddr().String())
				}
				pk, err := pl.ReadPacket()
				if pk, ok := pk.(*sunpacket.PlanetAuth); ok {
					if pk.Key == s.Key {
						s.AddPlanet(pl)
						continue
					}
				}
				if _, ok := s.PWarnings[pl.Conn().RemoteAddr().String()]; !ok {
					s.PWarnings[pl.Conn().RemoteAddr().String()] = 3
				}
				s.PWarnings[pl.Conn().RemoteAddr().String()]--
				if s.PWarnings[pl.Conn().RemoteAddr().String()] <= 0 {
					s.PWarnings[pl.Conn().RemoteAddr().String()] = 3
					s.PCooldowns[pl.Conn().RemoteAddr().String()] = time.Now().Add(300 * time.Second)
					_ = pl.WritePacket(&sunpacket.PlanetDisconnect{Message: fmt.Sprintf("You are on cooldown for %v seconds!",
						time.Now().Sub(s.PCooldowns[pl.Conn().RemoteAddr().String()]).Seconds(),
					)})
					_ = pl.Conn().Close()
				}
				_ = pl.WritePacket(
					&sunpacket.PlanetDisconnect{
						Message: fmt.Sprintf("Invalid Authorization Key Provided %v "+
							"Tries Remain Until A 300 Second Cooldown!",
							s.PWarnings[pl.Conn().RemoteAddr().String()])})
				continue
			}
		}()
	}
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

		r.SetRemote(remote.New(rconn, s.Hub))
		s.MakeRay(r)
	}
}

/*
Starts the proxy.
*/
func (s *Sun) Start() {
	s.main()
}

/**
ConnectToHub will attempt to connect a ray to the hub server.
If the said hub server rejects the connection for any reason
the proxy will then go through the overflow Balancer to find the next usable ip until it runs out.
*/
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

func (s *Sun) Dial(ray *ray.Ray, addr ip_addr.IpAddr) (*minecraft.Conn, error) {
	return minecraft.Dialer{
		ClientData:   ray.Conn().ClientData(),
		IdentityData: ray.Conn().IdentityData()}.Dial("raknet", addr.ToString())
}

/*
Adds a player to the sun and readies them
*/
func (s *Sun) MakeRay(ray *ray.Ray) {
	//start the player up
	var g sync.WaitGroup
	g.Add(2)
	var Gerr error
	go func() {
		if err := ray.Conn().StartGame(ray.Remote().Conn.GameData()); err != nil {
			_ = s.Logger.Errorf("Start Game Timeout on ray: %s", ray.Conn().IdentityData().DisplayName)
			Gerr = err
		}
		g.Done()
	}()
	go func() {
		if err := ray.Remote().Conn.DoSpawn(); err != nil {
			_ = s.Logger.Errorf("Do Spawn Timeout on remote: %s", ray.Remote().Addr().ToString())
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

/*
Closes a players session cleanly with a nice disconnection message!
*/
func (s *Sun) BreakRay(ray *ray.Ray) {
	_ = s.Listener.Disconnect(ray.Conn(), text.Colourf("<red>You Have been Disconnected!</red>"))
	_ = ray.Remote().Conn.Close()
	s.Status.playerc.Dec()
	delete(s.Rays, ray.Conn().IdentityData().Identity)
}

func (s *Sun) SendMessageToServers(Message string, Servers []string) {
	for _, server := range Servers {
		for _, r := range s.Rays {
			if r.Remote().Addr().ToString() == server {
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

func (s *Sun) AddPlanet(planet *planet.Planet) {
	id := uuid.New()
	planet.Id = id
	s.Planets[id] = planet
	s.handlePlanet(planet)
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
