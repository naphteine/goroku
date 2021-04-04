package main

import (
	"fmt"
	"net/http"
	"strconv"
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

func getCaptionID(caption string) (captionID int) {
	result := db.QueryRow("SELECT caption_id FROM captions WHERE caption=$1", caption)
	err := result.Scan(&captionID)

	if err != nil {
		fmt.Printf("ERROR getCaptionID(%s): %s\n", caption, err)
		return
	}

	return captionID
}

func getCaptionNameFromID(id int) (captionName string) {
	result := db.QueryRow("SELECT caption FROM captions WHERE caption_id=$1", id)
	err := result.Scan(&captionName)

	if err != nil {
		fmt.Printf("ERROR getCaptionNameFromID(%d): %s\n", id, err)
		return
	}

	return captionName
}

func captionHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	captionID, err := strconv.Atoi(vars["caption"])
	if err != nil {
		fmt.Printf("ERROR captionHandler Atoi: %s\n", err)
		panic(err)
	}

	rows, err := db.Query("SELECT entry, user_id, date FROM entries WHERE caption_id=$1", captionID)
	if err != nil {
		fmt.Printf("ERROR captionHandler 1: %s\n", err)
		panic(err)
	}
	defer rows.Close()

	captionName := getCaptionNameFromID(captionID)

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

	type CaptionData struct {
        PageTitle    string
		CaptionTitle string
		Username     string
		CaptionID    int
        Entries      []DisplayEntry
	}

	cData := &CaptionData{
        PageTitle: captionName+" "+Instance,
		CaptionTitle: captionName,
		Username:     getUserName(request),
		CaptionID:    captionID,
		Entries:      entries,
	}

	err = rows.Err()
	if err != nil {
		fmt.Printf("ERROR captionHandler 3: %s\n", err)
		panic(err)
	}

	err = tmpl[tmplCaption].Execute(response, cData)

	if err != nil {
		return
	}
}

func captionCreateHandler(response http.ResponseWriter, request *http.Request) {
	type user struct {
		Username string
	}

	userData := user{
		Username: getUserName(request),
	}

	if userData.Username == "" {
		http.Redirect(response, request, "/", 302)
	}

	err := tmpl[tmplCaptionCreate].Execute(response, userData)

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

	// Redirect user
	http.Redirect(response, request, urlCaption+"/"+strconv.Itoa(getCaptionID(strings.ToLower(caption))), 302)
}

func postEntryHandler(response http.ResponseWriter, request *http.Request) {
	// Check request method
	if request.Method != "POST" {
		return
	}
	// Get fields
	vars := mux.Vars(request)

	caption, err := strconv.Atoi(vars["caption"])

	if err != nil {
		fmt.Printf("ERROR postEntryHandler Atoi: %s", err)
		panic(err)
	}

	entry := request.FormValue("entry")
	username := getUserName(request)
	userid := getUserID(username)
	date := getDate()

	// Check if any of the fields are empty
	if caption < 0 || entry == "" || username == "" {
		http.Redirect(response, request, urlCaptionCreate, 302)
	}

	// Add entry to database
	if _, err = db.Query("INSERT INTO entries (caption_id,user_id,entry,date,hidden) VALUES ($1,$2,$3,$4,$5)", caption, userid, entry, date, false); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Redirect user
	http.Redirect(response, request, urlCaption+"/"+vars["caption"], 302)
}
