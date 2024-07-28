package main

import (
	"log"
	"os"
)

var (
	CommandPortPath    = getEnv("COMMAND_PORT_PATH", "/dev/ttyUSB0")
	NotifyPortPath     = getEnv("NOTIFY_PORT_PATH", "/dev/ttyUSB2")
	DataPersistencePath = getEnv("DATA_PERSISTENCE_PATH", "inbox.json")
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	AssertHuaweiModemMode()
	mon := NewMonitor(CommandPortPath, NotifyPortPath, DataPersistencePath)
	log.Printf("Initialized modem")

	log.Printf("Staring daemon at http://localhost:8080")
	if err := mon.Run(); err != nil {
		log.Fatalln(err)
	}
}
