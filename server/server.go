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

func receiveComputationHandler(w http.ResponseWriter, r *http.Request, computationsToPerform chan computation.Computation, validPrimes chan primes.Prime, invalidPrimes chan primes.Prime) {
	lock.Lock()
	defer lock.Unlock()
	decoder := json.NewDecoder(r.Body)
	var c computation.Computation
	err := decoder.Decode(&c)
	if err != nil {
		config.Logger.Fatal(err)
	}
	defer r.Body.Close()
	config.Logger.Print("Received: ", c)
	config.Logger.Print(len(computationsToPerform), c.ComputationId)
	if c.IsValid {
		invalidPrimes <- c.Prime
	} else if len(computationsToPerform) == 0 {
		validPrimes <- c.Prime
	}
}

func assignComputationHandler(w http.ResponseWriter, r *http.Request, computationsToPerform chan computation.Computation) {
	_, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	fmt.Printf("%s: Sending computation\n", port)

	c := <-computationsToPerform
	fmt.Println(c)
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
	computationsToPerform := make(chan computation.Computation)
	var primeBuffer storage.BigIntSlice

	go func() {
		for i := config.LastPrimeGenerated; true; i.Add(i, big.NewInt(2)) {
			numberToTest := big.NewInt(0).Set(i)
			numbersToCheck <- numberToTest
		}
	}()

	go func() {
		for elem := range validPrimes {
			primes.DisplayPrimePretty(elem.Value, elem.TimeTaken)
			if len(primeBuffer) == 1 {
				storage.FlushBufferToFile(primeBuffer)
				primeBuffer = nil
			}
		}
	}()

	go func() {
		for elem := range invalidPrimes {
			if config.ShowFails == false {
				primes.DisplayFailPretty(elem.Value, elem.TimeTaken)
			}
		}
	}()

	go func() {
		for i := range numbersToCheck {
			currentSolvingPrime := primes.Prime{
				TimeTaken: 0 * time.Second,
				Value:     i,
			}
			computation.GetComputationsToPerform(currentSolvingPrime, computationsToPerform)
		}
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assignComputationHandler(w, r, computationsToPerform)
	})
	http.HandleFunc("/finished", func(w http.ResponseWriter, r *http.Request) {
		receiveComputationHandler(w, r, computationsToPerform, validPrimes, invalidPrimes)
	})

	http.ListenAndServe(":8080", nil)
}
