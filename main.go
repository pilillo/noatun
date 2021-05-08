package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var signalChannel chan os.Signal
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}

func main() {
	StartEndpoint()

	log.Println("Waiting for Ctrl+C...")
	waitForCtrlC()

	CloseEndpoint()
}
