package client

import (
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
func getNextComputationToPerform() (computation.Computation, error) {
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
			nextComputation, err := getNextComputationToPerform()
			if err != nil {
				log.Print("Retrying connection")
				continue
			}
			computationsToPerform <- nextComputation
		}
	}()

	for c := range computationsToPerform {
		// Check its primatlity, then feed into channel if successful
		i := c.Prime.Value
		go func(i *big.Int) {
			start := time.Now()
			isPrime := primes.CheckPrimality(i)
			if isPrime == true {
				computationPrime := primes.Prime{
					TimeTaken: time.Now().Sub(start),
					Value:     i,
					Id:        config.Id,
				}
				validComputations <- computation.Computation{computationPrime, c.Divisor, c.Hash}
			} else {
				computationPrime := primes.Prime{
					TimeTaken: time.Now().Sub(start),
					Value:     i,
				}
				invalidComputations <- computation.Computation{computationPrime, c.Divisor, c.Hash}
			}
		}(i)
		fmt.Println(c)
	}
}
