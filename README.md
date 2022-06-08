# FileShare
FileShare is an effort to prototype a distributed, peer-to-peer file sharing (or torrenting) network. The Peers are different nodes or users in the network who may wish to share files with one another, while the SwarmMaster acts almost like an indexing server, helping a Peer to join the network and find files on other Peers.  

## Run Project in Local
```
cd src/main
go run main.go

```

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

#### Get Active nodes:
```
localhost:5000/getActivePeers
```
#### Upload file:
```
localhost:5000/upload
```