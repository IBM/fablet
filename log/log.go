package log

import (
	"fmt"
	"log"
	"sync"
)

// TODO To log to file.

type FabletLogger struct {
}

var commonLogger *FabletLogger
var once sync.Once

func Init() {
	once.Do(func() {
		if commonLogger == nil {
			log.Println("Initialize log.")
			commonLogger = &FabletLogger{}
		}
	})
}

func GetLogger() *FabletLogger {
	return commonLogger
}

func (logger *FabletLogger) Info(msg ...interface{}) {
	log.Println("[INF]", fmt.Sprint(msg...))
}

func (logger *FabletLogger) Infof(format string, msg ...interface{}) {
	log.Println("[INF]", fmt.Sprintf(format, msg...))
}

func (logger *FabletLogger) Debug(msg ...interface{}) {
	log.Println("[DBG]", fmt.Sprint(msg...))
}
func (logger *FabletLogger) Debugf(format string, msg ...interface{}) {
	log.Println("[DBG]", fmt.Sprintf(format, msg...))
}

func (logger *FabletLogger) Error(msg ...interface{}) {
	log.Println("[ERR]", fmt.Sprint(msg...))
}
func (logger *FabletLogger) Errorf(format string, msg ...interface{}) {
	log.Println("[ERR]", fmt.Sprintf(format, msg...))
}

func (logger *FabletLogger) Warn(msg ...interface{}) {
	log.Println("[WRN]", fmt.Sprint(msg...))
}
