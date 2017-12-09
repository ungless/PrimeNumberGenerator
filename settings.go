package main

import (
	"log"
	"os"
	"sync"
)

const (
	version = "0.2.2"
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
