package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	defaultBaseDirectory = home + "/.primes/"
	defaultStartingPrime = "1"
	defaultMaxFilesize   = 10000000
	defaultMaxBufferSize = 300
	defaultShowFails     = false
)

// IsConfigured returns whether the program is configured already
func IsConfigured() bool {
	return false
}

// RunConfigurator generates a program configuration according to
// user input
func RunConfigurator() {
	fmt.Println("A config will now be generated in $HOME/.primegenerator")
	base := getBaseDirectory()
	startingPrime := getStartingPrime()
	maxFilesize := getMaxFilesize()
	maxBufferSize := getMaxBufferSize()
	showFails := getShowFails()

	generateConfig(base, startingPrime, maxFilesize, maxBufferSize, showFails)
	fmt.Println("Done!")
}

// getBaseDirectory returns the user's preference for a base directory
func getBaseDirectory() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Base directory (default: %s/.primes/): ", home)
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	userChoice = strings.Trim(userChoice, " \n")
	if userChoice == "" {
		userChoice = defaultBaseDirectory
	}

	return userChoice
}

// getStartingPrime returns the user's preference for the prime to begin on
func getStartingPrime() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Prime to begin generation at (default: 1): ")
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	userChoice = strings.Trim(userChoice, " \n")
	if userChoice == "" {
		userChoice = defaultStartingPrime
	}

	return userChoice
}

// getMaxFilesize returns the user's preference for the maximum
// filesize
func getMaxFilesize() int {
	var userChoice int
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Maximum number of prime numbers in a file (default: 10000000): ")
	userChoiceString, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxFilesize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			log.Fatal(err)
		}
	}
	return userChoice
}

// getMaxBufferSize returns the user's preference for a maximum buffer
// size
func getMaxBufferSize() int {
	var userChoice int
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Maximum number of prime numbers in a buffer before flushing (default: 300): ")
	userChoiceString, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxBufferSize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			log.Fatal(err)
		}
	}
	return userChoice
}

// getShowFails returns the user's preference for whether to show fails or not
func getShowFails() bool {
	var userChoiceBoolean bool
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Show failed numbers (default: n) [y/n]: ")
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	if strings.ToLower(userChoice) == "y" {
		userChoiceBoolean = true
	} else if strings.ToLower(userChoice) == "n" {
		userChoiceBoolean = false
	}

	userChoice = strings.Trim(userChoice, " \n")
	if userChoice == "" {
		userChoiceBoolean = defaultShowFails
	}

	return userChoiceBoolean
}

// generateConfig concatinates the user's preferences into YAML format
func generateConfig(base string, startingPrime string, maxFilesize int, maxBufferSize int, showFails bool) {

}
