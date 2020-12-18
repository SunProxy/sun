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
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/text"
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
}

func LoadConfig() (Config, error) {
	if _, err := os.Stat("config.toml"); !os.IsNotExist(err) {
		data, _ := ioutil.ReadFile("config.toml")
		var config Config
		err = toml.Unmarshal(data, &config)
		if err != nil {
			return Config{}, err
		}
	}
	return Config{}, nil
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
		config.Hub.Ip = "0.0.0.0"
	}
	emptyStatus := minecraft.ServerStatus{}
	if config.Status == emptyStatus {
		config.Status.MaxPlayers = 50
		config.Status.PlayerCount = 0
		config.Status.ServerName = text.Colourf("<>")
	}
	return config
}
