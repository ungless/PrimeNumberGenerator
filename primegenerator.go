// Main package for primegenerator. Generates primes.
package main

import (
	"bufio"
	"fmt"
	// "io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"
)

var (
	count = get_total_count()
)

// format_filename formats inputted filename to create a proper file path.
func format_filename(filename string) string {
	return base + filename + ".txt"
}

// check_prime checks whether number is a prime.
func check_prime(number *big.Int) bool {
	return number.ProbablyPrime(1)
}

// display_prime_pretty displays successful prime generations nicely.
func display_prime_pretty(number *big.Int, start time.Time) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[4;30;42mSuccess\x1b[0m\t%s\t\x1b[1;37;37m#%d\x1b[0m\n",
		number,
		time.Now().Sub(start),
		count,
	)
}

// display_fail_pretty displays failed prime generations nicely.
func display_fail_pretty(number *big.Int, start time.Time) {
	fmt.Printf("\033[1;93mTesting \033[0m\033[1;32m%s\033[0m\t\x1b[2;1;41mFail\x1b[0m\t%s\t\x1b[1;37;37m#%d\x1b[0m\n",
		number,
		time.Now().Sub(start),
		count,
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

// open_latest_file returns an open os.File of the latest written to file
func open_latest_file(flag int, perm os.FileMode) *os.File {
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

	file, err := os.OpenFile(format_filename(latest_file), flag, perm)
	if err != nil {
		create_next_file()
		next_filename := get_next_file_name()
		created_next_file, err := os.OpenFile(format_filename(next_filename), flag, perm)
		if err != nil {
			panic(err)
		}
		return created_next_file
	}
	return file
}

// get_next_file_name generates the name of the possible file
func get_next_file_name() string {
	next_file := fmt.Sprintf("%s-%s", count, big.NewInt(0).Add(count, big.NewInt(max_filesize)))
	return next_file
}

// create_next_file creates the next file to be written to
// and writes its name to the directory
func create_next_file() {
	directory := open_directory(os.O_APPEND|os.O_WRONLY, 0600)
	defer directory.Close()

	next_file_name := get_next_file_name()
	directory.WriteString(next_file_name + "\n")
	fmt.Println("Creating next file.", next_file_name)

	_, err := os.Create(format_filename(next_file_name))
	if err != nil {
		panic(err)
	}
}

// new_file_needed returns a boolean of whether a new file is
// required or not based on variables in settings.go
func new_file_needed() bool {
	divisible_by_max_filesize := big.NewInt(0).Mod(count, big.NewInt(max_filesize)).Int64() == 0
	return divisible_by_max_filesize
}

// write_prime writes a number with an appropriate newline
// to the current working file
func write_prime(number *big.Int) {
	writing := fmt.Sprintf("\n%d", number)
	if new_file_needed() == true {
		create_next_file()
	}
	file := open_latest_file(os.O_APPEND|os.O_WRONLY, 0600)
	defer file.Close()
	file.WriteString(writing)
}

func main() {
	fmt.Println("Welcome to the Prime Number Generator.")
	last_prime := get_last_prime()
	for i := last_prime; true; i.Add(i, big.NewInt(2)) {
		start := time.Now()
		if check_prime(i) {
			count.Add(count, big.NewInt(1))
			display_prime_pretty(i, start)
			write_prime(i)
		} else if show_fails == true {
			display_fail_pretty(i, start)
		}
	}
}
