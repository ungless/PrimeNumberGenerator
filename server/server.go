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

func assignComputationHandler(w http.ResponseWriter, r *http.Request, c computation.Computation) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	config.Logger.Printf("Sending %v to %s\n", c, ip)
	json, err := computation.GetJSONFromComputation(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", json)
}

func LaunchServer(c *app.Context) {
	go fmt.Println("Launching server on port 8080...")
	numbersToCheck := make(chan *big.Int, 100)
	validPrimes := make(chan primes.Prime, 100)
	invalidPrimes := make(chan primes.Prime, 100)
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
		for p := range validPrimes {
			primes.DisplayPrimePretty(p.Value, p.TimeTaken)
			primeBuffer = append(primeBuffer, p.Value)
			if len(primeBuffer) == 1 {
				storage.FlushBufferToFile(primeBuffer)
				primeBuffer = nil
			}

		}
	}()

	go func() {
		for p := range invalidPrimes {
			if config.ShowFails == true {
				primes.DisplayFailPretty(p.Value, p.TimeTaken)
			}
		}
	}()

	go func() {
		for i := range numbersToCheck {
			currentSolvingPrime := primes.Prime{
				TimeTaken: 0 * time.Second,
				Value:     i,
			}
			currentComputationsToPerform := computation.GetComputationsToPerform(currentSolvingPrime)
			nOfComputationsForPrime = new(big.Int).Sub(big.NewInt(int64(len(currentComputationsToPerform))), big.NewInt(1))
			fmt.Println(nOfComputationsForPrime)
			for _, c := range currentComputationsToPerform {
				computationsToBeSent <- c
			}
			config.Logger.Print("PRIIIIIIIIIIIIIIIIIIIIIIIIME")
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

	http.HandleFunc(config.AssignmentPoint, func(w http.ResponseWriter, r *http.Request) {
		c := <-computationsToBeSent
		assignComputationHandler(w, r, c)
	})

	http.HandleFunc(config.ReturnPoint, func(w http.ResponseWriter, r *http.Request) {
		receiveComputationHandler(w, r, computationsReceived)
	})

	http.ListenAndServe(":8080", nil)
}
