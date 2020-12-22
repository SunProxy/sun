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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

/**
The basic Config file struct.
*/
type Config struct {
	Status minecraft.ServerStatus

	Hub IpAddr

	Port uint16

	Tcp struct {
		Enabled bool

		Servers []string
	}
}

func LoadConfig() (Config, error) {
	if _, err := os.Stat("config.toml"); !os.IsNotExist(err) {
		return LoadTomlConfig()
	}
	if _, err := os.Stat("config.json"); !os.IsNotExist(err) {
		return LoadJsonConfig()
	}
	if _, err := os.Stat("config.yml"); !os.IsNotExist(err) {
		return LoadYamlConfig()
	}
	if _, err := os.Stat("config.xml"); !os.IsNotExist(err) {
		return LoadXmlConfig()
	}
	if _, err := os.Stat("config.gob"); !os.IsNotExist(err) {
		return LoadGobConfig()
	}
	return LoadYamlConfig()
}

func LoadTomlConfig() (Config, error) {
	config := Config{}
	data, _ := ioutil.ReadFile("config.toml")
	_ = toml.Unmarshal(data, &config)
	config = defaultConfig(config)
	data, _ = toml.Marshal(config)
	_ = ioutil.WriteFile("config.toml", data, 0644)
	return config, nil
}

func LoadJsonConfig() (Config, error) {
	config := Config{}
	data, _ := ioutil.ReadFile("config.json")
	_ = json.Unmarshal(data, &config)
	config = defaultConfig(config)
	data, _ = json.Marshal(config)
	_ = ioutil.WriteFile("config.json", data, 0644)
	return config, nil
}

func LoadYamlConfig() (Config, error) {
	config := Config{}
	data, _ := ioutil.ReadFile("config.yml")
	_ = yaml.Unmarshal(data, &config)
	config = defaultConfig(config)
	data, _ = yaml.Marshal(config)
	_ = ioutil.WriteFile("config.yml", data, 0644)
	return config, nil
}

func LoadXmlConfig() (Config, error) {
	config := Config{}
	data, _ := ioutil.ReadFile("config.xml")
	_ = xml.Unmarshal(data, &config)
	config = defaultConfig(config)
	data, _ = xml.Marshal(config)
	_ = ioutil.WriteFile("config.xml", data, 0644)
	return config, nil
}

func LoadGobConfig() (Config, error) {
	config := Config{}
	data, _ := ioutil.ReadFile("config.gob")
	dec := gob.NewDecoder(bytes.NewReader(data))
	_ = dec.Decode(&config)
	config = defaultConfig(config)
	datab := bytes.Buffer{}
	enc := gob.NewEncoder(&datab)
	_ = enc.Encode(config)
	_ = ioutil.WriteFile("config.gob", datab.Bytes(), 0644)
	return config, nil
}

/**
Should take in a empty config
*/
func defaultConfig(config Config) Config {
	if config.Port == 0 {
		config.Port = 19132
	}
	emptyIp := IpAddr{}
	if config.Hub == emptyIp {
		config.Hub.Port = 19133
		config.Hub.Address = "0.0.0.0"
	}
	emptyStatus := minecraft.ServerStatus{}
	if config.Status == emptyStatus {
		config.Status.MaxPlayers = 50
		config.Status.PlayerCount = 0
		config.Status.ServerName = text.Colourf("<yellow>Sun Proxy</yellow>")
	}
	return config
}
