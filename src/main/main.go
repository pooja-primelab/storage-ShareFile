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

var m *fileshare.SwarmMaster
var r *mux.Router

func pingFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(`Pong`)
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
	p1.ConnectServer()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p1)
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

	tempFile, err := ioutil.TempFile("temp", "file.*.txt")
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

	staticFunctionality()
	setupRoutes()
}

func staticFunctionality() {
	//creating 5 peers and connect with server
	p1 := fileshare.MakePeer(1, "testdirs/peer1/", ":60121")
	p2 := fileshare.MakePeer(2, "testdirs/peer2/", ":60122")
	p3 := fileshare.MakePeer(3, "testdirs/peer3/", ":60123")
	p4 := fileshare.MakePeer(4, "testdirs/peer4/", ":60124")
	p5 := fileshare.MakePeer(5, "testdirs/peer5/", ":60125")
	p1.ConnectServer()
	p2.ConnectServer()
	p3.ConnectServer()
	p4.ConnectServer()
	p5.ConnectServer()

	//Register Files
	p1.RegisterFile("test.txt")
	p2.RegisterFile("test2.txt")
	p3.RegisterFile("test3.txt")
	p4.RegisterFile("test4.txt")
	p5.RegisterFile("test5.txt")

	//Peer Connecting with other peers and sharing files
	p1.ConnectPeer(":60122", 2)
	p1.RequestFile(":60122", 2, "test2.txt")

	p2.ConnectPeer(":60123", 3)
	p2.RequestFile(":60123", 3, "test3.txt")

	//Search file
	p1.SearchForFile("test3.txt")
	p2.SearchForFile("test4.txt")
	p3.SearchForFile("test9.txt")
}

func setupRoutes() {
	r = mux.NewRouter()

	r.HandleFunc("/ping", pingFunc).Methods("GET")
	r.HandleFunc("/getActivePeers", getActiveNodes).Methods("GET")
	r.HandleFunc("/upload", upload).Methods("POST")
	r.HandleFunc("/createPeer", createPeer).Methods("POST")

	log.Fatal(http.ListenAndServe(":5000", r))
}
