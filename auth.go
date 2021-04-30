package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

// Credentials holds user credential data: Username and Password
type Credentials struct {
	Username string
	Password []byte
	Email    string
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func getUserID(userName string) (userID int) {
	result := db.QueryRow("SELECT user_id FROM users WHERE username=$1", userName)
	err := result.Scan(&userID)

	if err != nil {
		fmt.Printf("ERROR getUserID(%s): %s\n", userName, err)
		return
	}

	return userID
}

func getUserNameFromID(userID int) (userName string) {
	result := db.QueryRow("SELECT username FROM users WHERE user_id=$1", userID)
	err := result.Scan(&userName)

	if err != nil {
		fmt.Printf("ERROR getUserNameFromID(%d): %s\n", userID, err)
		return
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

	err := tmpl[tmplLogin].Execute(response, nil)

	if err != nil {
		return
	}
}

// LogoutHandler clears session cookies and redirects user to homepage
func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

func registerHandler(response http.ResponseWriter, request *http.Request) {
	if getUserName(request) != "" {
		http.Redirect(response, request, "/", 302)
	}

	err := tmpl[tmplRegister].Execute(response, nil)

	if err != nil {
		return
	}
}

func postLoginHandler(response http.ResponseWriter, request *http.Request) {
	email := request.FormValue("email")
	pass := request.FormValue("passwd")
	redirectTarget := urlLogin

	if email != "" && pass != "" {
		var err error

		inputPassword := request.FormValue("passwd")

		u := Credentials{
			Email: request.FormValue("email"),
			Password: nil,
		}

		result := db.QueryRow("SELECT username, password FROM users WHERE email=$1", u.Email)

		if err != nil {
			// If there is an issue with the database, return a 500 error
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		storedCreds := &Credentials{}

		err = result.Scan(&u.Username, &storedCreds.Password)
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
			// If the two passwords DO NOT MATCH; return a 401 status
			response.WriteHeader(http.StatusUnauthorized)
			redirectTarget = urlLogin
		} else {
			// If passwords MATCH; set session cookie and send user to homepage
            setSession(u.Username, response)
			redirectTarget = "/"

			// Update user's last login date
			if _, err = db.Query("UPDATE users SET last_login = $1 WHERE email = $2", getDate(), email); err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("ERROR postLoginHandler: %s", err)
				return
			}
		}
	}
	http.Redirect(response, request, redirectTarget, 302)
}

func postRegisterHandler(response http.ResponseWriter, request *http.Request) {
	// Check request method first
	if request.Method != "POST" {
		fmt.Printf("ERROR postRegisterHandler MethodCheck")
		return
	}

	// Check if any field is empty
	if request.FormValue("name") == "" || request.FormValue("passwd") == "" || request.FormValue("email") == "" {
		fmt.Printf("ERROR postRegisterHandler ValueChecks")
		return
	}

	// Check if the e-mail is valid or not
	if isEmailValid(request.FormValue("email")) != true {
		fmt.Printf("ERROR postRegisterHandler EmailCheck(%s)", request.FormValue("email"))
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.FormValue("passwd")), hashCost)

	if err != nil {
		fmt.Printf("ERROR postRegisterHandler hashedPassword: %s", err)
		return
	}

	// Create data to hold them all for now
	u := Credentials{
		Username: request.FormValue("name"),
		Password: hashedPassword,
		Email:    request.FormValue("email"),
	}

	// Insert data into database
	if _, err = db.Query("INSERT INTO users (username,password,email,register_date,blocked) VALUES ($1,$2,$3,$4,$5)", u.Username, string(u.Password), u.Email, getDate(), false); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("ERROR postRegisterHandler intoDatabase: %s", err)
		return
	}

	// Redirect user to login page
	http.Redirect(response, request, urlLogin, 302)
}
