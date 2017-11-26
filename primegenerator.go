// Copyright (C) 2017 by Max Ungless
// Main package for primegenerator. Generates primes.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/big"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var (
	globalCount        = big.NewInt(0)
	id          uint64 = uint64(Round(float64(GetMaximumId()), float64(maxBufferSize)))
	mu          sync.Mutex
)

type prime struct {
	id        uint64
	value     *big.Int
	timeTaken time.Duration
}

type bigIntSlice []*big.Int

func (s bigIntSlice) Len() int           { return len(s) }
func (s bigIntSlice) Less(i, j int) bool { return s[i].Cmp(s[j]) < 0 }
func (s bigIntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// formatFilePath formats inputted filename to create a proper file path.
func formatFilePath(filename string) string {
	return base + filename + ".txt"
}

// checkPrimality checks whether number is a prime.
func checkPrimality(number *big.Int) bool {
	return number.ProbablyPrime(1)
}

// displayPrimePretty displays successful prime generations nicely.
func displayPrimePretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[4;30;42mSuccess\x1b[0m\t%s\x1b[0m\n",
		number,
		timeTaken,
	)
}

// displayFailPretty displays failed prime generations nicely.
func displayFailPretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[2;1;41mFail\x1b[0m\t%s\t\x1b[0m\n",
		number,
		timeTaken,
	)
}

// showHelp shows help to the user.
func showHelp() {
	fmt.Println("\nCOMMANDS")
	fmt.Println("count \t Displays the total number of generated primes.")
	fmt.Println("configure \t Generates a configuration for the program.")
	fmt.Println("run \t Runs the program indefinitely.")
	fmt.Println("help \t Displays this screen. Gives help.")
}

// showProgramDetails prints details about the program to STDOUT
func showProgramDetails() {
	fmt.Printf("PrimeNumberGenerator %s", version)
	fmt.Println("\nCopyright (C) 2017 by Max Ungless")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.\nThis is free software, and you are welcome to redistribute it\nunder the condiditions set in the GNU General Public License version 3. See the file named LICENSE for details.")
	fmt.Println("\nFor bugs, send mail to max@maxungless.com")
	fmt.Println()
}

// GetMaximumId retrieves the total prime count from previous runs.
func GetMaximumId() uint64 {
	var maximumId uint64

	openDirectory := OpenDirectory(os.O_RDONLY, 0600)
	defer openDirectory.Close()
	scanner := bufio.NewScanner(openDirectory)

	for scanner.Scan() {
		filename := scanner.Text()
		file, err := os.Open(formatFilePath(filename))
		if err != nil {
			break
		}

		fileScanner := bufio.NewScanner(file)
		for fileScanner.Scan() {
			maximumId += 1
		}
		file.Close()
	}
	return maximumId
}

// getLastPrime() searches for last generated prime
// in all prime storage files.
func getLastPrime() *big.Int {
	latestFile := OpenLatestFile(os.O_RDONLY, 0666)
	defer latestFile.Close()

	var lastPrimeGenerated string
	scanner := bufio.NewScanner(latestFile)
	for scanner.Scan() {
		lastPrimeGenerated = scanner.Text()
	}

	if lastPrimeGenerated == "0" || lastPrimeGenerated == "" {
		lastPrimeGenerated = startingPrime
	}
	foundPrime := new(big.Int)
	foundPrime.SetString(lastPrimeGenerated, 10)
	return foundPrime
}

// Round() is used to round numbers to the nearest x
func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

// convertPrimesToWritableFormat() takes a buffer of primes and converts them to a string
// with each prime separated by a newline
func convertPrimesToWritableFormat(buffer []*big.Int) string {
	var formattedBuffer bytes.Buffer
	for _, prime := range buffer {
		formattedBuffer.WriteString(prime.String() + "\n")
	}
	return formattedBuffer.String()
}

// FlushBufferToFile() takes a buffer of primes and flushes them to the latest file
func FlushBufferToFile(buffer bigIntSlice) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Println("Writing buffer....")
	sort.Sort(buffer)
	atomic.AddUint64(&id, uint64(maxBufferSize))

	file := OpenLatestFile(os.O_APPEND|os.O_WRONLY, 0600)
	defer file.Close()
	readableBuffer := convertPrimesToWritableFormat(buffer)

	file.WriteString(readableBuffer)
	fmt.Println("Finished writing buffer.")
}

// ComputePrimes computes primes concurrently until KeyboardInterrupt
func ComputePrimes(lastPrime *big.Int, writeToFile bool, toInfinity bool, maxNumber *big.Int) {
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan prime, 100)
	invalidPrimes := make(chan prime, 100)
	var primeBuffer bigIntSlice

	go func() {
		if toInfinity {
			fmt.Println(lastPrime)
			for i := lastPrime; true; i.Add(i, big.NewInt(2)) {
				numberToTest := big.NewInt(0).Set(i)
				numbersToCheck <- numberToTest
			}
		} else {
			for i := lastPrime; i.Cmp(maxNumber) == -1; i.Add(i, big.NewInt(2)) {
				numberToTest := big.NewInt(0).Set(i)
				numbersToCheck <- numberToTest
			}
		}
	}()

	go func() {
		for elem := range validPrimes {
			primeBuffer = append(primeBuffer, elem.value)
			if len(primeBuffer) == maxBufferSize {
				if writeToFile {
					FlushBufferToFile(primeBuffer)
				}
				primeBuffer = nil
			}
			displayPrimePretty(elem.value, elem.timeTaken)
		}
	}()

	go func() {
		for elem := range invalidPrimes {
			if showFails == true {
				displayFailPretty(elem.value, elem.timeTaken)
			}
		}
	}()

	for i := range numbersToCheck {
		go func(i *big.Int) {
			start := time.Now()
			isPrime := checkPrimality(i)
			if isPrime == true {
				validPrimes <- prime{
					timeTaken: time.Now().Sub(start),
					value:     i,
					id:        id,
				}
			} else {
				invalidPrimes <- prime{
					timeTaken: time.Now().Sub(start),
					value:     i,
				}
			}
		}(i)
	}
}

func main() {
	showProgramDetails()
	arguments := os.Args
	if len(arguments) == 2 {
		switch arguments[1] {
		case "count":
			ShowCurrentCount()
		case "run":
			ComputePrimes(getLastPrime(), true, true, big.NewInt(0))
		case "help":
			showHelp()
		case "configure":
			RunConfigurator()
		default:
			fmt.Println("Please specify a valid command.")
			showHelp()
		}
	} else if len(arguments) == 1 {
		showHelp()
	}
}
