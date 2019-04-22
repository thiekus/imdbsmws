package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	sqlite "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const appVersion = "0.1.1"

var appConfig ConfigData
var appServer http.Server
var appServerVer = fmt.Sprintf("ThkTwinkle %s", appVersion)
var appCookieStore *sessions.CookieStore

// Endpoint to perform application shutdown from http request.
// Needs authentication to admin user.
func shutdownEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check authentication before do shutdown
	if checkAdminAuth(r) {
		log := newLog()
		log.Print("Requesting shutdown...")
		go func() {
			time.Sleep(5000 * time.Millisecond)
			log.Print("Shutting down...")
			if err := appServer.Shutdown(context.Background()); err != nil {
				log.Printf("Shutdown error: %s", err)
			}
		}()
		fmt.Fprintf(w, "Goodbye! Server will be shutdown in 5 seconds...")
	} else {
		http.Redirect(w, r, "login", 302)
	}
}

// Main middleware, invoking some tweaks
func appMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := newLog()
		if !strings.HasPrefix(r.URL.Path, "/importStatus") {
			log.Printf("Client %s accessing %s", r.RemoteAddr, r.URL.Path)
		}
		w.Header().Add("Server", appServerVer)
		if strings.HasPrefix(r.URL.Path, "/movies") {
			w.Header().Add("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Printf("Thiekus IMDb Movies WebService codename Twinkle v%s\n", appVersion)
	fmt.Println("Copyright (C) Thiekus 2019")
	sqliteVer, _, _ := sqlite.Version()
	fmt.Printf("Built using %s, SQLite %s\n\n", runtime.Version(), sqliteVer)

	// Easter egg
	now := time.Now()
	if (now.Hour() >= 0) && (now.Hour() < 5) {
		hideData := "TXVuZ2tpbiBoYW55YSBvcmFuZyBnaWxhIGRhbiBrdXJhbmcga2VyamFhbiBiaWtpbiBhcGxpa2FzaSBrYXlhayBnaW5pLCBkaSBqYW0gc2VnaW5pIDop"
		hideStr, _ := base64.StdEncoding.DecodeString(hideData)
		fmt.Printf("%s\n\n", hideStr)
	}
	log := newLog()

	// Get configuration data
	appConfig = getConfigData()
	appCookieStore = sessions.NewCookieStore([]byte(appConfig.SessionKey))

	// Initialize cache dir if not available
	cacheDir := "./caches"
	if !isFileExists(cacheDir) {
		if err := os.Mkdir(cacheDir, os.ModePerm); err == nil {
			log.Printf("Cache directory created at %s", cacheDir)
		} else {
			panic(err)
		}
	}

	// Define our webservice routing
	r := mux.NewRouter()
	r.Use(appMiddleware)
	// Main webservice endpoints
	r.HandleFunc("/movies", moviesGetEndpoint).Methods("GET")
	r.HandleFunc("/movies/{id}", moviesGetEndpoint).Methods("GET")
	r.HandleFunc("/movies", moviesPostEndpoint).Methods("POST")
	r.HandleFunc("/movies/{id}", moviesPutEndpoint).Methods("PUT")
	r.HandleFunc("/movies/{id}", moviesDeleteEndpoint).Methods("DELETE")
	//r.PathPrefix("/caches/").Handler(http.StripPrefix("/caches/", http.FileServer(http.Dir("./caches"))))
	// Admin feature endpoints
	r.HandleFunc("/login", loginEndpoint)
	r.HandleFunc("/logout", LogoutEndpoint)
	r.HandleFunc("/admin", adminGetEndpoint).Methods("GET")
	r.HandleFunc("/admin/{action}", adminPostEndpoint).Methods("POST")
	r.HandleFunc("/importStatus", importStatusEndpoint).Methods("GET")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	// Shutdown endpoints
	r.HandleFunc("/shutdown", shutdownEndpoint).Methods("GET")
	log.Print("Routers have been initialized...")

	// Establish our server
	appServer = http.Server{Addr: fmt.Sprintf(":%d", appConfig.ListeningPort), Handler: r}
	if err := appServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
