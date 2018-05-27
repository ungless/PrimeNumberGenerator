package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	app "github.com/urfave/cli"
)

var lock sync.Mutex

// sendPrimeResult sends a JSON string through POST to the server
// of the results of a computation
func sendPrimeResult(p primes.Prime) error {
	lock.Lock()
	defer lock.Unlock()
	url := "http://" + config.Address + config.ReturnPoint
	log.Print(url)
	json, err := json.Marshal(p)
	if err != nil {
		config.Logger.Print(err)
	}
	log.Print("Sending ", string(json))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		config.Logger.Print(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
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
	config.Logger.Print("Received prime number from ", url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	prime := getUnMarshalledPrime(string(body))
	return prime, nil
}

// getUnMarshalledPrime produces a computation from a JSON string
func getUnMarshalledPrime(body string) primes.Prime {
	var p primes.Prime
	err := json.Unmarshal([]byte(body), &p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

// sendComputationResult sends a JSON string through POST to the server
// of the results of a computation
func sendComputationResult(c computation.Computation) {
	url := "http://" + config.Address + config.HeavyReturnPoint
	json, err := json.Marshal(c)
	if err != nil {
		config.Logger.Print(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		config.Logger.Print(err)
	}
	defer resp.Body.Close()
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
	url := "http://" + config.Address + config.HeavyAssignmentPoint
	resp, err := http.Get(url)
	if err != nil {
		log.Print("Cannot connect to server")
		return computation.Computation{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	computation := getUnMarshalledComputation(string(body))
	return computation, nil
}

// LaunchClient launches the client application, and manages
// goroutines
func LaunchClient(c *app.Context) {
	isHeavy := c.Bool("heavy")
	sc := make(chan os.Signal, 1)
	stemComputations := false
	signal.Notify(sc, os.Interrupt)
	go func() {
		counter := 1
		for sig := range sc {
			stemComputations = true
			if counter == 2 {
				os.Exit(1)
			}
			fmt.Printf("\n\nCaptured %v, stemming acception of computations.. (Ctrl-C again to quit)\n\n", sig)
			counter++
		}
	}()
	if isHeavy {
		computationsToPerform := make(chan computation.Computation, 10)
		validComputations := make(chan computation.Computation, 10)
		invalidComputations := make(chan computation.Computation, 10)

		go func() {
			for {
				if stemComputations == false {
					nextComputation, err := fetchNextComputationToPerform()
					if err != nil {
						time.Sleep(1 * time.Second)
						log.Print("Retrying connection")
						continue
					}
					computationsToPerform <- nextComputation
				}
			}
		}()

		go func() {
			for c := range validComputations {
				config.Logger.Printf("%s / %s valid.", c.Prime.Value, c.Divisor)
				sendComputationResult(c)
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
				computationIsValid := computation.RunDistributedComputation(c)
				duration := time.Now().Sub(start)
				newPrimeDuration := c.Prime.TimeTaken + duration
				if computationIsValid == true {
					computationPrime := primes.Prime{
						TimeTaken: newPrimeDuration,
						Value:     i,
						Id:        config.Id,
					}
					validComputations <- computation.Computation{
						Prime:         computationPrime,
						Divisor:       c.Divisor,
						IsValid:       computationIsValid,
						TimeTaken:     duration,
						ComputationId: c.ComputationId,
						Hash:          c.Hash,
					}
				} else {
					computationPrime := primes.Prime{
						TimeTaken: newPrimeDuration,
						Value:     i,
					}
					invalidComputations <- computation.Computation{
						Prime:         computationPrime,
						Divisor:       c.Divisor,
						IsValid:       computationIsValid,
						TimeTaken:     duration,
						ComputationId: c.ComputationId,
						Hash:          c.Hash,
					}
				}
			}(i, c)
		}
	} else if !isHeavy {
		primesToCompute := make(chan primes.Prime, 100)
		validPrimes := make(chan primes.Prime, 100)
		invalidPrimes := make(chan primes.Prime, 100)

		go func() {
			for {
				if stemComputations == false {
					nextPrime, err := fetchNextPrimeToPerform()
					if err != nil {
						time.Sleep(1 * time.Second)
						log.Print("Retrying connection")
						continue
					}
					primesToCompute <- nextPrime
				}
			}
		}()

		go func() {
			for p := range validPrimes {
				primes.DisplayPrimePretty(p.Value, p.TimeTaken)
				err := sendPrimeResult(p)
				for err != nil {
					time.Sleep(1 * time.Second)
					log.Print("Cannot send data back to server, trying again...")
					err = sendPrimeResult(p)
				}
			}
		}()

		go func() {
			for p := range invalidPrimes {
				primes.DisplayFailPretty(p.Value, p.TimeTaken)
				err := sendPrimeResult(p)
				for err != nil {
					time.Sleep(1 * time.Second)
					log.Print("Cannot send data back to server, trying again...")
					err = sendPrimeResult(p)
				}
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
}
