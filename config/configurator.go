// This is the file for the Configurator - something which generates
// a YAML config for this program.
//
// An example (with the default values):
//     base: /home/max/.primes/
//     startingprime: 1
//     maxfilesize: 10000000
//     maxbuffersize: 300
//     showfails: false

package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
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
	defaultServerIP      = "192.168.1.66"
)

type Config struct {
	Base          string `json:"base"`
	StartingPrime string `json:"startingprime"`
	MaxFilesize   int    `json:"maxfilesize"`
	MaxBufferSize int    `json:"maxbuffersize"`
	ShowFails     bool   `json:"showfails"`
	ServerIP      string `json:"serverip"`
}

// GetUserHome returns the current user's home directory
func GetUserHome() string {
	currentUser, err := user.Current()
	if err != nil {
		Logger.Fatal(err)
	}
	return currentUser.HomeDir
}

// GetUserConfig returns a Config object containing the user's configuration
func GetUserConfig() Config {
	Logger.Print("Searching for user's configuration")
	config := Config{}
	if IsConfigured() {
		y, err := ioutil.ReadFile(configurationFile)
		if err != nil {
			Logger.Fatal(err)
		}
		err = yaml.Unmarshal(y, &config)
		if err != nil {
			Logger.Fatal(err)
		}
		Logger.Print("Found user's already existing configuration")
		return config
	} else {
		Logger.Print("No configuration found")
		EnsureUserWantsNewConfig()
		Logger.Fatal("Restart the program in order to apply this configuration.")
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
		Logger.Fatal(err)
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
	serverIP := getServerIP()

	generateConfig(base, startingPrime, maxFilesize, maxBufferSize, showFails, serverIP)
	fmt.Println("Your configuration has now been generated.")
}

// getBaseDirectory returns the user's preference for a base directory
func getBaseDirectory() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Base directory (default: %s/.primes/): ", home)
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		Logger.Fatal(err)
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
		Logger.Fatal(err)
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
		Logger.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxFilesize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			Logger.Fatal(err)
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
		Logger.Fatal(err)
	}

	userChoiceString = strings.Trim(userChoiceString, " \n")
	if userChoiceString == "" {
		userChoice = defaultMaxBufferSize
	} else {
		userChoice, err = strconv.Atoi(userChoiceString)
		if err != nil {
			Logger.Fatal(err)
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
		Logger.Fatal(err)
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

// getserverIP returns the user's preference for the ip to
// connect to as the server
func getServerIP() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Address to connect to as server (default: 192.168.1.66): ")
	userChoice, err := reader.ReadString('\n')
	if err != nil {
		Logger.Fatal(err)
	}
	userChoice = strings.Trim(userChoice, " \n")
	if userChoice == "" {
		userChoice = defaultServerIP
	}
	return userChoice
}

// generateConfig concatinates the user's preferences into YAML format
func generateConfig(base string, startingPrime string, maxFilesize int, maxBufferSize int, showFails bool, serverIP string) {
	config, err := os.Create(home + "/.primegenerator.yaml")
	defer config.Close()
	if err != nil {
		Logger.Fatal(err)
	}
	c := Config{base, startingPrime, maxFilesize, maxBufferSize, showFails, serverIP}
	yaml, err := yaml.Marshal(c)
	if err != nil {
		Logger.Fatal(err)
	}
	config.Write(yaml)
}
