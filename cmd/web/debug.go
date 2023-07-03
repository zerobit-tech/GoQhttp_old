package main

import (
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/onlysumitg/GoQhttp/env"
)

func debugMe(params parameters) {
	if env.IsInDebugMode() {
		//goroutine
		go func() {
			addr, _ := params.getHttpAddressForProfile()

			log.Printf("Profiling Server is active a port(http) %s%s \n", addr, "/debug/pprof/")
			log.Println(http.ListenAndServe(addr, nil))

		}()
	}

}
