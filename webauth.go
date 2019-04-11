package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
)

const SessionName = "ThiekusIMDbMoviesWSSession"

func getSessions(r *http.Request) (*sessions.Session, error) {
	return appCookieStore.Get(r, SessionName)
}

func checkAdminAuth(r *http.Request) bool {
	if sess, err:= getSessions(r); err == nil {
		return (sess.Values["LoggedUser"] != nil) && (sess.Values["LoggedUser"] != "")
	}
	return false
}

func LogoutEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	if sess, err:= getSessions(r); err == nil {
		log.Printf("Logging out %s...", sess.Values["LoggedUser"])
		sess.Values["LoggedUser"] = ""
		sess.Save(r, w)
	}
	http.Redirect(w, r, "login", 302)
}

func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	if !checkAdminAuth(r) {
		if r.Method == http.MethodPost {
			if err:= r.ParseForm(); err == nil {
				username:= r.FormValue("username")
				password:= fmt.Sprintf("%x", sha256.Sum256([]byte(r.FormValue("password"))))
				log.Printf("Logging in user %s from %s", username, r.RemoteAddr)
				if appConfig.Username == username && appConfig.Password == password {
					sess, _:= getSessions(r)
					sess.Values["LoggedUser"] = username
					sess.Save(r, w)
					log.Printf("User %s from %s success to logon!", username, r.RemoteAddr)
					http.Redirect(w, r, "admin", 302)
				} else {
					log.Printf("User %s from %s failed to logon!", username, r.RemoteAddr)
				}
			}
		}
		tpl:= template.Must(template.ParseFiles("./login_tpl.html"))
		data:= NewAdminPanelData(r)
		if err:= tpl.Execute(w, data); err != nil {
			http.Error(w, "Internal Server Error", 500)
		}
	} else {
		http.Redirect(w, r, "admin", 302)
	}
}
