package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/onlysumitg/GoQhttp/ssl"
	"golang.org/x/crypto/acme/autocert"
)

func (app *application) getCertificateToUse() *tls.Config {

	log.Println("certi::: using", app.domain)
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(app.domain),
		Cache:      autocert.DirCache("certs"),
	}

	tlsConfig := certManager.TLSConfig()
	tlsConfig.GetCertificate = app.getSelfSignedOrLetsEncryptCert(&certManager)

	return tlsConfig
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) getSelfSignedOrLetsEncryptCert(certManager *autocert.Manager) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {

		if app.useletsencrypt {

			fmt.Printf("\nFalling back to Letsencrypt\n")
			c, err := certManager.GetCertificate(hello)
			if err != nil {
				log.Panicln("Letsencrypt failed:", err)
			}
			return c, err

		} else {
			fmt.Println("Loaded selfsigned certificate.")
			c, err := getSelfSignedCertificate()

			if err != nil {
				log.Panicln("Self signed certificate failed:", err)
			}
			return c, err
		}
	}
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func getSelfSignedCertificate() (*tls.Certificate, error) {
	gomockapi_crt, err := ssl.SSLCertificats.ReadFile("cert/gomockapi.crt")
	if err != nil {
		return &tls.Certificate{}, err
	}
	gomockapi_api, err := ssl.SSLCertificats.ReadFile("cert/gomockapi.key")
	if err != nil {
		return &tls.Certificate{}, err
	}
	cert, err := tls.X509KeyPair(gomockapi_crt, gomockapi_api)
	if err != nil {
		return &tls.Certificate{}, err
	}

	return &cert, nil
}
