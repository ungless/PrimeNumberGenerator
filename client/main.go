package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	app "github.com/urfave/cli"
)

// getNextComputation returns a computation hash given by
// the server
func getNextComputation() (string, error) {
	log.Print("Requesting next computation")
	resp, err := http.Get("http://localhost:8080/test")
	if err != nil {
		log.Print("Cannot connect to server")
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("Found computation %s", body)
	return string(body), nil
}

// LaunchClient launches the client application, and manages
// goroutines
func LaunchClient(c *app.Context) {
	computationsToPerform := make(chan string, 10)
	go func() {
		for {
			time.Sleep(3 * time.Second)
			nextComputation, err := getNextComputation()
			if err != nil {
				log.Print("Retrying connection")
				continue
			}
			computationsToPerform <- nextComputation
		}
	}()

	for computation := range computationsToPerform {
		fmt.Println(computation)
	}
}
