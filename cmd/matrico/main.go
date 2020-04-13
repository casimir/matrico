package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/casimir/matrico"
)

var (
	argAddr = flag.String("address", "localhost:8765", "server address and port")
)

func main() {
	s := matrico.NewServer()
	log.Printf("listening on %s", *argAddr)
	log.Printf("---- routes -----")
	for _, route := range s.ListRoutes() {
		log.Print("> " + route)
	}
	log.Printf("-----------------")
	log.Fatal(http.ListenAndServe(*argAddr, &s))
}
