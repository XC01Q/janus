package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8081", "port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Dummy backend listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
