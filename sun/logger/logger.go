package logger

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"io"
	"os"
	"regexp"
	"time"
)

type Logger struct {
	File *os.File
	//Whether or not the Logger should print debug type logs.
	ShowDebug bool
	//Should only be the stdout.
	Stdout io.Writer
}

const (
	LogLevelInfo = iota
	LogLevelDebug
	LogLevelSuccess
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

//used to clean ascii codes yes it stolen lmao
var reg = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func (l Logger) Info(Message string) error {
	return l.Log(Message, LogLevelInfo)
}

func (l Logger) Debug(Message string) error {
	return l.Log(Message, LogLevelDebug)
}

func (l Logger) Success(Message string) error {
	return l.Log(Message, LogLevelSuccess)
}

func (l Logger) Warn(Message string) error {
	return l.Log(Message, LogLevelWarning)
}

func (l Logger) Error(Error string) error {
	return l.Log(Error, LogLevelError)
}

func (l Logger) Fatal(Error string) error {
	return l.Log(Error, LogLevelFatal)
}

func (l Logger) Infof(Message string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Message, a...), LogLevelInfo)
}

func (l Logger) Debugf(Message string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Message, a...), LogLevelDebug)
}

func (l Logger) Successf(Message string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Message, a...), LogLevelSuccess)
}

func (l Logger) Warnf(Message string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Message, a...), LogLevelWarning)
}

func (l Logger) Errorf(Error string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Error, a...), LogLevelError)
}

func (l Logger) Fatalf(Error string, a ...interface{}) error {
	return l.Log(fmt.Sprintf(Error, a...), LogLevelFatal)
}

func (l Logger) InfoColor(Message string, c *color.Color) error {
	return l.Log(c.Sprint(Message), LogLevelInfo)
}

func (l Logger) InfoColorf(c *color.Color, Message string, a ...interface{}) error {
	return l.Log(c.Sprintf(Message, a...), LogLevelInfo)
}

func (l Logger) Log(Message string, Level uint16) error {
	Message += "\n"
	switch Level {
	case LogLevelInfo:
		_, err := l.Stdout.Write([]byte(Message))
		if err != nil {
			return err
		}
		_, err = l.File.WriteString(reg.ReplaceAllString(Message, ""))
		return err
	case LogLevelDebug:
		if l.ShowDebug {
			_, err := l.Stdout.Write([]byte(color.BlueString("DEBUG[%s] %s", time.Now().String(), Message)))
			if err != nil {
				return err
			}
			_, err = l.File.WriteString(fmt.Sprintf("DEBUG[%s] %s", time.Now().String(),
				reg.ReplaceAllString(Message, "")))
			return err
		}
	case LogLevelSuccess:
		_, err := l.Stdout.Write([]byte(color.GreenString(Message)))
		if err != nil {
			return err
		}
		_, err = l.File.WriteString(reg.ReplaceAllString(Message, ""))
		return err
	case LogLevelWarning:
		_, err := l.Stdout.Write([]byte(color.HiYellowString("WARN[%s] %s", time.Now().String(), Message)))
		if err != nil {
			return err
		}
		_, err = l.File.WriteString(fmt.Sprintf("WARN[%s] %s", time.Now().String(),
			reg.ReplaceAllString(Message, "")))
		return err
	case LogLevelError:
		_, err := l.Stdout.Write([]byte(color.RedString("ERROR[%s] %s", time.Now().String(), Message)))
		if err != nil {
			return err
		}
		_, err = l.File.WriteString(fmt.Sprintf("ERROR[%s] %s", time.Now().String(),
			reg.ReplaceAllString(Message, "")))
		return err
	case LogLevelFatal:
		_, err := l.Stdout.Write([]byte(color.HiRedString("FATAL[%s] %s", time.Now().String(), Message)))
		if err != nil {
			return err
		}
		_, err = l.File.WriteString(fmt.Sprintf("FATAL[%s] %s", time.Now().String(),
			reg.ReplaceAllString(Message, "")))
		os.Exit(1)
	}
	return nil
}

func New(FileName string, debug bool) Logger {
	File, _ := os.OpenFile(FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return Logger{
		File:      File,
		ShowDebug: debug,
		Stdout:    colorable.NewColorableStdout(),
	}
}
