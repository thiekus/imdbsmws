package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/mux"
	sqlite "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"strings"
)

type AdminPanelData struct {
	BaseUrl string
	AppVersion string
	GoVersion string
	SQLiteVersion string
	Config ConfigData
}

func NewAdminPanelData(r *http.Request) AdminPanelData {
	sqliteVer, _, _:= sqlite.Version()
	data:= AdminPanelData{
		BaseUrl:       fmt.Sprintf("http://%s/", r.Host),
		AppVersion:    appVersion,
		GoVersion:     runtime.Version(),
		SQLiteVersion: sqliteVer,
		Config:        appConfig,
	}
	return data
}

func adminGetEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	log:= newLog()
	if checkAdminAuth(r) {
		tpl:= template.Must(template.ParseFiles("./admin_tpl.html"))
		data:= NewAdminPanelData(r)
		if err:= tpl.Execute(w, data); err != nil {
			log.Error(err)
			http.Error(w, "Internal Server Error", 500)
		}
	} else {
		log.Printf("User %s not logged in, throw to login page...", r.RemoteAddr)
		http.Redirect(w, r, "login", 302)
	}
}

func adminPostEndpoint(w http.ResponseWriter, r *http.Request) {
	if checkAdminAuth(r) {
		vars:= mux.Vars(r)
		action:= vars["action"]
		switch action {
		case "importDatabase":
			if err:= r.ParseForm(); err == nil {
				basicUrl:= r.FormValue("basicDataUrl")
				filterType:= strings.ToLower(r.FormValue("filterType"))
				filterGenres:= strings.ToLower(r.FormValue("filterGenres"))
				filterFromYear, err:= strconv.Atoi(r.FormValue("filterYearFrom"))
				if err != nil {
					filterFromYear = 0
				}
				filterToYear, err:= strconv.Atoi(r.FormValue("filterYearTo"))
				if err != nil {
					filterToYear = 0
				}
				filterAdult:= len(r.Form["filterAdult"]) > 0
				filter:= ImportFilter{
					Type: filterType,
					Genres: filterGenres,
					FromYear: filterFromYear,
					ToYear: filterToYear,
					ExcludeAdult: filterAdult,
				}
				saveCache:= len(r.Form["saveCache"]) > 0
				useCache:= len(r.Form["useCache"]) > 0
				importDatabase(basicUrl, filter, useCache, saveCache)
				fmt.Fprint(w, "OK")
			}
		case "updateConfig":
			if err:= r.ParseForm(); err == nil {
				username:= r.FormValue("username")
				password:= r.FormValue("password")
				password2:= r.FormValue("password2")
				if (password != password2) || password == "" {
					password = appConfig.Password
				} else {
					password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
				}
				sessionKey:= r.FormValue("sessionKey")
				port:= r.FormValue("port")
				portNum, _:= strconv.Atoi(port)
				newCfg:= ConfigData{
					Username: username,
					Password: password,
					SessionKey: sessionKey,
					ListeningPort: portNum,
				}
				// Save and reload data
				saveConfigData(newCfg)
				appConfig = getConfigData()
				fmt.Fprint(w, "OK")
			}
		default:
			http.Error(w, "404 Not Found", 404)
		}
	} else {
		http.Redirect(w, r, "login", 302)
	}
}
