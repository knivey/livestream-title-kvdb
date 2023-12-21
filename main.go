package main

import (
	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
)


func checkKey(key string) bool {
	for _, val := range config.Keys {
		if val == key {
			return true
		}
	}
	return false
}

func update(w http.ResponseWriter, r *http.Request) {
	k, ok := r.Header["Bearer"]
	if !ok {
		http.Error(w, "You must provide a Bearer key", http.StatusForbidden)
		return
	}
	key := k[0]
	if !checkKey(key) {
		http.Error(w, "Invalid Bearer key", http.StatusForbidden)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("root"))
		if err != nil {
			return err
		}
		err = bucket.Put([]byte("title"), data)
		return err
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "DB error", http.StatusInternalServerError)
	}
}
func get(w http.ResponseWriter, r *http.Request) {
	var val []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("root"))
		if bucket == nil {
			return nil
		}
		val = bucket.Get([]byte("title"))
		return nil
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "DB error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(val)
	if err != nil {
		log.Println(err)
	}
}

var db *bolt.DB

type Config struct {
	Listen string
	Keys []string
}
var config *Config

func main() {
	var err error
	db, err = bolt.Open("kvdb.db", 0600, nil)
	if err != nil {
		log.Fatalf("error opening database: %s\n", err)
	}
	defer db.Close()

	dat, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("error reading config: %s\n", err)
	}
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/update", update)
	http.HandleFunc("/", get)


	log.Fatal(http.ListenAndServe(config.Listen, nil))
}