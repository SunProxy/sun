package plugin

import (
	"github.com/robertkrimen/otto"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sunproxy/sun/sun/event"
	"github.com/sunproxy/sun/sun/logger"
	"github.com/sunproxy/sun/sun/ray"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Manager struct {
	//The javascript vm
	VM *otto.Otto
	//For errors and shit atm
	Logger logger.Logger
	//map of the said plugins loaded
	Plugins map[string]Plugin
	//The handler that we provide
	Handler ray.Handler
}

func NewManager(log logger.Logger) *Manager {
	iso := otto.New()
	_ = iso.Set("events", false)
	_ = iso.Set("logger", log)
	return &Manager{Logger: log, Plugins: make(map[string]Plugin), VM: iso, Handler: nil}
}

func (m *Manager) LoadPluginDir() {
	plugins, err := ioutil.ReadDir("plugins")
	if err != nil {
		//Make plugin dir and return because if the dir doesn't exist there will obv be no plugins...
		_ = os.Mkdir("plugins", 0644)
		return
	}
	for _, plugindir := range plugins {
		if plugindir.IsDir() {
			yml, err := ioutil.ReadFile("plugins/" + plugindir.Name() + "/plugin.yml")
			if err != nil {
				_ = m.Logger.Errorf("Failed to load plugin %s, plugin.yml could not be read!",
					plugindir.Name())
				continue
			}
			plugin := Plugin{}
			err = yaml.Unmarshal(yml, &plugin)
			if err != nil {
				_ = m.Logger.Errorf("Syntactic error in plugin.yml of plugin %s, error: %s", plugindir.Name(),
					err.Error())
				continue
			}
			if err = m.LoadPlugin(plugin); err != nil {
				_ = m.Logger.Errorf("Failed to load plugin %s, error: %s",
					plugindir.Name(), err)
				continue
			}
		}
	}
}

func (m *Manager) LoadPlugin(plugin Plugin) error {
	data, err := ioutil.ReadFile("plugins/" + plugin.Name + "/" + plugin.Main)
	if err != nil {
		return err
	}
	_, err = m.VM.Run(string(data))
	if val, _ := m.VM.Get("events"); val.IsBoolean() {
		if enabled, _ := val.ToBoolean(); enabled {
			m.Handler = EventHandler{VM: m.VM}
		}
	}
	return err
}

type EventHandler struct {
	VM *otto.Otto
}

func (e EventHandler) HandlePacketReceive(ctx *event.Context, pk packet.Packet, ray *ray.Ray) {
	_, err := e.VM.Call("packet_receive", nil, ctx, pk, ray)
	if err != nil {
		panic(err)
	}
}

func (e EventHandler) HandlePacketSend(ctx *event.Context, pk packet.Packet, ray *ray.Ray) {
	_, err := e.VM.Call("packet_send", nil, ctx, pk, ray)
	if err != nil {
		panic(err)
	}
}
