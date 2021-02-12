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

package plugin

import (
	"github.com/robertkrimen/otto"
	"github.com/sunproxy/sun/sun/logger"
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
}

func NewManager(log logger.Logger) *Manager {
	iso := otto.New()
	//just an alias to sun.Logger.FUNC
	_ = iso.Set("logger", log)
	return &Manager{Logger: log, Plugins: make(map[string]Plugin), VM: iso}
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
	return err
}
