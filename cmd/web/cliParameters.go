package main

import (
	"flag"
	"fmt"

	"github.com/onlysumitg/GoQhttp/env"
)

type parameters struct {
	host            string
	port            int
	superuseremail  string
	superuserpwd    string
	https           bool
	testmode        bool
	domain          string
	redirectToHttps bool
	useletsencrypt  bool

	//staticDir string
	//flag      bool
}

func (p *parameters) getHttpAddress() (string, string) {
	addr := p.host

	if p.port > 0 {
		addr = fmt.Sprintf("%s:%d", addr, p.port)
	}

	protocol := "http://"
	if p.https || p.redirectToHttps {
		protocol = "https://"
	}

	if p.domain == "localhost" || p.domain == "0.0.0.0" {
		p.domain = fmt.Sprintf("%s:%d", p.domain, p.port)
	}

	return addr, fmt.Sprintf("%s%s", protocol, p.domain)
}

func (p *parameters) getHttpAddressForProfile() (string, string) {
	addr := p.host

	port := env.GetEnvVariable("PPROFPORT", "6060")

	addr = fmt.Sprintf("%s:%s", addr, port)

	protocol := "http://"
	// if p.https || p.redirectToHttps {
	// 	protocol = "https://"
	// }

	if p.domain == "localhost" || p.domain == "0.0.0.0" {
		p.domain = fmt.Sprintf("%s:%s", p.domain, port)
	}

	return addr, fmt.Sprintf("%s%s", protocol, p.domain)
}

func (params *parameters) Load() {
	flag.StringVar(&params.host, "host", "", "Http Host Name")
	flag.IntVar(&params.port, "port", 4081, "Port")

	flag.StringVar(&params.superuseremail, "superuseremail", "admin2@example.com", "Super User email")
	flag.StringVar(&params.superuserpwd, "superuserpwd", "adminpass", "Super User password")

	flag.BoolVar(&params.https, "https", true, "Use http or https")
	flag.BoolVar(&params.useletsencrypt, "useletsencrypt", false, "Use let's encrypt ssl certificate")

	flag.BoolVar(&params.testmode, "testmode", false, "Enable test mode")
	flag.StringVar(&params.domain, "domain", "0.0.0.0", "Domain name")

	flag.BoolVar(&params.redirectToHttps, "redirecttohttps", false, "Redirect to https")

	flag.Parse()
}
