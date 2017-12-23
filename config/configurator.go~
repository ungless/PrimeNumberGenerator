// This is the file for the Configurator - something which generates
// a YAML config for this program.
//
// An example (with the default values):
//     base: /home/max/.primes/
//     startingprime: 1
//     maxfilesize: 10000000
//     maxbuffersize: 300
//     showfails: false

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

var (
	defaultBaseDirectory = home + "/.primes/"
	defaultStartingPrime = "1"
	defaultMaxFilesize   = 10000000
	defaultMaxBufferSize = 300
	defaultShowFails     = false
)

type Config struct {
	Base          string `json:"base"`
	StartingPrime string `json:"startingprime"`
	MaxFilesize   int    `json:"maxfilesize"`
	MaxBufferSize int    `json:"maxbuffersize"`
	ShowFails     bool   `json:"showfails"`
}

// GetUserConfig returns a Config object containing the user's configuration
func GetUserConfig() Config {
	logger.Print("Searching for user's configuration")
	config := Config{}
	if IsConfigured() {
		y, err := ioutil.ReadFile(configurationFile)
		if err != nil {
			logger.Fatal(err)
		}
		err = yaml.Unmarshal(y, &config)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Print("Found user's already existing configuration")
		return config
	} else {
		logger.Print("No configuration found")
		EnsureUserWantsNewConfig()
		logger.Fatal("Restart the program in order to apply this configuration.")
	}
	return config
}

// EnsureUserWantsNewConfig ensures user wants a new config and if so, runs the
// configurator
func EnsureUserWantsNewConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("A configuration file could not be found.\nWould you like to generate one now? [y/n] ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		logger.Fatal(err)
	}
	choice = strings.Trim(choice, " \n")

	if strings.ToLower(choice) == "y" {
		RunConfigurator()
	} else {
		os.Exit(1)
	}

}

// IsConfigured returns whether the program is configured already
func IsConfigured() bool {
	if _, err := os.Stat(configurationFile); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// RunConfigurator generates a program configuration according to
// user input
func RunConfigurator() {
	fmt.Printf("A config will now be generated in %s\n", configurationFile)
	base := getBaseDirectory()
	startingPrime := getStartingPrime()
	maxFilesize := getMaxFilesize()
	maxBufferSize := getMaxBufferSize()
	showFails := getShowFails()

	generateConfig(base, startingPrime, maxFilesize, maxBufferSize, showFails)
	fmt.Println("Your configuration has now been generated.")
}

// getBaseDirectory returns the user's preference for a base directory
func getBaseDirectory() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Base directory (default: %s/.primes/): ", home)
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		logger.Fatal(err)
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
		logger.Fatal(err)
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
		logger.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxFilesize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			logger.Fatal(err)
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
		logger.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxBufferSize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			logger.Fatal(err)
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
		logger.Fatal(err)
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
	config, err := os.Create(home + "/.primegenerator.yaml")
	defer config.Close()
	if err != nil {
		logger.Fatal(err)
	}
	c := Config{base, startingPrime, maxFilesize, maxBufferSize, showFails}
	yaml, err := yaml.Marshal(c)
	if err != nil {
		logger.Fatal(err)
	}
	config.Write(yaml)
}

// ensureConfigExists continually checks whether the user's config is loaded exists on the user's system
func ensureConfigExists() {
	configLoaded := false
	for configLoaded == false {
		if id != 0 {
			configLoaded = true
		}
	}
}

// SetConfiguration sets the global configuration variables
func SetConfiguration() {
	config = GetUserConfig()
	startingPrime = config.StartingPrime
	maxFilesize = config.MaxFilesize
	maxBufferSize = config.MaxBufferSize
	showFails = config.ShowFails
}

// SetId sets the gloabl id variable
func SetId() {
	id = GetCurrentId()
}

// SetLastPrimeGenerated sets the global lastprimegenerated variable
func SetLastPrimeGenerated() {
	lastPrimeGenerated = getLastPrime()
}
