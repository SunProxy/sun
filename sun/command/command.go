package command

import "github.com/sunproxy/sun/sun/logger"

type Command interface {
	Execute(logger.Logger) error
	Info() Info
}

type Info struct {
	Name        string
	Description string
	Usage       string
}
