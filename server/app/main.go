// app/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	v1 "smuggr.xyz/gatecam/api/v1"
	"smuggr.xyz/gatecam/common/config"
	"smuggr.xyz/gatecam/core/cameras"
	"smuggr.xyz/gatecam/core/devices"
)

func WaitForTermination() {
	callChan := make(chan os.Signal, 1)
	signal.Notify(callChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	fmt.Println("waiting for termination signal...")
	<-callChan
	fmt.Println("termination signal received")
}

func Cleanup() {
	fmt.Println("cleaning up...")

	cameras.Server.CloseAll()
}

func main() {
	if err := config.Initialize(); err != nil {
		panic(err)
	}

	if err := cameras.Initialize(); err != nil {
		panic(err)
	}

	devices.Initialize()

	errCh := v1.Initialize()

	defer Cleanup()

	if err := <-errCh; err != nil {
		panic(err)
	}

	WaitForTermination()
}
