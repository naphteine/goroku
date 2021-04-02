package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string
	Password []byte
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
	if getUserName(request) != "" {
		http.Redirect(response, request, "/", 302)
	}

	t, err := template.ParseFiles(templateLogin)

	if err != nil {
		return
	}

	err = t.Execute(response, nil)

	if err != nil {
		return
	}
}

func registerHandler(response http.ResponseWriter, request *http.Request) {
	if getUserName(request) != "" {
		http.Redirect(response, request, "/", 302)
	}

	t, err := template.ParseFiles(templateRegister)

	if err != nil {
		return
	}

	err = t.Execute(response, nil)

	if err != nil {
		return
	}
}

func postLoginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	pass := request.FormValue("passwd")
	redirectTarget := urlLogin

	if name != "" && pass != "" {
		var err error

		inputPassword := request.FormValue("passwd")

		u := Credentials{
			Username: request.FormValue("name"),
			Password: nil,
		}

		result := db.QueryRow("select password from users where username=$1", u.Username)

		if err != nil {
			// If there is an issue with the database, return a 500 error
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		storedCreds := &Credentials{}

		err = result.Scan(&storedCreds.Password)
		if err != nil {
			// If an entry with the username does not exist, send an "Unauthorized"(401) status
			if err == sql.ErrNoRows {
				response.WriteHeader(http.StatusUnauthorized)
				return
			}
			// If the error is of any other type, send a 500 status
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Compare the stored hashed password, with the hashed version of the password that was received
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(inputPassword)); err != nil {
			// If the two passwords don't match, return a 401 status
			response.WriteHeader(http.StatusUnauthorized)
			redirectTarget = urlLogin
		} else {
			setSession(name, response)
			redirectTarget = "/"
		}
	}
	http.Redirect(response, request, redirectTarget, 302)
}

func postLogoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

func postRegisterHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		return
	}
	inputPassword := request.FormValue("passwd")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(inputPassword), hashCost)

	if err != nil {
		return
	}

	if request.FormValue("name") == "" {
		return
	}

	u := Credentials{
		Username: request.FormValue("name"),
		Password: hashedPassword,
	}

	if _, err = db.Query("insert into users values ($1, $2)", u.Username, string(u.Password)); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(response, request, "/", 302)
}
