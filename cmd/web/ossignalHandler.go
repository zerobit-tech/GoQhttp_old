package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
)

// initialize signal handler
func initSignals(cleanUpFunc func()) {

	defer concurrent.Recoverer("initSignals")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	var captureSignal = make(chan os.Signal, 1)
	signal.Notify(captureSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	signalHandler(<-captureSignal, cleanUpFunc)
}

// signal handler
func signalHandler(signal os.Signal, cleanUpFunc func()) {

	fmt.Println("")
	log.Printf("Caught signal: %+v\n", signal)
	// log.Println("Wait for 1 second to finish processing")
	// time.Sleep(1 * time.Second)

	switch signal {

	case syscall.SIGHUP: // kill -SIGHUP XXXX
		log.Println("- got hungup")

	case syscall.SIGINT: // kill -SIGINT XXXX or Ctrl+c
		log.Println("- got ctrl+c")

	case syscall.SIGTERM: // kill -SIGTERM XXXX
		log.Println("- got force stop")

	case syscall.SIGQUIT: // kill -SIGQUIT XXXX
		log.Println("- stop and core dump")

	default:
		log.Println("- unknown signal")
	}

	cleanUpFunc()
	log.Println("Finished server cleanup")
	fmt.Println("---")
	time.Sleep(1 * time.Second)

	os.Exit(0)
}
