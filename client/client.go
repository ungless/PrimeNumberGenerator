package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	app "github.com/urfave/cli"
)

// sendComputationResult sends a JSON string through POST to the server
// of the results of a computation
func sendComputationResult(c computation.Computation) {
	url := "http://localhost:8080/finished"

	json, err := computation.GetJSONFromComputation(c)
	if err != nil {
		config.Logger.Fatal(err)
	}
	config.Logger.Print("Sending JSON: ", string(json))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

// getUnMarshalledComputation produces a computation from a JSON string
func getUnMarshalledComputation(body string) computation.Computation {
	var c computation.Computation
	err := json.Unmarshal([]byte(body), &c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// getNextComputation returns a computation hash given by
// the server
func fetchNextComputationToPerform() (computation.Computation, error) {
	log.Print("Requesting next computation")
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		log.Print("Cannot connect to server")
		return computation.Computation{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Print("Found computation")
	computation := getUnMarshalledComputation(string(body))
	return computation, nil
}

// LaunchClient launches the client application, and manages
// goroutines
func LaunchClient(c *app.Context) {
	computationsToPerform := make(chan computation.Computation, 10)
	validComputations := make(chan computation.Computation, 10)
	invalidComputations := make(chan computation.Computation, 10)
	go func() {
		for {
			nextComputation, err := fetchNextComputationToPerform()
			if err != nil {
				time.Sleep(1 * time.Second)
				log.Print("Retrying connection")
				continue
			}
			fmt.Println("Passing computation to computationsToPerform")
			computationsToPerform <- nextComputation
		}
	}()

	go func() {
		for c := range validComputations {
			config.Logger.Printf("%s / %s valid.", c.Prime.Value, c.Divisor)
			sendComputationResult(c)
			config.Logger.Fatal()
		}
	}()

	go func() {
		for c := range invalidComputations {
			config.Logger.Printf("%s / %s invalid.", c.Prime.Value, c.Divisor)
			sendComputationResult(c)
		}
	}()

	for c := range computationsToPerform {
		i := c.Prime.Value
		go func(i *big.Int, c computation.Computation) {
			start := time.Now()
			fmt.Println("Computing")
			computationIsValid := computation.RunDistributedComputation(c)
			fmt.Println("Finished computation:", computationIsValid)
			duration := time.Now().Sub(start)
			newPrimeDuration := c.Prime.TimeTaken + duration
			if computationIsValid == true {
				computationPrime := primes.Prime{
					TimeTaken: newPrimeDuration,
					Value:     i,
					Id:        config.Id,
				}
				validComputations <- computation.Computation{
					Prime:     computationPrime,
					Divisor:   c.Divisor,
					IsValid:   computationIsValid,
					TimeTaken: duration,
					Hash:      c.Hash,
				}
			} else {
				computationPrime := primes.Prime{
					TimeTaken: newPrimeDuration,
					Value:     i,
				}
				invalidComputations <- computation.Computation{
					Prime:     computationPrime,
					Divisor:   c.Divisor,
					IsValid:   computationIsValid,
					TimeTaken: duration,
					Hash:      c.Hash,
				}
			}
		}(i, c)
	}
}
