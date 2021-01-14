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
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"io/ioutil"
	"os"
	"path/filepath"
)

func LoadResourcePacks(path string) []*resource.Pack {
	_, err := os.Stat(path)
	if err != nil {
		_ = os.Mkdir(path, 0644)
	}

	var packs []*resource.Pack

	files, _ := ioutil.ReadDir(path + "/")
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if ext != ".mcpack" && ext != ".zip" {
			fmt.Printf("Could not load resource pack %v: Invalid extension %v\n", f.Name(), ext)
			continue
		}

		pack, err := resource.Compile(path + "/" + f.Name())
		if err != nil {
			fmt.Printf("Could not load resource pack %v: %v\n", f.Name(), err)
			continue
		}

		packs = append(packs, pack)
		fmt.Printf("Resource pack %v v%v loaded!", f.Name(), pack.Version())
	}

	return packs
}
