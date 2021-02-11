package command

import "github.com/sunproxy/sun/sun/logger"

type Command interface {
	Execute(logger.Logger) error
	Info() CommandInfo
}

type CommandInfo struct {
	Name        string
	Description string
	Usage       string
}
