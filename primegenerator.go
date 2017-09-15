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
	id          uint64 = 0
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

// format_filename formats inputted filename to create a proper file path.
func format_filename(filename string) string {
	return base + filename + ".txt"
}

// check_prime checks whether number is a prime.
func check_prime(number *big.Int) bool {
	return number.ProbablyPrime(1)
}

// display_prime_pretty displays successful prime generations nicely.
func display_prime_pretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[4;30;42mSuccess\x1b[0m\t%s\x1b[0m\n",
		number,
		timeTaken,
	)
}

// display_fail_pretty displays failed prime generations nicely.
func display_fail_pretty(number *big.Int, timeTaken time.Duration) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[2;1;41mFail\x1b[0m\t%s\t\x1b\n",
		number,
		timeTaken,
	)
}

// get_total_count retrieves the total prime count from previous runs.
func get_total_count() *big.Int {
	total_count := big.NewInt(0)

	open_directory := open_directory(os.O_RDONLY, 0600)
	defer open_directory.Close()
	scanner := bufio.NewScanner(open_directory)

	for scanner.Scan() {
		filename := scanner.Text()
		file, err := os.Open(format_filename(filename))
		if err != nil {
			break
		}

		file_scanner := bufio.NewScanner(file)
		for file_scanner.Scan() {
			total_count.Add(total_count, big.NewInt(1))
		}
		file.Close()
	}
	return total_count
}

// get_last_prime() searches for last generated prime
// in all prime storage files.
func get_last_prime() *big.Int {
	latest_file := open_latest_file(os.O_RDONLY, 0666)
	defer latest_file.Close()

	var last_prime string
	scanner := bufio.NewScanner(latest_file)
	for scanner.Scan() {
		last_prime = scanner.Text()
	}

	last_prime_as_int, err := strconv.Atoi(last_prime)
	if err != nil {
		last_prime_as_int = starting_prime
	}
	return big.NewInt(int64(last_prime_as_int))
}

// create_directory creates the directory.txt file as defined
// in settings.go
func create_directory() {
	_, err := os.Create(directory)
	if err != nil {
		panic(err)
	}
}

// open_directory returns an open os.File of the directory.txt
// as defined in settings.go
func open_directory(flag int, perm os.FileMode) *os.File {
	open_directory, err := os.OpenFile(directory, flag, perm)
	if err != nil {
		create_directory()
		opened_created_directory, err := os.OpenFile(directory, flag, perm)
		if err != nil {
			panic(err)
		}
		return opened_created_directory
	}
	return open_directory
}

func getLastFileWritten() string {
	directory := open_directory(os.O_RDONLY, 0600)
	defer directory.Close()
	
	var latest_file string
	scanner := bufio.NewScanner(directory)
	for scanner.Scan() {
		scanned_text := scanner.Text()
		if scanned_text == "" {
			break
		}
		latest_file = scanner.Text()
	}
	return latest_file
}

func isNewFileNeeded(id uint64) bool {
	divisibleByMaxFilesize := big.NewInt(0).Mod(big.NewInt(int64(id)), big.NewInt(max_filesize)).Int64() == 0 
	return divisibleByMaxFilesize
}

// open_latest_file returns an open os.File of the latest written to file
func open_latest_file(flag int, perm os.FileMode) *os.File {
	lastFileWritten := getLastFileWritten()
	file, err := os.OpenFile(format_filename(lastFileWritten), flag, perm)
	if err != nil || isNewFileNeeded(id) {
		fmt.Println("new file")

		newFileName := getNewFileName(id)
		createNextFile(newFileName)
		created_next_file, err := os.OpenFile(format_filename(newFileName), flag, perm)
		if err != nil {
			panic(err)
		}
		return created_next_file
	}
	return file
}

// get_next_file_name generates the name of the possible file
func getNewFileName(id uint64) string {
	next_file := fmt.Sprintf("%d-%d", id, id+max_filesize)
	return next_file
}

// create_next_file creates the next file to be written to
// and writes its name to the directory
func createNextFile(newFileName string) {
	directory := open_directory(os.O_APPEND|os.O_WRONLY, 0600)
	defer directory.Close()

	directory.WriteString(newFileName + "\n")
	fmt.Println("Creating next file.", newFileName)

	_, err := os.Create(format_filename(newFileName))
	if err != nil {
		panic(err)
	}
}

func ConvertPrimesToWritableFormat(buffer []*big.Int) string {
	var formattedBuffer bytes.Buffer
	for _, prime := range buffer {
		formattedBuffer.WriteString(prime.String() + "\n")
	}
	return formattedBuffer.String()
}

func FlushBufferToFile(buffer bigIntSlice) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Println("Writing buffer....")
	sort.Sort(buffer)
	fmt.Println(buffer)

	file := open_latest_file(os.O_APPEND|os.O_WRONLY, 0600)
	defer file.Close()
	readableBuffer := ConvertPrimesToWritableFormat(buffer)
	
	file.WriteString(readableBuffer)
	fmt.Println("Finished writing buffer.")
}

func main() {
	fmt.Println("Welcome to the Prime Number Generator.")
	// last_prime := get_last_prime()
	last_prime := big.NewInt(1)
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan prime, 100)
	invalidPrimes := make(chan prime, 100)
	var primeBuffer bigIntSlice

	go func() {
		for i := last_prime; true; i.Add(i, big.NewInt(2)) {
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
			display_prime_pretty(elem.value, elem.timeTaken)
		}
	}()

	go func() {
		for elem := range invalidPrimes {
			if show_fails == true {
				display_fail_pretty(elem.value, elem.timeTaken)
			}
		}
	}()

	for i := range numbersToCheck {
		go func(i *big.Int) {
			start := time.Now()
			is_prime := check_prime(i)
			if is_prime == true {
				atomic.AddUint64(&id, 1)
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
