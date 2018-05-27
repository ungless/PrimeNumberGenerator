package server

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"

	app "github.com/urfave/cli"
)

var lock sync.Mutex

// receiveComputationHandler handles a computation being received via POST
func receiveComputationHandler(w http.ResponseWriter, r *http.Request, computationsReceived chan computation.Computation) {
	lock.Lock()
	defer lock.Unlock()
	decoder := json.NewDecoder(r.Body)
	var c computation.Computation
	err := decoder.Decode(&c)
	if err != nil {
		config.Logger.Fatal(err)
	}
	defer r.Body.Close()
	computationsReceived <- c
}

// receiveComputationHandler handles a computation to assign
func assignComputationHandler(w http.ResponseWriter, r *http.Request, c computation.Computation) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	json, err := json.Marshal(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	config.Logger.Printf("Sending %s to %s\n", string(json), ip)
	fmt.Fprintf(w, "%s", json)
}

// receivePrimeHandler receives POST data from clients
func receivePrimeHandler(w http.ResponseWriter, r *http.Request, primesReceived chan primes.Prime) {
	lock.Lock()
	defer lock.Unlock()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	decoder := json.NewDecoder(r.Body)
	var p primes.Prime
	err := decoder.Decode(&p)
	if err != nil {
		config.Logger.Fatal(err)
	}
	config.Logger.Printf("Received %s as %v from %s", p.Value, p.IsValid, ip)
	defer r.Body.Close()
	primesReceived <- p
}

// assignPrimeHandler returns the next prime needed to be calculated
func assignPrimeHandler(w http.ResponseWriter, r *http.Request, p primes.Prime) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	json, err := json.Marshal(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	config.Logger.Printf("Sending %s to %s\n", string(json), ip)
	fmt.Fprintf(w, "%s", json)
}

// LaunchServer runs a server on the configured IP and port
func LaunchServer(c *app.Context) {
	go fmt.Printf("Launching server on port %s...\n", config.Port)

	numbersToCheck := make(chan *big.Int)
	validPrimes := make(chan primes.Prime, 100)
	invalidPrimes := make(chan primes.Prime, 100)
	primesToBeSent := make(chan primes.Prime)
	primesReceived := make(chan primes.Prime)
	computationsToBeSent := make(chan computation.Computation)
	computationsReceived := make(chan computation.Computation)
	nOfComputationsForPrime := new(big.Int)

	var primeBuffer storage.BigIntSlice

	go func() {
		for i := new(big.Int).Add(config.LastPrimeGenerated, big.NewInt(2)); true; i.Add(i, big.NewInt(2)) {
			numberToTest := big.NewInt(0).Set(i)
			numbersToCheck <- numberToTest
		}
	}()

	go func() {
		for i := range numbersToCheck {
			currentSolvingPrime := primes.Prime{
				TimeTaken: 0 * time.Second,
				Value:     i,
				IsValid:   false,
			}
			currentComputationsToPerform := computation.GetComputationsToPerform(currentSolvingPrime)
			nOfComputationsForPrime = new(big.Int).Sub(big.NewInt(int64(len(currentComputationsToPerform))), big.NewInt(1))
			for _, c := range currentComputationsToPerform {
				computationsToBeSent <- c
			}
		}
	}()

	go func() {
		for c := range computationsReceived {
			if c.IsValid {
				invalidPrimes <- c.Prime
			} else if nOfComputationsForPrime.Cmp(c.ComputationId) == 0 {
				validPrimes <- c.Prime
			}
		}
	}()

	go func() {
		for p := range validPrimes {
			primes.DisplayPrimePretty(p.Value, p.TimeTaken)
			primeBuffer = append(primeBuffer, p.Value)
			if len(primeBuffer) == config.MaxBufferSize {
				storage.FlushBufferToFile(primeBuffer)
				primeBuffer = nil
			}
		}
	}()

	go func() {
		for i := range numbersToCheck {
			primeToCheck := primes.Prime{
				TimeTaken: 0 * time.Second,
				Value:     i,
				IsValid:   false,
			}
			primesToBeSent <- primeToCheck
		}
	}()

	go func() {
		for p := range primesReceived {
			if p.IsValid {
				validPrimes <- p
			} else {
				invalidPrimes <- p
			}
		}
	}()

	http.HandleFunc(config.HeavyAssignmentPoint, func(w http.ResponseWriter, r *http.Request) {
		c := <-computationsToBeSent
		assignComputationHandler(w, r, c)
	})

	http.HandleFunc(config.HeavyReturnPoint, func(w http.ResponseWriter, r *http.Request) {
		receiveComputationHandler(w, r, computationsReceived)
	})

	http.HandleFunc(config.AssignmentPoint, func(w http.ResponseWriter, r *http.Request) {
		p := <-primesToBeSent
		assignPrimeHandler(w, r, p)
	})

	http.HandleFunc(config.ReturnPoint, func(w http.ResponseWriter, r *http.Request) {
		receivePrimeHandler(w, r, primesReceived)
	})

	http.ListenAndServe(":"+config.Port, nil)
}
