package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Program details
const (
	Program   = "goroku"
	Version   = "v1.0"
	Copyright = "All rights reserved. (c) 2021"
	Host      = "localhost"
	Port      = 8080
	hashCost  = 10
	tmplIndex = "templates/index.html"
	tmplParts = "templates/parts.html"
	/*
		Instance = "goroku"
		dbHost   = "localhost"
		dbPort   = 5432
		dbUser   = "user"
		dbPasswd = "pass"
		dbName   = "database"
	*/
)

var router = mux.NewRouter()
var db *sql.DB
var tmpl = make(map[string]*template.Template)

func getDate() string {
	current := time.Now().UTC()
	return current.Format("2006-01-02 15:04:05 -0700")
}

func indexHandler(response http.ResponseWriter, request *http.Request) {
	// Get and prepare data
	rows, err := db.Query("SELECT caption_id, caption FROM captions ORDER BY date DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	type CaptionData struct {
		ID      int
		Caption string
		Link    string
	}

	var captions []CaptionData

	for rows.Next() {
		var id int
		var caption string
		err = rows.Scan(&id, &caption)
		if err != nil {
			panic(err)
		}

		link := urlCaption + "/" + strconv.Itoa(id)
		captions = append(captions, CaptionData{ID: id, Caption: caption, Link: link})
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	// Prepare index data
	type IndexData struct {
		Username string
		Captions []CaptionData
	}

	data := &IndexData{
		Username: getUserName(request),
		Captions: captions,
	}

	// Execute template with prepared data
	err = tmpl[tmplIndex].Execute(response, data)

	if err != nil {
		return
	}
}

func main() {
	// Get flags
	portPtr := flag.Int("port", Port, "HTTP server port")
	flag.Parse()

	// Print details at start
	fmt.Printf("%s :: %s\nStarting up...\n", Copyright, Program)
	fmt.Printf("Version:\t%s\n", Version)
	fmt.Printf("Date:\t\t%s\n", getDate())
	fmt.Printf("Port:\t\t%s\n", strconv.Itoa(*portPtr))

	// Init DB connection
	var err error
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPasswd, dbName)

	if db, err = sql.Open("postgres", psqlconn); err != nil {
		panic(err)
	}

	// Parse the template files
	tmpl[tmplIndex] = template.Must(template.ParseFiles(tmplIndex, tmplParts))
	tmpl[tmplCaption] = template.Must(template.ParseFiles(tmplCaption, tmplParts))
	tmpl[tmplCaptionCreate] = template.Must(template.ParseFiles(tmplCaptionCreate, tmplParts))
	tmpl[tmplLogin] = template.Must(template.ParseFiles(tmplLogin, tmplParts))
	tmpl[tmplRegister] = template.Must(template.ParseFiles(tmplRegister, tmplParts))

	// Handle pages
	router.HandleFunc("/", indexHandler)

	// Auth routing
	router.HandleFunc(urlLogin, loginHandler)
	router.HandleFunc(urlRegister, registerHandler)
	router.HandleFunc(urlPostLogin, postLoginHandler).Methods("POST")
	router.HandleFunc(urlPostLogout, postLogoutHandler).Methods("POST")
	router.HandleFunc(urlPostRegister, postRegisterHandler).Methods("POST")

	// Caption-entry routing
	router.HandleFunc(urlCaptionCreate, captionCreateHandler)
	router.HandleFunc(urlPostCaption, postCaptionCreateHandler).Methods("POST")
	router.HandleFunc(urlCaption+"/{caption:[0-9]+}", captionHandler)
	router.HandleFunc(urlPostEntry+"/{caption:[0-9]+}", postEntryHandler).Methods("POST")

	// IDK what this does but seems necessary
	http.Handle("/", router)

	// Start server
	log.Fatal(http.ListenAndServe(Host+":"+strconv.Itoa(*portPtr), nil))

	// Shutting down
	db.Close()
}
