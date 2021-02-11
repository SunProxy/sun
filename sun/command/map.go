package command

import "fmt"

type Map map[string]Command

func NewMap() Map {
	return make(map[string]Command)
}


/*
Register adds the given Command to by name to the Map with a option to override if the said command already exists.
*/
func (m Map) Register(name string, cmd Command, override bool) {
	if _, ok := m[name]; ok && override {
		m[name] = cmd
	} else if _, ok := m[name]; !ok {
		m[name] = cmd
	}
}

/*
Unregister removes the given command by its name from the command map.
*/
func (m Map) Unregister(name string) {
	if _, ok := m[name]; ok {
		delete(m, name)
	}
}

/*
Get returns the given command registered to this map by its name.
ex `m.Get("help")`
*/
func (m Map) Get(name string) (Command, err) {
	if cmd, ok := m[name]; ok {
		return cmd, nil
	}
	return nil, fmt.Errorf("unknown command: %s", name)
}
