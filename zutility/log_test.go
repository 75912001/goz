package zutility

import (
	"testing"
)

func TestLog(t *testing.T) {
	var log *Log = new(Log)
	log.Init("test_log", 1000)
	for i := 1; i < 200000; i++ {
		log.Emerg("debug")
	}
}
