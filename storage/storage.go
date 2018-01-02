package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
)

var mu sync.Mutex

type BigIntSlice []*big.Int

func (s BigIntSlice) Len() int           { return len(s) }
func (s BigIntSlice) Less(i, j int) bool { return s[i].Cmp(s[j]) < 0 }
func (s BigIntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// convertPrimesToWritableFormat() takes a buffer of primes and converts them to a string
// with each prime separated by a newline
func convertPrimesToWritableFormat(buffer []*big.Int) string {
	var formattedBuffer bytes.Buffer
	for _, prime := range buffer {
		formattedBuffer.WriteString(prime.String() + "\n")
	}
	return formattedBuffer.String()
}

// FormatFilePath formats inputted filename to create a proper file path.
func FormatFilePath(filename string) string {
	return config.Base + filename + ".txt"
}

// createPrimesBase makes the base directory
func createPrimesBase() {
	log.Print("Creating base directory")
	os.Mkdir(config.Base, os.ModePerm)
}

// createDirectory creates the directory.txt file as defined
// in settings.go
func createDirectory() {
	_, err := os.Create(config.Directory)
	if err != nil {
		createPrimesBase()
		_, err := os.Create(config.Directory)
		if err != nil {
			panic(err)
		}
	}
}

// OpenDirectory returns an open os.File of the directory.txt
// as defined in settings.go
func OpenDirectory(flag int, perm os.FileMode) *os.File {
	openDirectory, err := os.OpenFile(config.Directory, flag, perm)
	if err != nil {
		createDirectory()
		openedCreatedDirectory, err := os.OpenFile(config.Directory, flag, perm)
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
	directory := OpenDirectory(os.O_RDONLY, 0600)
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

// isNewFileNeeded() checks whether a new file is needed by asserting that
// the id is divisible by maxFilesize - as defined in settings.go
func isNewFileNeeded(id uint64) bool {
	modulusIdAndMaxFilesize := big.NewInt(0).Mod(big.NewInt(int64(id)), big.NewInt(int64(config.MaxFilesize)))
	divisibleByMaxFilesize := modulusIdAndMaxFilesize.Int64() == 0
	return divisibleByMaxFilesize
}

// openLatestFile() returns an open os.File of the latest written to file
func OpenLatestFile(flag int, perm os.FileMode) *os.File {
	lastFileWritten := getLastFileWritten()
	file, err := os.OpenFile(FormatFilePath(lastFileWritten), flag, perm)
	newFileNeeded := isNewFileNeeded(config.Id)
	if err != nil || newFileNeeded {
		newFileName := getNewFileName(config.Id)
		createNextFile(newFileName)
		createdNextFile, err := os.OpenFile(FormatFilePath(newFileName), flag, perm)
		if err != nil {
			panic(err)
		}
		return createdNextFile
	}
	return file
}

// getNextFileName() generates the name of the possible file
func getNewFileName(id uint64) string {
	nextFile := fmt.Sprintf("%d-%d", id, id+uint64(config.MaxFilesize))
	return nextFile
}

// createNextFile() creates the next file to be written to
// and writes its name to the directory
func createNextFile(newFileName string) {
	directory := OpenDirectory(os.O_APPEND|os.O_WRONLY, 0600)
	defer directory.Close()
	directory.WriteString(newFileName + "\n")
	log.Print("Creating next file. ", newFileName)
	_, err := os.Create(FormatFilePath(newFileName))
	if err != nil {
		panic(err)
	}
}

// FlushBufferToFile() takes a buffer of primes and flushes them to the latest file
func FlushBufferToFile(buffer BigIntSlice) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Println("Writing buffer....")
	sort.Sort(buffer)
	atomic.AddUint64(&config.Id, uint64(config.MaxBufferSize))

	file := OpenLatestFile(os.O_APPEND|os.O_WRONLY, 0600)
	defer file.Close()
	readableBuffer := convertPrimesToWritableFormat(buffer)

	file.WriteString(readableBuffer)
	fmt.Println("Finished writing buffer.")
}
