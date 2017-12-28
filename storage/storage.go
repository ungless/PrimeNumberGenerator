package storage

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
)

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

// isNewFileNeeded() checks wether a new file is needed by asserting that
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
