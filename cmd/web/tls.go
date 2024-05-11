package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	embdedTLS "github.com/zerobit-tech/GoQhttp/tls"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

//https://www.captaincodeman.com/automatic-https-with-free-ssl-certificates-using-go-lets-encrypt

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func (app *application) getCertificateAndManager() (*tls.Config, *autocert.Manager) {

	//log.Println("certi::: using", app.domain)
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(app.domain),
		Cache:      autocert.DirCache("certs"),
	}
	tlsConfig := &tls.Config{
		Rand:           rand.Reader,
		Time:           time.Now,
		NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
		MinVersion:     tls.VersionTLS12,
		GetCertificate: app.getSelfSignedOrLetsEncryptCert(certManager),
	}
	return tlsConfig, certManager
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func (app *application) getCertificateToUseOrg() *tls.Config {

	//log.Println("certi::: using", app.domain)
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

		app.tlsMutex.Lock()
		defer app.tlsMutex.Unlock()

		if app.tlsCertificate != nil {
			return app.tlsCertificate, nil
		}
		if app.useletsencrypt {

			fmt.Printf("\nUsing Letsencrypt\n")
			c, err := certManager.GetCertificate(hello)
			if err != nil {
				log.Panicln("Letsencrypt failed:", err)
			}

			app.tlsCertificate = c

			return c, err

		} else {

			// first check cert folder for goqhttp.crt and goqhttp.key
			log.Println("Loading self signed certificate from cert directory...")
			c, err := getSelfSignedCertificate()

			if err == nil {
				app.tlsCertificate = c

				return c, nil
			}
			log.Println("Loading self signed certificate from cert directory failed.", err.Error())

			// check embded certificat
			log.Println("Loading emdedded self signed certificate ...")

			c, err = getEmdededSelfSignedCertificate()

			if err != nil {
				log.Panicln("Self signed certificate failed:", err)
			}
			app.tlsCertificate = c
			return c, err
		}
	}
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func getEmdededSelfSignedCertificate() (*tls.Certificate, error) {

	goqhttp_crt, err := embdedTLS.SSLCertificats.ReadFile("cert/goqhttp.crt")
	if err != nil {
		return &tls.Certificate{}, err
	}
	goqhttp_api, err := embdedTLS.SSLCertificats.ReadFile("cert/goqhttp.key")
	if err != nil {
		return &tls.Certificate{}, err
	}
	cert, err := tls.X509KeyPair(goqhttp_crt, goqhttp_api)
	if err != nil {
		return &tls.Certificate{}, err
	}

	return &cert, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func getSelfSignedCertificate() (*tls.Certificate, error) {

	goqhttp_crt, err := os.ReadFile("cert/qhttp.crt")
	if err != nil {
		return &tls.Certificate{}, err
	}
	goqhttp_api, err := os.ReadFile("cert/qhttp.key")
	if err != nil {
		return &tls.Certificate{}, err
	}
	cert, err := tls.X509KeyPair(goqhttp_crt, goqhttp_api)
	if err != nil {
		return &tls.Certificate{}, err
	}

	return &cert, nil
}
