package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"

	//	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	app "github.com/urfave/cli"
)

// sendComputationResult sends a JSON string through POST to the server
// of the results of a computation
func sendPrimeResult(p primes.Prime) {
	url := "http://" + config.Address + config.ReturnPoint
	json, err := json.Marshal(p)
	if err != nil {
		config.Logger.Fatal(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Fatal(err)
	}
	defer resp.Body.Close()
}

// getUnMarshalledPrime produces a computation from a JSON string
func getUnMarshalledPrime(body string) primes.Prime {
	var c primes.Prime
	err := json.Unmarshal([]byte(body), &c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// fetchNextPrimeToPerform returns a computation hash given by
// the server
func fetchNextPrimeToPerform() (primes.Prime, error) {
	url := "http://" + config.Address + config.AssignmentPoint
	resp, err := http.Get(url)
	if err != nil {
		log.Print("Cannot connect to server")
		return primes.Prime{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	computation := getUnMarshalledPrime(string(body))
	return computation, nil
}

// LaunchClient launches the client application, and manages
// goroutines
func LaunchClient(c *app.Context) {
	primesToCompute := make(chan primes.Prime, 10)
	validPrimes := make(chan primes.Prime, 10)
	invalidPrimes := make(chan primes.Prime, 10)
	go func() {
		for {
			nextPrime, err := fetchNextPrimeToPerform()
			if err != nil {
				time.Sleep(1 * time.Second)
				log.Print("Retrying connection")
				continue
			}
			primesToCompute <- nextPrime
		}
	}()

	go func() {
		for p := range validPrimes {
			primes.DisplayPrimePretty(p.Value, p.TimeTaken)
			sendPrimeResult(p)
		}
	}()

	go func() {
		for p := range invalidPrimes {
			primes.DisplayFailPretty(p.Value, p.TimeTaken)
			sendPrimeResult(p)
		}
	}()

	for p := range primesToCompute {
		i := p.Value
		go func(i *big.Int, p primes.Prime) {
			start := time.Now()
			p.IsValid = primes.CheckPrimality(i)
			duration := time.Now().Sub(start)
			p.TimeTaken = duration
			if p.IsValid == true {
				validPrimes <- p
			} else {
				invalidPrimes <- p
			}
		}(i, p)
	}
}
