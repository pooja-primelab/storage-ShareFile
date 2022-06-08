package main

import (
	"FileShare/src/fileshare"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

var m *fileshare.SwarmMaster
var r *mux.Router

func pingFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(`Pong`)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10)

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	defer file.Close()
	fmt.Println(handler.Filename, "uploaded")

	tempFile, err := ioutil.TempFile("temp", handler.Filename)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(`Successfully Uploaded File`)
}

func getActiveNodes(w http.ResponseWriter, r *http.Request) {
	nodes := m.GetActiveNodes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func main() {
	m = fileshare.MakeSwarmMaster()
	m.MasterTest() // this isn't really needed - we can move the code to

	setupRoutes()
}

func setupRoutes() {
	r = mux.NewRouter()

	r.HandleFunc("/ping", pingFunc).Methods("GET")
	r.HandleFunc("/nodes", getActiveNodes).Methods("GET")
	r.HandleFunc("/upload", upload).Methods("POST")

	log.Fatal(http.ListenAndServe(":5000", r))
}
