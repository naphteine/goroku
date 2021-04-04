package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const captionCount = 20

// Caption holds poster, caption and hidden (bool) data
type Caption struct {
	poster  string
	caption string
	hidden  bool
}

// DisplayEntry represent entry details shown to user
type DisplayEntry struct {
	Poster string
	Entry  string
	Date   string
}

var captionIDList [captionCount]int
var captionList [captionCount]string
var captionPosterList [captionCount]string

func getCaptionID(caption string) (captionID int) {
	result := db.QueryRow("SELECT caption_id FROM captions WHERE caption=$1", caption)
	err := result.Scan(&captionID)

	if err != nil {
		fmt.Printf("ERROR getCaptionID(%s): %s\n", caption, err)
		return
	}

	return captionID
}

func getCaptionsAndPosters() {
	rows, err := db.Query("SELECT caption_id, caption, user_id FROM captions ORDER BY caption_id DESC LIMIT $1", captionCount)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var count = 0
	for rows.Next() {
		var id int
		var caption string
		var poster string
		err = rows.Scan(&id, &caption, &poster)
		if err != nil {
			panic(err)
		}

		if count < captionCount {
			captionIDList[count] = id
			captionList[count] = caption
			captionPosterList[count] = poster
		}

		count++
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
}

func captionHandler(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles(tmplCaption)

	vars := mux.Vars(request)

	rows, err := db.Query("SELECT entry, user_id, date FROM entries WHERE caption_id=$1", vars["caption"])
	if err != nil {
		fmt.Printf("ERROR captionHandler 1: %s\n", err)
		panic(err)
	}
	defer rows.Close()

	var entries []DisplayEntry

	for rows.Next() {
		var entry string
		var poster int
		var date string
		err = rows.Scan(&entry, &poster, &date)
		if err != nil {
			fmt.Printf("ERROR captionHandler 2: %s\n", err)
			panic(err)
		}

		entries = append(entries, DisplayEntry{Poster: getUserNameFromID(poster), Entry: entry, Date: date})
	}

	err = rows.Err()
	if err != nil {
		fmt.Printf("ERROR captionHandler 3: %s\n", err)
		panic(err)
	}

	err = t.Execute(response, entries)

	if err != nil {
		return
	}
}

func captionCreateHandler(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles(tmplCaptionCreate)

	type user struct {
		Username string
	}

	userData := user{
		Username: getUserName(request),
	}

	if err != nil {
		return
	}

	if userData.Username == "" {
		http.Redirect(response, request, "/", 302)
	}

	err = t.Execute(response, userData)

	if err != nil {
		return
	}
}

func postCaptionCreateHandler(response http.ResponseWriter, request *http.Request) {
	// Check request method
	if request.Method != "POST" {
		return
	}

	// Get fields
	caption := request.FormValue("caption")
	entry := request.FormValue("entry")
	username := getUserName(request)
	userid := getUserID(username)
	date := getDate()

	// Check if any of the fields are empty
	if caption == "" || entry == "" || username == "" {
		http.Redirect(response, request, urlCaptionCreate, 302)
	}
	// Add caption to database
	var err error
	if _, err = db.Query("INSERT INTO captions (user_id,caption,date,hidden) VALUES ($1,$2,$3,$4)", userid, strings.ToLower(caption), date, false); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add entry to database
	if _, err = db.Query("INSERT INTO entries (caption_id,user_id,entry,date,hidden) VALUES ($1,$2,$3,$4,$5)", getCaptionID(strings.ToLower(caption)), userid, entry, date, false); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update homepage list
	getCaptionsAndPosters()

	// Redirect user
	http.Redirect(response, request, "/", 302)
}
