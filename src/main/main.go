/*
	This file acts as a "test" program to simulate a FileShare network.
*/

package main

import (
	"FileShare/src/fileshare"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func pingFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Contetnt-Type", "application/json")
	json.NewEncoder(w).Encode(`Pong`)
}

func main() {
	m := fileshare.MakeSwarmMaster()

	// Create Swarm Master
	m.MasterTest()

	router := mux.NewRouter()

	// All the routes
	router.HandleFunc("/ping", pingFunc).Methods("GET")
	router.HandleFunc("/nodes", func(writer http.ResponseWriter, request *http.Request) {
		nodes := m.GetActiveNodes()

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(nodes)
	})

	log.Fatal(http.ListenAndServe(":5000", router))
}
