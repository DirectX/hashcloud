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
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", MainHandler).Methods("GET")
	router.HandleFunc("/upload", UploadHandler).Methods("POST")
	router.HandleFunc("/get/{hash}", GetHandler).Methods("GET")

	log.Println("Listening at port 3001...")
	log.Fatal(http.ListenAndServe(":3001", router))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>File Upload</title></head><body><form action=\"/upload\" enctype=\"multipart/form-data\" method=\"post\"><p><input type=\"file\" name=\"file\"><input type=\"submit\" value=\"Upload\"></p></form></body></html>");
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
	
	dataFileName := fmt.Sprintf("./storage/data/%s", hash)

	if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
		err = ioutil.WriteFile(dataFileName, data, 0644)

		if err != nil {
			http.Error(w, "Error writing file", 500)
			return
		} else {
			metaFileName := fmt.Sprintf("./storage/meta/%s", hash)

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

	metaFileName := fmt.Sprintf("./storage/meta/%s", hash)
	
	if _, err := os.Stat(metaFileName); os.IsNotExist(err) {
		http.NotFound(w, r)
	} else {
		meta, err := ioutil.ReadFile(metaFileName)

		if err != nil {
			http.Error(w, "Error reading file", 500)
		} else {
			var fileMeta FileMeta
			json.Unmarshal(meta, &fileMeta)

			dataFileName := fmt.Sprintf("./storage/data/%s", hash)
			
			if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				w.Header().Set("Content-Type", fileMeta.ContentType)
				http.ServeFile(w, r, dataFileName)
			}
		}
	}
}