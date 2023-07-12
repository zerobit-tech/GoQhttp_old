package concurrent

import (
	"fmt"
	"log"
	"runtime/debug"
)

func Recoverer(id string) {
	if r := recover(); r != nil {
		fmt.Println("Recovered ", id, " ", r)
	}
}

func RecoverAndRestart(maxPanics int, id string, f func()) { 

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("HERE", id)
			fmt.Println(err)
			if maxPanics == 0 {
				log.Println("Recovered but can not restart....:", id)
			} else {
				go RecoverAndRestart(maxPanics-1, id, f)
			}
		}
	}()

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	// call the actual function
	f()
}
