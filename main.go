package main

import "log"

const (
	CommandPortPath = "/dev/ttyUSB0"
	NotifyPortPath  = "/dev/ttyUSB2"
)

func main() {
	AssertHuaweiModemMode()
	log.Printf("Staring daemon at http://localhost:8080")
	mon := NewMonitor(CommandPortPath, NotifyPortPath, "inbox.json")

	if err := mon.Run(); err != nil {
		log.Fatalln(err)
	}
}
