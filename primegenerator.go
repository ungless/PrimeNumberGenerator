// Main package for primegenerator. Generates primes.
package main

import (

	"bufio"
	"bytes"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	globalCount        = big.NewInt(0)
	id          uint64 = uint64(Round(float64(getMaximumId()), maxBufferSize))
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
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[2;1;41mFail\x1b[0m\t%s\t\x1b\n",
		number,
		timeTaken,
	)
}

// getMaximumId retrieves the total prime count from previous runs.
func getMaximumId() uint64 {
	var maximumId uint64

	openDirectory := openDirectory(os.O_RDONLY, 0600)
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
	fmt.Println(maximumId)
	return maximumId
}

// getLastPrime() searches for last generated prime
// in all prime storage files.
func getLastPrime() *big.Int {
	latestFile := openLatestFile(os.O_RDONLY, 0666)
	defer latestFile.Close()

	var lastPrimeGenerated string
	scanner := bufio.NewScanner(latestFile)
	for scanner.Scan() {
		lastPrimeGenerated = scanner.Text()
	}

	lastPrimeAsInt, err := strconv.Atoi(lastPrimeGenerated)
	if err != nil {
		lastPrimeAsInt = startingPrime
	}
	fmt.Println(lastPrimeAsInt)
	return big.NewInt(int64(lastPrimeAsInt))
}

// Round() is used to round numbers to the nearest x
func Round(x, unit float64) float64 {
    return float64(int64(x/unit+0.5)) * unit
}

// createDirectory creates the directory.txt file as defined
// in settings.go
func createDirectory() {
	_, err := os.Create(directory)
	if err != nil {
		panic(err)
	}
}

// openDirectory returns an open os.File of the directory.txt
// as defined in settings.nnngo
func openDirectory(flag int, perm os.FileMode) *os.File {
	openDirectory, err := os.OpenFile(directory, flag, perm)
	if err != nil {
		createDirectory()
		openedCreatedDirectory, err := os.OpenFile(directory, flag, perm)
		if err != nil {
			panic(err)
		}
		return openedCreatedDirectory
	}
	return openDirectory
}

// getLastFileWritten() searches the directory for the final line,
// and returns it.
func getLastFileWritten() string {
	directory := openDirectory(os.O_RDONLY, 0600)
	defer directory.Close()

	var latestFile string
	scanner := bufio.NewScanner(directory)
	for scanner.Scan() {
		scannedText := scanner.Text()
		if scannedText == "" {
			break
		}
		latestFile = scanner.Text()
	}
	return latestFile
}

// isNewFileNeeded() checks wether a new file is needed by asserting that
// the id is divisible by maxFilesize - as defined in settings.go
func isNewFileNeeded(id uint64) bool {
	divisibleByMaxFilesize := big.NewInt(0).Mod(big.NewInt(int64(id)), big.NewInt(maxFilesize)).Int64() == 0
	return divisibleByMaxFilesize
}

// openLatestFile() returns an open os.File of the latest written to file
func openLatestFile(flag int, perm os.FileMode) *os.File {
	lastFileWritten := getLastFileWritten()
	file, err := os.OpenFile(formatFilePath(lastFileWritten), flag, perm)
	newFileNeeded := isNewFileNeeded(id)
	if err != nil || newFileNeeded {
		newFileName := getNewFileName(id)
		createNextFile(newFileName)
		createdNextFile, err := os.OpenFile(formatFilePath(newFileName), flag, perm)
		if err != nil {
			panic(err)
		}
		return createdNextFile
	}
	return file
}

// getNextFileName() generates the name of the possible file
func getNewFileName(id uint64) string {
	nextFile := fmt.Sprintf("%d-%d", id, id+maxFilesize)
	return nextFile
}

// createNextFile() creates the next file to be written to
// and writes its name to the directory
func createNextFile(newFileName string) {
	directory := openDirectory(os.O_APPEND|os.O_WRONLY, 0600)
	defer directory.Close()

	directory.WriteString(newFileName + "\n")
	fmt.Println("Creating next file.", newFileName)

	_, err := os.Create(formatFilePath(newFileName))
	if err != nil {
		panic(err)
	}
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
	fmt.Println(buffer)
	atomic.AddUint64(&id, maxBufferSize)
	fmt.Println(id)
	
	file := openLatestFile(os.O_APPEND|os.O_WRONLY, 0600)
	defer file.Close()
	readableBuffer := convertPrimesToWritableFormat(buffer)

	file.WriteString(readableBuffer)
	fmt.Println("Finished writing buffer.")
}

func main() {
	fmt.Println("Welcome to the Prime Number Generator.")
	lastPrime := getLastPrime()
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan prime, 100)
	invalidPrimes := make(chan prime, 100)
	var primeBuffer bigIntSlice

	go func() {
		for i := lastPrime; true; i.Add(i, big.NewInt(2)) {
			numberToTest := big.NewInt(0).Set(i)
			numbersToCheck <- numberToTest
		}
	}()

	go func() {
		for elem := range validPrimes {
			primeBuffer = append(primeBuffer, elem.value)
			if len(primeBuffer) == maxBufferSize {
				FlushBufferToFile(primeBuffer)
				primeBuffer = nil
				// os.Exit(1)
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
