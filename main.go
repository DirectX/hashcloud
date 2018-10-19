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
	"strings"
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
	router.HandleFunc("/api/v1/users/{address}/files", UserFilesOptionsHandler).Queries("signature", "{signature}").Methods("OPTIONS")
	router.HandleFunc("/api/v1/users/{address}/files", UserFilesPostHandler).Queries("signature", "{signature}").Methods("POST")
	router.HandleFunc("/api/v1/users/{address}/files", UserFilesGetHandler).Queries("signature", "{signature}").Methods("GET")
	router.HandleFunc("/api/v1/users/{address}/files/{hash}", UserFileOptionsHandler).Queries("signature", "{signature}").Methods("OPTIONS")
	router.HandleFunc("/api/v1/users/{address}/files/{hash}", UserFileGetHandler).Queries("signature", "{signature}").Methods("GET")
	router.HandleFunc("/api/v1/users/{address}/files/{hash}", UserFileUpdateHandler).Queries("signature", "{signature}").Methods("UPDATE")
	router.HandleFunc("/api/v1/users/{address}/files/{hash}", UserFileDeleteHandler).Queries("signature", "{signature}").Methods("DELETE")
	
	log.Println("Listening at port 3010...")
	log.Fatal(http.ListenAndServe(":3010", router))
}

func UserFilesOptionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Options")

	response := Response{
		OK: true,
		ErrorMessage: "",
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// Bulk file upload
func UserFilesPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	signature := vars["signature"]

	hashes := []string{}         // All file hashes
	hashesUploaded := []string{} // Hashes of actually uploaded files

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
					ACL:         map[string]int{address: RoleOwner},
					Filename:    handler.Filename,
					ContentType: handler.Header.Get("Content-Type"),
					ContentSize: handler.Size,
				}

				meta, _ := json.Marshal(fileMeta)
				_ = ioutil.WriteFile(metaFileName, meta, 0644)

				// Creating user's directory if necessary
				storageUserPath := filepath.Join(storagePath, "users", address)
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

	if CheckSignature("upload+" + strings.Join(hashes, "+"), signature, address) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if err := json.NewEncoder(w).Encode(hashesUploaded); err != nil {
			panic(err)
		}
	} else {
		http.Error(w, "Forbidden", 403)
	}
}

func UserFilesGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	signature := vars["signature"]

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

	if CheckSignature("list+" + address, signature, address) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if err := json.NewEncoder(w).Encode(fileMetas); err != nil {
			panic(err)
		}
	} else {
		http.Error(w, "Forbidden", 403)
	}
}

func UserFileOptionsHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		OK: true,
		ErrorMessage: "",
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,UPDATE,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func UserFileGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	hash := vars["hash"]
	signature := vars["signature"]

	metaFileName := filepath.Join(storagePath, "meta", hash)
	if _, err := os.Stat(metaFileName); os.IsNotExist(err) {
		http.NotFound(w, r)
	} else {
		meta, err := ioutil.ReadFile(metaFileName)

		if err != nil {
			http.Error(w, "Error reading file", 500)
		} else {
			var fileMeta FileMeta
			json.Unmarshal(meta, &fileMeta)

			currentUserRole, ok := fileMeta.ACL[address]
			if !ok || !(currentUserRole == RoleOwner || currentUserRole == RoleManager || currentUserRole == RoleViewer) {
				http.Error(w, "Forbidden", 403)
			} else {
				dataFileName := filepath.Join(storagePath, "data", hash)

				if _, err := os.Stat(dataFileName); os.IsNotExist(err) {
					http.NotFound(w, r)
				} else {
					if CheckSignature("download+" + hash, signature, address) {
						w.Header().Set("Access-Control-Allow-Origin", "*")
						w.Header().Set("Access-Control-Allow-Methods", "GET")
						w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
						w.Header().Set("Content-Type", fileMeta.ContentType)
						http.ServeFile(w, r, dataFileName)
					} else {
						http.Error(w, "Forbidden", 403)
					}
				}
			}
		}
	}
}

func UserFileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	hash := vars["hash"]
	signature := vars["signature"]

	metaFileName := filepath.Join(storagePath, "meta", hash)
	if _, err := os.Stat(metaFileName); os.IsNotExist(err) {
		http.NotFound(w, r)
	} else {
		meta, err := ioutil.ReadFile(metaFileName)

		if err != nil {
			http.Error(w, "Error reading file", 500)
		} else {
			var fileMeta FileMeta
			json.Unmarshal(meta, &fileMeta)

			if fileMeta.ACL[address] == RoleNone || fileMeta.ACL[address] == RoleViewer {
				http.Error(w, "Forbidden", 403)
			} else {
				currentUserRole, ok := fileMeta.ACL[address]
				if !ok || !(currentUserRole == RoleOwner || currentUserRole == RoleManager) {
					http.Error(w, "Forbidden", 403)
					return
				}

				if !CheckSignature("share+" + hash, signature, address) {
					http.Error(w, "Forbidden", 403)
					return
				}

				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				var newACL ACL
				err = json.Unmarshal(b, &newACL)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				for newAddress, newRole := range newACL {
					if address != newAddress {
						if newRole > currentUserRole || newRole == RoleNone {
							if newRole != RoleNone {
								fileMeta.ACL[newAddress] = newRole

								// Creating user's directory if necessary
								storageUserPath := filepath.Join(storagePath, "users", newAddress)
								if _, err := os.Stat(storageUserPath); os.IsNotExist(err) {
									os.Mkdir(storageUserPath, 0755)
								}

								// Creating empty file hash placeholder in user's directory
								userFileName := filepath.Join(storageUserPath, hash)
								_ = ioutil.WriteFile(userFileName, []byte{}, 0644)
							} else {
								delete(fileMeta.ACL, newAddress)

								userFileName := filepath.Join(storagePath, "users", newAddress, hash)
								os.Remove(userFileName)
							}
						}
					}
				}

				meta, _ := json.Marshal(fileMeta)
				_ = ioutil.WriteFile(metaFileName, meta, 0644)

				response := Response{
					OK: true,
					ErrorMessage: "",
				}

				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "UPDATE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")

				if err := json.NewEncoder(w).Encode(response); err != nil {
					panic(err)
				}
			}
		}
	}
}

func UserFileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	hash := vars["hash"]
	signature := vars["signature"]

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

			if fileMeta.ACL[address] != RoleOwner {
				http.Error(w, "Forbidden", 403)
			} else {
				currentUserRole, ok := fileMeta.ACL[address]
				if !ok || currentUserRole != RoleOwner {
					http.Error(w, "Forbidden", 403)
				}

				if !CheckSignature("delete+" + hash, signature, address) {
					http.Error(w, "Forbidden", 403)
					return
				}

				for address, _ := range fileMeta.ACL {
					delete(fileMeta.ACL, address)

					userFileName := filepath.Join(storagePath, "users", address, hash)
					os.Remove(userFileName)

					storageUserPath := filepath.Join(storagePath, "users", address)
					os.Remove(storageUserPath)
				}

				metaFileName := filepath.Join(storagePath, "meta", hash)
				os.Remove(metaFileName)

				dataFileName := filepath.Join(storagePath, "data", hash)
				os.Remove(dataFileName)

				response := Response{
					OK: true,
					ErrorMessage: "",
				}

				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")

				if err := json.NewEncoder(w).Encode(response); err != nil {
					panic(err)
				}
			}
		}
	}
}