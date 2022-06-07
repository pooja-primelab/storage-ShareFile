/*
	This file acts as a "test" program to simulate a FileShare network.
*/

package main

import (
	"FileShare/src/fileshare"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func pingFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Contetnt-Type", "application/json")
	json.NewEncoder(w).Encode(`Pong`)
}

func main() {
	start := time.Now()
	m := fileshare.MakeSwarmMaster()
	t1 := time.Now()
	elapsed := t1.Sub(start)

	router := mux.NewRouter()

	// Create Swarm Master
	fmt.Printf("SwarmMaster start time: %v\n", elapsed)
	m.MasterTest()

	// All the routes
	router.HandleFunc("/ping", pingFunc).Methods("GET")

	log.Fatal(http.ListenAndServe(":5000", router))
}
