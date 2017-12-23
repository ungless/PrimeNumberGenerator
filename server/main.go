package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/MaxTheMonster/PrimeNumberGenerator/computation"
	app "github.com/urfave/cli"
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	fmt.Printf("%s: Sending hash", port)
	c := computation.GetNextComputation()
	json, err := computation.GetJSON(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", json)
}

func LaunchServer(c *app.Context) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
