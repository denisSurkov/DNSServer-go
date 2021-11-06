package main

import (
	"DNSServer/lib"
	"log"
)

func main() {
	exit := make(chan bool)
	go lib.RequestsReceiver(exit)

	defer func() {
		// In case program would not call exit on its own
		err := recover()
		log.Printf("caught err %s", err)
		exit <- true
	}()
	<-exit
	log.Println("DNS Server gracefully shut down")
}
