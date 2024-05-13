package log

import (
	"log"
	"testing"
)

func TestLevel(t *testing.T) {
	logex := New("test", true)
	logex.colorFile = NewColorFile()
	logex.colorFile.Load("color_sample.json")
	logex.enableColor = true
	logex.Debugf("%d %s %v", 1, "hello", "world")
	logex.Error("hello1")
	logex.Error("2")
	logex.Info("no")

}

func TestMyLog(t *testing.T) {
	logex := New("test2", true)
	logex.colorFile = NewColorFile()
	logex.colorFile.Load("color_sample.json")
	logex.enableColor = true
	logex.Debug("hello1")

	logex.Debug("hello3")
}

func TestSystemLog(t *testing.T) {
	log.Println("hello1")
	log.Println("hello2")
	log.Println("hello3")
}
