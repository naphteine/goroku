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

func getDate() string {
	current := time.Now().UTC()
	return current.String()
}

func indexHandler(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles("templates/index.html")

	type user struct {
		Username string
	}

	userData := user{
		Username: getUserName(request),
	}

	if err != nil {
		return
	}

	err = t.Execute(response, userData)

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

	// Handle pages
	router.HandleFunc("/", indexHandler)

	// Auth routing
	router.HandleFunc(urlLogin, loginHandler)
	router.HandleFunc(urlRegister, registerHandler)
	router.HandleFunc("/post/login", postLoginHandler).Methods("POST")
	router.HandleFunc("/post/logout", postLogoutHandler).Methods("POST")
	router.HandleFunc("/post/register", postRegisterHandler).Methods("POST")

	// IDK what this does but seems necessary
	http.Handle("/", router)

	// Start server
	log.Fatal(http.ListenAndServe(Host+":"+strconv.Itoa(*portPtr), nil))

	// Shutting down
	db.Close()
}
