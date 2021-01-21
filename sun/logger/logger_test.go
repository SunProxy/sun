package logger

import "testing"

func TestLogger_Log(t *testing.T) {
	log := New("sun.log", true)
	log.Log("Test", LogLevelDebug)
}
