package main

import (
	"log"
	"os"
)

const (
	version = "0.1.3"
)

var (
	home              = GetUserHome()
	base              = home + "/.primes/"
	directory         = base + "directory.txt"
	configurationFile = home + "/.primegenerator.yaml"

	config        = GetUserConfig()
	startingPrime = config.StartingPrime
	maxFilesize   = config.MaxFilesize
	maxBufferSize = config.MaxBufferSize
	showFails     = config.ShowFails

	logger = log.New(os.Stdout, "", log.LstdFlags)
)
