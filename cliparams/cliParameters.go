package cliparams

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/onlysumitg/GoQhttp/env"
)

type Parameters struct {
	Host           string
	Port           int
	Superuseremail string
	Superuserpwd   string

	Domain string
	//redirectToHttps bool
	Useletsencrypt bool
	ValidateSetup  bool

	Featureset string

	//staticDir string
	//flag      bool
}

func (p *Parameters) GetHttpAddress() (string, string) {
	addr := p.Host

	if p.Port > 0 {
		addr = fmt.Sprintf("%s:%d", addr, p.Port)
	}

	protocol := "https://"

	if p.Domain == "localhost" || p.Domain == "0.0.0.0" {
		p.Domain = fmt.Sprintf("%s:%d", p.Domain, p.Port)
	}

	return addr, fmt.Sprintf("%s%s", protocol, p.Domain)
}

func (p *Parameters) GetHttpAddressForProfile() (string, string) {
	addr := p.Host

	port := env.GetEnvVariable("PPROFPORT", "6060")

	addr = fmt.Sprintf("%s:%s", addr, port)

	protocol := "http://"
	// if p.https || p.redirectToHttps {
	// 	protocol = "https://"
	// }

	if p.Domain == "localhost" || p.Domain == "0.0.0.0" {
		p.Domain = fmt.Sprintf("%s:%s", p.Domain, port)
	}

	return addr, fmt.Sprintf("%s%s", protocol, p.Domain)
}

func (params *Parameters) Load() {
	flag.StringVar(&params.Host, "host", "", "Http Host Name")
	flag.IntVar(&params.Port, "port", 4081, "Port")

	flag.StringVar(&params.Superuseremail, "superuseremail", "admin2@example.com", "Super User email")
	flag.StringVar(&params.Superuserpwd, "superuserpwd", "adminpass", "Super User password")

	flag.BoolVar(&params.Useletsencrypt, "useletsencrypt", false, "Use let's encrypt ssl certificate")
	flag.BoolVar(&params.ValidateSetup, "validate", false, "Validate os setup")

	domain := "0.0.0.0"
	if runtime.GOOS == "windows" {
		domain = "localhost"
	}

	flag.StringVar(&params.Domain, "domain", domain, "Domain name")

	//flag.BoolVar(&params.redirectToHttps, "redirecttohttps", false, "Redirect to https")

	flag.Parse()

	envPort := os.Getenv("PORT")
	port, err := strconv.Atoi(envPort)
	if err == nil {

		params.Port = port
		log.Println("Using port>>> ", port, params.Port)
	}
}
