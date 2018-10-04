package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	storagePath := filepath.Join(".", "storage")
	storageDataPath := filepath.Join(storagePath, "data")
	storageMetaPath := filepath.Join(storagePath, "meta")

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0755)

		if _, err := os.Stat(storageDataPath); os.IsNotExist(err) {
			os.Mkdir(storageDataPath, 0755)
		}

		if _, err := os.Stat(storageMetaPath); os.IsNotExist(err) {
			os.Mkdir(storageMetaPath, 0755)
		}
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", MainHandler).Methods("GET")
	router.HandleFunc("/upload", UploadHandler).Methods("POST")
	router.HandleFunc("/get/{hash}", GetHandler).Methods("GET")

	log.Println("Listening at port 3010...")
	log.Fatal(http.ListenAndServe(":3010", router))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(".", "html", "index.html"))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error getting file", 500)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	h := sha256.New()
	h.Write(data)
	hash := fmt.Sprintf("%x", h.Sum(nil))
	
	dataFileName := filepath.Join(".", "storage", "data", hash)

	if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
		err = ioutil.WriteFile(dataFileName, data, 0644)

		if err != nil {
			http.Error(w, "Error writing file", 500)
			return
		} else {
			metaFileName := filepath.Join(".", "storage", "meta", hash)

			fileMeta := FileMeta{
				Filename: handler.Filename,
				ContentType: handler.Header.Get("Content-Type"),
			}

			meta, _ := json.Marshal(fileMeta)

			_ = ioutil.WriteFile(metaFileName, meta, 0644)
		}
	} else {
		log.Println("Data file already exists, skipping...");
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>File Upload</title></head><body><a href=\"/get/%s\">/get/%s</a></body></html>", hash, hash);
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	metaFileName := filepath.Join(".", "storage", "meta", hash)
	
	if _, err := os.Stat(metaFileName); os.IsNotExist(err) {
		http.NotFound(w, r)
	} else {
		meta, err := ioutil.ReadFile(metaFileName)

		if err != nil {
			http.Error(w, "Error reading file", 500)
		} else {
			var fileMeta FileMeta
			json.Unmarshal(meta, &fileMeta)

			dataFileName := filepath.Join(".", "storage", "data", hash)
			
			if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				w.Header().Set("Content-Type", fileMeta.ContentType)
				http.ServeFile(w, r, dataFileName)
			}
		}
	}
}