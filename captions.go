package main

import (
	"html/template"
	"net/http"
)

const captionCount = 20

type Caption struct {
	poster  string
	caption string
	hidden  bool
}

type Entry struct {
	poster  string
	content string
	caption string
}

var captionIdList [captionCount]int
var captionList [captionCount]string
var captionPosterList [captionCount]string

func captionHandler(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles(templateCaption)

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

func postCaptionHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		return
	}

	caption := request.FormValue("caption")
	entry := request.FormValue("entry")
	username := getUserName(request)

	if caption == "" || entry == "" || username == "" {
		http.Redirect(response, request, urlCaption, 302)
	}

	var err error

	cap := Caption{
		poster:  username,
		caption: caption,
		hidden:  false,
	}

	et := Entry{
		poster:  username,
		content: entry,
		caption: caption,
	}

	if _, err = db.Query("insert into captions (poster, caption, hidden) values ($1, $2, $3)", cap.poster, cap.caption, cap.hidden); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = db.Query("insert into entries (poster, content, caption) values ($1, $2, $3)", et.poster, et.content, et.caption); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	getCaptionsAndPosters()
	http.Redirect(response, request, "/", 302)
}

func getCaptionsAndPosters() {
	rows, err := db.Query("SELECT id, caption, poster FROM captions ORDER BY id DESC LIMIT $1", captionCount)
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
			captionIdList[count] = id
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
