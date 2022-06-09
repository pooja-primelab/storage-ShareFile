# FileShare
This is simple distributed P2P file sharing network allowing nodes to connect, register and share files.   

## Run Project in Local
```
cd src/main
go run main.go
```
This starts a HTTP server running on port: 5000

## Contents 
* `src/`: Contains the code for FileShare, a peer-to-peer distributed file sharing system.
    * `fileshare/`: Contains the Peer and SwarmMaster code, all part of the `fileshare` package.  
    * `main/`: Contains `main.go` and Peer directories, used for running a test case of the FileShare system. 

## How to use
After running ```go run main.go``` a localhost http server is start that runs on port ```5000```. Below are available endpoints

#### Create new peer:
```
localhost:5000/createPeer
```

#### Get nodes:
```
localhost:5000/getActivePeers
```
#### Upload file:
```
localhost:5000/upload
```