package fileshare

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
)

const (
	dbPath       = "./database"
	cryptoKey    = "teteteteteetesdsdsdsdsdt"
	EncryptedLoc = "./chunks/encrypted/"
	DecryptedLoc = "./chunks/decrypted/"
	chunkSize    = 4
)

func ReadDir(dirname string) []os.FileInfo {

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func ReadFile(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func (chain *FileDB) GetEncryptedFiles(fileName string, ownername string) []File {

	var chunks []File

	chain.Database.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fileName)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			fmt.Println("K >", string(k))

			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			fmt.Println("K >", string(k))
			var p2 File

			json.Unmarshal(valCopy, &p2)

			if p2.Ownername == ownername {
				chunks = append(chunks, p2)
			}

		}
		return nil
	})
	return chunks
}

func (chain *FileDB) ConvertDecryptFiles(fileName string, ownername string) {

	chunks := chain.GetEncryptedFiles(fileName, ownername)

	filename := "./testdirs/" + "final.txt"

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range chunks {
		databyte := ReadFile(f.FilePath)
		data := DecryptFile(string(databyte))
		length, err := io.WriteString(file, data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("file length is ", length, data)
	}
	defer file.Close()
	databyte := ReadFile(filename)
	fmt.Println("Actual data of saved file is ", string(databyte))
}

func CreateChunksAndEncrypt(filepath string, m *SwarmMaster, name string) {

	file, err := os.Open(filepath)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	texts := make([]string, 0)
	for scanner.Scan() {
		text := scanner.Text()
		texts = append(texts, text)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	lengthPerSplit := len(texts) / chunkSize
	allChunks := make([]string, 0)

	for i := 0; i < chunkSize; i++ {
		if i+1 == chunkSize {
			chunkTexts := texts[i*lengthPerSplit:]
			allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
		} else {
			chunkTexts := texts[i*lengthPerSplit : (i+1)*lengthPerSplit]
			allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
		}
	}

	writefile(allChunks, filepath, m, name)
}

func writefile(data []string, filePath string, m *SwarmMaster, name string) {

	nodes := []int{}

	nodesLen := len(nodes)
	counter := 0
	for index, chunk := range data {

		// IF NUMBER OF CHUNKS ARE MORE THAN NUMBER OF CONNECTED PEERS
		if (counter == nodesLen) || (counter > nodesLen) {
			counter = 0
		}

		fileChunk := name + "-" + strconv.Itoa(index) + "-chunks-" + uuid.New().String() + ".txt"
		path := "../main/testdirs/peer" + strconv.Itoa(registerPeers[counter].PeerID) + "/" + fileChunk
		file, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		registerPeers[counter].RegisterFile(name)

		// MAINTAIN MENIFEST FILE
		var chunks File

		chunks.Chunkname = fileChunk
		chunks.FilePath = path
		chunks.FileName = name
		chunks.Ownername = "Amandeep"
		chunks.NodeAddress = strconv.Itoa(registerPeers[counter].PeerID)
		chunks.BlockHash = []byte("SomeHash")
		chunks.ChuckIndex = index
		chunks.Port = registerPeers[counter].Port

		inst := SaveFileInfo(chunks)
		inst.Database.Close()

		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		file.WriteString(EncryptFile(chunk))
		counter++
	}
}

func (chain *FileDB) SearchFiles(filename string, ownername string) []File {

	chunks := chain.GetChunksByPrefix(filename, ownername)

	// for index, peer := range registerPeers {

	// 	fmt.Println("peer files with indexx ", index)
	// 	fmt.Println("peer.PeerID ", peer.PeerID)
	// 	fmt.Println("peer.directory ", peer.directory)
	// 	fmt.Println("peer.files ", peer.files)
	// 	fmt.Println("peer.numFiles ", peer.numFiles)
	// 	fmt.Println("peer.numPeers ", peer.numPeers)
	// 	fmt.Println("peer.Port ", peer.Port)
	// }
	return chunks
}

type File struct {
	Chunkname   string
	FilePath    string
	Ownername   string
	FileName    string
	NodeAddress string
	BlockHash   []byte
	ChuckIndex  int
	Port        string
}

type FileDB struct {
	Database *badger.DB
}

func GetDBinstacnce() *FileDB {
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	blockchain := FileDB{Database: db}
	return &blockchain
}

func SaveFileInfo(chunk File) *FileDB {
	db := GetDBinstacnce()

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(chunk)
	key := chunk.Chunkname
	fmt.Println("key is ", key)
	db.Database.Update(func(txn *badger.Txn) error {
		err2 := txn.Set([]byte(key), reqBodyBytes.Bytes())
		return err2
	})

	blockchain := FileDB{Database: db.Database}
	return &blockchain
}

func (chain *FileDB) GetChunkByKey(key string) string {

	var file []byte

	chain.Database.View(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(key))
		if err != nil {
			fmt.Println("Key not found.")
			return err
		}
		file, err = item.Value()
		fmt.Println("Item: ", string(file))
		Handle(err)
		return err
	})

	return string(file)
}

func (chain *FileDB) GetChunksByPrefix(prefix string, ownername string) []File {

	var chunks []File
	// var dst []string

	chain.Database.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()

			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			fmt.Println("K >", string(k))
			var p2 File

			json.Unmarshal(valCopy, &p2)

			if p2.Ownername == ownername {
				chunks = append(chunks, p2)
			}
		}
		return nil
	})
	return chunks
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
