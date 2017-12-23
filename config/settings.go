package config

import (
	"log"
	"os"
	"sync"

	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"
)

const (
	version  = "0.4.4"
	appName  = "PrimeNumberGenerator"
	appUsage = "Generate prime numbers forever"

	descConfigure = "Runs auto-configuration wizard"
	descCount     = "Displays the estimated curren n prime numbers"
	descRun       = "Begins computation of primes"

	descClient = "Launches a new instance of a client"
	descServer = "Launches a new instance of a server"
)

var (
	home              = storage.GetUserHome()
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
