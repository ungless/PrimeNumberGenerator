package main

var (
	home              = GetUserHome()
	base              = home + "/.primes/"
	directory         = base + "directory.txt"
	configurationFile = home + "/.primegenerator.yaml"

	startingPrime = "1"
	maxFilesize   = 10000000
	maxBufferSize = 200
	showFails     = false
)

const (
	version = "0.1.3"
)
