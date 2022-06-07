/*
	This file acts as a "test" program to simulate a FileShare network.
*/

package main

import (
	"FileShare/src/fileshare"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func pingFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pong")
}

func createPeer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: CreatePeer")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter id in order to create new Peer")
	}

	resBytes := []byte(reqBody)        // Converting the string "res" into byte array
	var jsonRes map[string]interface{} // declaring a map for key names as string and values as interface
	_ = json.Unmarshal(resBytes, &jsonRes)

	var id int = int(jsonRes["id"].(float64))
	fmt.Println("id", id)
	testDirectory := "testdirs/peer" + strconv.Itoa(id) + "/"
	port := ":" + strconv.Itoa(60120+id)
	p1 := fileshare.MakePeer(id, testDirectory, port)

	w.Header().Set("Contetnt-Type", "application/json")
	json.NewEncoder(w).Encode(p1)
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
	router.HandleFunc("/createPeer", createPeer).Methods("POST")

	log.Fatal(http.ListenAndServe(":5000", router))
}
