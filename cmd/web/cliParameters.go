package main

import "fmt"

type parameters struct {
	host              string
	port              int
	superuseremail    string
	superuserpwd      string
	https             bool
	testmode          bool
	domain            string
	redirectToHttps   bool
	useletsencrypt bool

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
