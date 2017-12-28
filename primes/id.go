package primes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"
)

// Round() is used to round numbers to the nearest x
func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}

// GetCurrentId returns the current id, rounded to nearest hundred
func GetCurrentId() uint64 {
	maximumId := GetTotalPrimeCount()
	currentId := uint64(Round(float64(maximumId), float64(config.MaxBufferSize)))
	return currentId
}

// GetTotalPrimeCount finds the number of lines in each file
func GetTotalPrimeCount() uint64 {
	var maximumId uint64
	openDirectory := storage.OpenDirectory(os.O_RDONLY, 0600)
	defer openDirectory.Close()
	scanner := bufio.NewScanner(openDirectory)
	for scanner.Scan() {
		filename := scanner.Text()
		file, err := os.Open(storage.FormatFilePath(filename))
		if err != nil {
			break
		}

		r := bufio.NewReader(file)
		linesInFile, err := getLinesInFile(r)
		if err != nil {
			config.Logger.Fatal(err)
		}
		maximumId += uint64(linesInFile)
	}
	return maximumId
}

// getLinesInFile counts the lines of a given file
func getLinesInFile(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// ShowCurrentCount displays the n number of primes calculated
func ShowCurrentCount() {
	fmt.Printf("Total (to the nearest hundred) prime numbers calculated and stored: #%d\n", config.Id)
}
