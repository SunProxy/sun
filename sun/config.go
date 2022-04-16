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
	"math/rand"
	"os"
	"time"
)

// Config is the configuration file for the sun proxy.
type Config struct {
	Status minecraft.ServerStatus

	// Hub The First Server the proxy will redirect players too (after @see LoadBalancer)
	Hub string

	Proxy struct {

		// Port The Sun Proxy should run on.
		Port uint16

		// XboxAuthentication A boolean representing wether or not the proxy should use Xbox Auth.
		XboxAuthentication bool

		// IPForwarding Unused as of now.
		IPForwarding bool

		// LoadBalancer Used to redirect players to a new server after the hub is unusable.
		LoadBalancer struct {

			// Enabled A boolean value representing if sun should use LoadBalancing
			Enabled bool

			// Balancers A list of servers to try to balance too after the hub is unusable.
			Balancers []string
		}

		// StatusCommand A boolean value representing if the /status command should be overridden.
		StatusCommand bool

		// Pprof An embedded struct used to indicate if the proxy should enable pprof profiling and what port the webserver should run on.
		Pprof struct {

			// Enabled A boolean representing whether the pprof webserver should be started.
			Enabled bool

			// Port The port at which the pprof webserver should run on.
			Port uint16
		}

		Logger struct {

			// File The path to the file which the logger entries should be written to
			File string

			// Debugs Represents if the logger should display debug messages!
			Debug bool
		}

		// MOTDFoward A boolean representing whether to forward the motd to the hub or the first open LoadBalancer
		MOTDForward bool
	}

	// TODO: REDO ALL THIS SHIT

	/*
		The configuration for the tcp server in sun
	*/
	Tcp struct {

		/*
			Specifies if the proxy should run the tcp server
		*/
		Enabled bool

		/*
			Used to login into the tcp server
		*/
		Key string
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
	data, _ = json.MarshalIndent(config, "", " ")
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

/*
Should take in a empty config
*/
func defaultConfig(config Config) Config {
	if config.Proxy.Port == 0 {
		config.Proxy.Port = 19132
	}

	if config.Hub == "" {
		config.Hub = "hub-1.mydomain.com:19132"
	}

	emptyStatus := minecraft.ServerStatus{}
	if config.Status == emptyStatus {
		config.Status.MaxPlayers = 50
		config.Status.PlayerCount = 0
		config.Status.ServerName = text.Colourf("<yellow>Sun Proxy</yellow>")
	}

	//Generate a random Key if its empty
	if config.Tcp.Key == "" {
		config.Tcp.Key = GenKey()
	}

	if config.Proxy.LoadBalancer.Balancers == nil {
		config.Proxy.LoadBalancer.Balancers = make([]string, 1)
		config.Proxy.LoadBalancer.Balancers[0] = "hub-2.mydomain.com:19132"
	}

	if config.Proxy.Pprof.Port == 0 {
		config.Proxy.Pprof.Port = 8080
	}

	if config.Proxy.Logger.File == "" {
		config.Proxy.Logger.File = "sun.log"
	}
	return config
}

func GenKey() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	Chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var key = make([]rune, 25)
	for i := range key {
		key[i] = rune(Chars[rnd.Intn(len(Chars))])
	}
	return string(key)
}
