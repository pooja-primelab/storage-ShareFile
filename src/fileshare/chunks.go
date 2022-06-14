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
	"github.com/ledongthuc/pdf"
)

const (
	dbPath            = "./database"
	cryptoKey         = "teteteteteetesdsdsdsdsdt"
	chunkFileSize int = 256 // bytes
)

var Colors = fileExtentions()

func fileExtentions() *fileExtention {
	return &fileExtention{
		txt:  ".txt",
		pdf:  ".pdf",
		docx: ".docx",
	}
}

type fileExtention struct {
	txt  string
	pdf  string
	docx string
}

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

func GetEncryptedFiles(fileName string, ownername string) []File {

	var chunks []File

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		if string(c.Ownername) == ownername && strings.HasPrefix(c.FileName, fileName) {
			chunks = append(chunks, c)
		}
	}
	return chunks
}

func readPdf(path string) (string, error) {

	f, r, err := pdf.Open(path)
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func ConvertDecryptFiles(fileName string, ownername string) string {

	chunks := GetEncryptedFiles(fileName, ownername)

	tempfile := "./testdirs/" + "final" + chunks[0].FileExtension

	file, err := os.Create(tempfile)
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
	return chunks[0].FileExtension
}

func CreateChunksAndEncrypt(filepath string, m *SwarmMaster, name string, fileExtension string) {

	allChunks := make([]string, 0)

	switch fileExtension {

	case fileExtentions().txt:
		file, err := os.Open(filepath)
		Handle(err)
		defer file.Close()

		info, err := os.Stat(filepath)
		Handle(err)
		chunkSize := (int(info.Size()) / chunkFileSize) + 1

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

		for i := 0; i < chunkSize; i++ {
			if i+1 == chunkSize {
				chunkTexts := texts[i*lengthPerSplit:]
				allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
			} else {
				chunkTexts := texts[i*lengthPerSplit : (i+1)*lengthPerSplit]
				allChunks = append(allChunks, strings.Join(chunkTexts, "\n"))
			}
		}

	case fileExtentions().pdf:
		data, err := readPdf(filepath)
		Handle(err)

		info, err := os.Stat(filepath)
		Handle(err)

		var m int64 = info.Size()
		nSize := int(m)
		chunkSize := nSize / 10000 // 10 KB per chunk
		lengthPerSplit := len(data) / chunkSize

		for i := 0; i < chunkSize; i++ {
			if i+1 == chunkSize {
				chunkTexts := data[i*lengthPerSplit:]
				allChunks = append(allChunks, chunkTexts)
			} else {
				chunkTexts := data[i*lengthPerSplit : (i+1)*lengthPerSplit]
				allChunks = append(allChunks, chunkTexts)
			}
		}

	case fileExtentions().docx:

	default:
		fmt.Println("Not supported File Type")
	}
	fmt.Println("allchunks final", allChunks)

	writefile(allChunks, filepath, m, name, fileExtension)
}

func writefile(data []string, filePath string, m *SwarmMaster, name string, fileExtension string) {

	nodes := m.GetActiveNodes()

	nodesLen := len(nodes)
	counter := 0
	for index, chunk := range data {

		// IF NUMBER OF CHUNKS ARE MORE THAN NUMBER OF CONNECTED PEERS
		if (counter == nodesLen) || (counter > nodesLen) {
			counter = 0
		}

		fileChunk := name + "-" + strconv.Itoa(index) + "-chunks-" + uuid.New().String() + fileExtension
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
		chunks.FileExtension = fileExtension
		chunks.FileName = name
		chunks.Ownername = "StorageTeam"
		chunks.NodeAddress = strconv.Itoa(registerPeers[counter].PeerID)
		chunks.BlockHash = []byte("SomeHash")
		chunks.ChuckIndex = index
		chunks.Port = registerPeers[counter].Port

		SaveFileInfo(chunks)

		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
		file.WriteString(EncryptFile(chunk))
		counter++
	}
}

func SearchFiles(filename string, ownername string) []File {

	chunks := GetChunksByPrefix(filename, ownername)
	return chunks
}

type File struct {
	Chunkname     string
	FilePath      string
	Ownername     string
	FileName      string
	FileExtension string
	NodeAddress   string
	BlockHash     []byte
	ChuckIndex    int
	Port          string
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

func SaveFileInfo(chunk File) []File {

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(chunk)
	key := chunk.Chunkname
	fmt.Println("key is ", key)

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)
	data = append(data, chunk)

	reqBodyBytes2 := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes2).Encode(data)
	ioutil.WriteFile("output.json", reqBodyBytes2.Bytes(), 0644)

	return data
}

func GetChunkByKey(key string) string {

	var finalResponse []byte
	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		newKey := string(c.Chunkname)
		if newKey == key {
			reqBodyBytes2 := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes2).Encode(c)
			finalResponse = reqBodyBytes2.Bytes()
		}
	}
	if len(finalResponse) == 0 {
		return string(`{"message": "No data found"}`)
	}
	return string(finalResponse)
}

func GetChunksByPrefix(prefix string, ownername string) []File {

	var chunks []File

	file, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}
	var data []File
	err = json.Unmarshal(file, &data)

	for _, c := range data {
		if string(c.Ownername) == ownername && strings.HasPrefix(c.FileName, prefix) {
			chunks = append(chunks, c)
		}
	}
	return chunks
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
