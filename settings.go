package main

import (
	"log"
	"math/big"
	"os"
	"sync"
)

const (
	version = "0.2.1"
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

	globalCount        = big.NewInt(0)
	id          uint64 = 0
	mu          sync.Mutex
)
