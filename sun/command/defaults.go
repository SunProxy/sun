package command

import (
	"github.com/fatih/color"
	"github.com/sunproxy/sun/sun/logger"
	"strings"
)

func (p Processor) RegisterDefaults() {
	p.Map.Register("help", Help{Map: p.Map}, true)
}

type Help struct {
	Map Map
}

func (h Help) Execute(logger logger.Logger) error {
	_ = logger.InfoColor(strings.Repeat("~", 10), color.New(color.FgYellow))
	for _, cmd := range h.Map {
		info := cmd.Info()
		_ = logger.Infof("/%s, Description: %s, Usage: %s", info.Name, info.Description, info.Usage)
	}
	err := logger.InfoColor(strings.Repeat("~", 10), color.New(color.FgYellow))
	return err
}

func (h Help) Info() CommandInfo {
	return CommandInfo{Name: "help", Description: "Returns help on all commands!", Usage: "help"}
}
