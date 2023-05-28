package main

import (
	"log"

	"net/http"
	_ "net/http/pprof"
)

func debugMe(params parameters) {
	if params.testmode {
		go func() {
			addr, _ := params.getHttpAddressForProfile()

			log.Printf("Profiling Server is active a port(http) %s%s \n", addr, "/debug/pprof/")
			log.Println(http.ListenAndServe(addr, nil))

		}()
	}

}
