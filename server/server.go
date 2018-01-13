package server

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/MaxTheMonster/PrimeNumberGenerator/config"
	"github.com/MaxTheMonster/PrimeNumberGenerator/primes"
	"github.com/MaxTheMonster/PrimeNumberGenerator/storage"

	app "github.com/urfave/cli"
)

var lock sync.Mutex

// receivePrimeHandler receives POST data from clients
func receivePrimeHandler(w http.ResponseWriter, r *http.Request, primesReceived chan primes.Prime) {
	lock.Lock()
	defer lock.Unlock()
	decoder := json.NewDecoder(r.Body)
	var p primes.Prime
	err := decoder.Decode(&p)
	if err != nil {
		config.Logger.Fatal(err)
	}
	defer r.Body.Close()
	primesReceived <- p
}

// assignPrimeHandler returns the next prime needed to be calculated
func assignPrimeHandler(w http.ResponseWriter, r *http.Request, p primes.Prime) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	config.Logger.Printf("Sending %v to %s\n", p, ip)
	json, err := json.Marshal(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", json)
}

// LaunchServer runs a server on the configured IP and port
func LaunchServer(c *app.Context) {
	go fmt.Println("Launching server on port 8080...")
	numbersToCheck := make(chan *big.Int)
	validPrimes := make(chan primes.Prime, 100)
	invalidPrimes := make(chan primes.Prime, 100)
	primesToBeSent := make(chan primes.Prime)
	primesReceived := make(chan primes.Prime)

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
			if len(primeBuffer) == config.MaxBufferSize {
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

	http.HandleFunc(config.AssignmentPoint, func(w http.ResponseWriter, r *http.Request) {
		p := <-primesToBeSent
		assignPrimeHandler(w, r, p)
	})

	http.HandleFunc(config.ReturnPoint, func(w http.ResponseWriter, r *http.Request) {
		receivePrimeHandler(w, r, primesReceived)
	})

	http.ListenAndServe(":"+config.Port, nil)
}
