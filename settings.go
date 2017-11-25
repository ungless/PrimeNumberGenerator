package main

var (
	home      = GetUserHome()
	base      = home + "/.primes/"
	directory = base + "directory.txt"
)

const (
	version       = "0.0.1"
	startingPrime = "1"
	maxFilesize   = 10000000
	maxBufferSize = 200
	showFails     = false
)
