// Copyright (C) 2017 by Max Ungless
// Main package for primegenerator. Generates primes.
package main

import (
	"bufio"
	"fmt"
	//	"io/ioutil"
	"math/big"
	"os"

	"github.com/MaxTheMonster/PrimeNumberGenerator/client"
	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/server"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"

	"github.com/urfave/cli"
)

const (
	version  = "1.1.3"
	appName  = "PrimeNumberGenerator"
	appUsage = "Generate prime numbers forever"

	descConfigure = "Runs auto-configuration wizard"
	descCount     = "Displays the estimated curren n prime numbers"
	descRun       = "Begins computation of primes"
	descClient    = "Launches a new instance of a client"
	descServer    = "Launches a new instance of a server"

	appHelpTemplate = `{{if .VisibleCommands}}COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`
)

// SetConfiguration sets the global configuration variables
func SetConfiguration() {
	config.LocalConfig = config.GetUserConfig()
	config.StartingPrime = config.LocalConfig.StartingPrime
	config.MaxFilesize = config.LocalConfig.MaxFilesize
	config.MaxBufferSize = config.LocalConfig.MaxBufferSize
	config.ShowFails = config.LocalConfig.ShowFails
	config.Host = config.LocalConfig.ServerIP
}

// SetId sets the gloabl id variable
func SetId() {
	config.Id = primes.GetCurrentId()
}

// SetLastPrimeGenerated sets the global lastprimegenerated variable
func SetLastPrimeGenerated() {
	config.LastPrimeGenerated = getLastPrime()
}

// showHelp shows help to the user.
func showHelp() {
	fmt.Println("COMMANDS")
	fmt.Println("count \t Displays the total number of generated primes.")
	fmt.Println("configure \t Generates a configuration for the program.")
	fmt.Println("run \t Runs the program indefinitely.")
	fmt.Println("version \t Displays the version of the program.")
	fmt.Println("help \t Displays this screen. Gives help.")

	os.Exit(1)
}

// showProgramDetails prints details about the program to STDOUT
func showProgramDetails() {
	fmt.Printf("PrimeNumberGenerator %s LITE", version)
	fmt.Println("\nCopyright (C) 2017-2018 by Max Ungless")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.\nThis is free software, and you are welcome to redistribute it\nunder the condiditions set in the GNU General Public License version 3.\nSee the file named LICENSE for details.")
	fmt.Println("\nFor bugs, send mail to max@maxungless.com")
	fmt.Println()
}

// getLastPrime() searches for last generated prime
// in all prime storage files.
func getLastPrime() *big.Int {
	latestFile := storage.OpenLatestFile(os.O_RDONLY, 0666)
	defer latestFile.Close()
	var lastPrimeGenerated string
	scanner := bufio.NewScanner(latestFile)
	for scanner.Scan() {
		lastPrimeGenerated = scanner.Text()
	}
	if lastPrimeGenerated == "0" || lastPrimeGenerated == "" {
		lastPrimeGenerated = config.StartingPrime
	}
	foundPrime := new(big.Int)
	foundPrime.SetString(lastPrimeGenerated, 10)
	return foundPrime
}

func init() {
	//	config.Logger.SetOutput(ioutil.Discard)
	showProgramDetails()
	SetConfiguration()
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Version = version
	cli.AppHelpTemplate = appHelpTemplate

	app.Commands = []cli.Command{
		{
			Name:    "configure",
			Aliases: []string{"cn"},
			Usage:   descConfigure,
			Action: func(c *cli.Context) error {
				config.RunConfigurator()
				return nil
			},
		},
		{
			Name:    "count",
			Aliases: []string{"ct"},
			Usage:   descCount,
			Before: func(c *cli.Context) error {
				SetId()
				return nil
			},
			Action: func(c *cli.Context) error {
				primes.ShowCurrentCount()
				return nil
			},
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   descRun,
			Before: func(c *cli.Context) error {
				SetId()
				SetLastPrimeGenerated()
				return nil
			},
			Action: func(c *cli.Context) error {
				computation.ComputePrimes(config.LastPrimeGenerated, true, true, big.NewInt(0))
				return nil
			},
		},
		{
			Name:    "client",
			Aliases: []string{"cl"},
			Usage:   descClient,
			Action:  client.LaunchClient,
		},
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   descServer,
			Before: func(c *cli.Context) error {
				SetId()
				SetLastPrimeGenerated()
				return nil
			},
			Action: server.LaunchServer,
		},
	}
	app.Run(os.Args)
}
