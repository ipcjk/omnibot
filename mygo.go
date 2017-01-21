package main

import (
	"database/sql"
	_ "database/sql/driver"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"io"
)

type transcoder struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Uid int64 `json:"uid"`
	Timestamp int64 `json:"timestamp"`
}

type Handler struct {
	db *sql.DB
}

type GetTranscodes Handler
type GetTranscoder Handler
type DeleteTranscoder Handler
type PutTranscoder Handler
type ModifyTranscoder Handler
type AddTranscoder Handler

func main() {
	db, err := sql.Open("sqlite3", "omnibot.db")
	checkErr(err)
	defer db.Close()

	r := mux.NewRouter()
	r.Handle("/transcode", &GetTranscodes{db: db}).Methods("GET")
	r.Handle("/transcode", &AddTranscoder{db: db}).Methods("POST")
	r.Handle("/transcode/{key}", &GetTranscoder{db: db}).Methods("GET")
	r.Handle("/transcode/{key}", &ModifyTranscoder{db: db}).Methods("PUT")
	r.Handle("/transcode/{key}", &DeleteTranscoder{db: db}).Methods("DEL")
	log.Fatal(http.ListenAndServe(":8080", r))
}


func AddSafeHeaders(w http.ResponseWriter) {
	w.Header().Set("Server", "go1.7.3 darwin/amd64")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
}

func (h *GetTranscodes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AddSafeHeaders(w)
	var hash, name string
	var myTranscoders []transcoder

	rows, err := h.db.Query("select hash,name from omnibot")
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&hash, &name)
		if err != nil {
			log.Fatal(err)
		}
		myTranscoders = append(myTranscoders, transcoder{Name: name, Hash: hash})
	}
	err = rows.Err()
	checkErr(err)

	bytes, err := json.Marshal(myTranscoders)
	checkErr(err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (h *GetTranscoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AddSafeHeaders(w)
	var hash = "FEF3332"
	var name string

	err := h.db.QueryRow("select hash,name from omnibot where hash  = ?", hash).Scan(&hash, &name)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.Marshal(&transcoder{Name: name, Hash: hash})
	checkErr(err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (h *DeleteTranscoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AddSafeHeaders(w)
	vars := mux.Vars(r)
	key := vars["key"]

	stmt, err := h.db.Prepare("delete from omnibot where hash=?")
	checkErr(err)

	res, err := stmt.Exec(key)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "Deleted %d elements", affect)
}

func (h *AddTranscoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var NewTranscoder transcoder
	var name, hash string
	var uid int64

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	checkErr(err)

	err = json.Unmarshal(body, &NewTranscoder)
	checkErr(err)

	stmt, err := h.db.Prepare("insert into omnibot (name, hash) values(?, ?)")
	checkErr(err)

	res, err := stmt.Exec(NewTranscoder.Name,NewTranscoder.Hash)
	checkErr(err)

	uid, err = res.LastInsertId()
	checkErr(err)

	err = h.db.QueryRow("select uid,hash,name from omnibot where uid  = ?", uid).Scan(&uid, &hash, &name)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.Marshal(&transcoder{Name: name, Hash: hash, Uid: uid})
	checkErr(err)

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (h *ModifyTranscoder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Modify Transcode  FMT!\n"))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
