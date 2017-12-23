package server

import (
	"fmt"
	"net"
	"net/http"

	app "github.com/urfave/cli"
)

func handler(w http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}

	fmt.Printf("%s: Sending hash", port)
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func LaunchServer(c *app.Context) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
