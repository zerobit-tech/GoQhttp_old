package main

import (
	"log"
	"runtime/debug"

	"net/http"
	_ "net/http/pprof"

	"github.com/onlysumitg/GoQhttp/cliparams"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

func debugMe(params cliparams.Parameters) {
	if env.IsInDebugMode() {
		//goroutine
		go func() {
			defer concurrent.Recoverer("debugMe")
			defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

			addr, _ := params.GetHttpAddressForProfile()

			log.Printf("Profiling Server is active a port(http) %s%s \n", addr, "/debug/pprof/")
			log.Println(http.ListenAndServe(addr, nil))

		}()
	}

}
