package main

import (
	"log"
	"os"
	"sync"
)

const (
	version  = "0.1.1"
	appName  = "PrimeNumberGenerator"
	appUsage = "Generate prime numbers forever"

	descConfigure = "Runs auto-configuration wizard"
	descCount     = "Displays the estimated curren n prime numbers"
	descRun       = "Begins computation of primes"

	descClient = "Launches a new instance of a client"
	descServer = "Launches a new instance of a server"
)

var (
	home              = GetUserHome()
	base              = home + "/.primes/"
	directory         = base + "directory.txt"
	configurationFile = home + "/.primegenerator.yaml"

	config        = Config{}
	startingPrime string
	maxFilesize   int
	maxBufferSize int
	showFails     bool

	logger = log.New(os.Stdout, "", log.LstdFlags)
	mu     sync.Mutex
)
