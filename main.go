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
	"time"
)

var storagePath string

func main() {
	storagePath = filepath.Join(".", "storage")

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0755)

		storageDataPath := filepath.Join(storagePath, "data")
		if _, err := os.Stat(storageDataPath); os.IsNotExist(err) {
			os.Mkdir(storageDataPath, 0755)
		}

		storageMetaPath := filepath.Join(storagePath, "meta")
		if _, err := os.Stat(storageMetaPath); os.IsNotExist(err) {
			os.Mkdir(storageMetaPath, 0755)
		}

		storageUsersPath := filepath.Join(storagePath, "users")
		if _, err := os.Stat(storageUsersPath); os.IsNotExist(err) {
			os.Mkdir(storageUsersPath, 0755)
		}
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/v1/upload", UploadHandler).Methods("POST")
	router.HandleFunc("/api/v1/list/{address}", ListHandler).Methods("GET")
	router.HandleFunc("/api/v1/download/{hash}", DownloadHandler).Methods("POST")

	log.Println("Listening at port 3010...")
	log.Fatal(http.ListenAndServe(":3010", router))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(".", "html", "index.html"))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	hashes := []string{}         // All file hashes
	hashesUploaded := []string{} // HAshes of actually uploaded files

	requestMeta := RequestMeta{}
	requestMetaString := r.FormValue("meta")
	if err := json.Unmarshal([]byte(requestMetaString), &requestMeta); err != nil {
		panic(err)
	}

	r.ParseMultipartForm(32 << 20)
	files := r.MultipartForm.File["files"]
	for _, handler := range files {
		file, err := handler.Open()
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		h := sha256.New()
		h.Write(data)
		hash := fmt.Sprintf("%x", h.Sum(nil))
		hashes = append(hashes, hash)

		dataFileName := filepath.Join(storagePath, "data", hash)
		if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
			err = ioutil.WriteFile(dataFileName, data, 0644)

			if err != nil {
				http.Error(w, "Error writing file", 500)
				return
			} else {
				metaFileName := filepath.Join(storagePath, "meta", hash)

				fileMeta := FileMeta{
					Hash:        hash,
					IP:          GetIP(r),
					Timestamp:   time.Now(),
					ACL:         map[string]int{requestMeta.Owner: RoleOwner},
					Filename:    handler.Filename,
					ContentType: handler.Header.Get("Content-Type"),
					ContentSize: handler.Size,
				}

				meta, _ := json.Marshal(fileMeta)
				_ = ioutil.WriteFile(metaFileName, meta, 0644)

				// Creating user's directory if necessary
				storageUserPath := filepath.Join(storagePath, "users", requestMeta.Owner)
				if _, err := os.Stat(storageUserPath); os.IsNotExist(err) {
					os.Mkdir(storageUserPath, 0755)
				}
				
				// Creating empty file hash placeholder in user's directory
				userFileName := filepath.Join(storageUserPath, hash)
				_ = ioutil.WriteFile(userFileName, []byte{}, 0644)

				hashesUploaded = append(hashesUploaded, hash)
			}
		} else {
			log.Println("Data file already exists, skipping...")
		}
	}

	// TODO: check signature

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(hashesUploaded); err != nil {
		panic(err)
	}
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	fileMetas := []FileMetaPublic{}
	files, err := ioutil.ReadDir(filepath.Join(storagePath, "users", address))
    if err == nil {
	    for _, f := range files {
    		hash := f.Name()

			jsonFile, err := os.Open(filepath.Join(storagePath, "meta", hash))
			if err != nil {
				continue
			}
			defer jsonFile.Close()

			byteValue, _ := ioutil.ReadAll(jsonFile)

			var fileMeta FileMetaPublic
			json.Unmarshal(byteValue, &fileMeta)

			fileMetas = append(fileMetas, fileMeta)
    	}
    }

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(fileMetas); err != nil {
		panic(err)
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	requestMeta := RequestMeta{}
	requestMetaString := r.FormValue("meta")
	if err := json.Unmarshal([]byte(requestMetaString), &requestMeta); err != nil {
		panic(err)
	}

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
				// TODO: check signature and owner

				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "POST")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
				w.Header().Set("Content-Type", fileMeta.ContentType)

				http.ServeFile(w, r, dataFileName)
			}
		}
	}
}