package main

import (
	"errors"
	"flag"
	"strings"
)

type parameters struct {
	client     string
	email      string
	expiryDays int

	//staticDir string
	//flag      bool
}

func (p *parameters) Validate() error {
	p.client = strings.TrimSpace(p.client)
	p.email = strings.TrimSpace(p.email)

	if p.client == "" {
		return errors.New("Client name is required.")

	}

	if p.email == "" {
		return errors.New("Client email is required.")

	}

	if p.expiryDays <= 0 {
		return errors.New("Expiry days must be > zero")

	}

	return nil
}

func (params *parameters) Load() {
	flag.StringVar(&params.client, "client", "", "Client Name")
	flag.StringVar(&params.email, "email", "", "Client Email")
	flag.IntVar(&params.expiryDays, "expiryDays", 365, "Expiry days")

	flag.Parse()
}
