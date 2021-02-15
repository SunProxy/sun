package command

import (
	"bufio"
	"github.com/sunproxy/sun/sun/logger"
	"os"
	"regexp"
	"strings"
)

var CommandRegex = regexp.MustCompile(`(?m)("[^"]+"|[^\s"]+)`)

type Processor struct {
	Map     Map
	On      func(cmd Command)
	Logger  logger.Logger
	running bool
}

/*
StartProcessing starts the command processor reading from the given std in (*io.File in golang terms.)
*/
func (p Processor) StartProcessing(in *os.File) {
	p.running = true
	go func() {
		scnr := bufio.NewScanner(in)
		// so when ProcStop is called the goroutine isn't sitting dead.
		for p.running {
			scnr.Scan()
			line := scnr.Text()
			args := strings.Split(line, " ")
			cmd, err := p.Map.Get(args[0])
			if err != nil {
				_ = p.Logger.Errorf("Unknown Command %s", args[0])
				continue
			}
			_ = cmd.Execute(p.Logger)
		}
		return
	}()
}

/*
StopProcessing stops the command processor from occupying the stdin and will stop all future command execution.
*/
func (p Processor) StopProcessing() {
	p.running = false
}

func NewProcessor(logger logger.Logger, callback func(cmd Command)) Processor {
	return Processor{Map: NewMap(), On: callback, Logger: logger}
}
