package main

import (
	"errors"
	"flag"
	"strings"
)

type parameters struct {
	client     string
	email      string
	days       int
	checkemail bool

	//staticDir string
	//flag      bool
}

func (p *parameters) Validate() error {

	if p.checkemail {
		return nil
	}
	p.client = strings.TrimSpace(p.client)
	p.email = strings.TrimSpace(p.email)

	if p.client == "" {
		return errors.New("Client name is required.")

	}

	if p.email == "" {
		return errors.New("Client email is required.")

	}

	if p.days <= 0 {
		return errors.New("Expiry days must be > zero")

	}

	return nil
}

func (params *parameters) Load() {
	flag.StringVar(&params.client, "client", "", "Client Name")
	flag.StringVar(&params.email, "email", "", "Client Email")
	flag.IntVar(&params.days, "days", 365, "Expiry days")
	flag.BoolVar(&params.checkemail, "checkemail", false, "Start checking emails.")

	flag.Parse()
}
