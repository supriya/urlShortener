package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url_shortener/models"

	"github.com/gorilla/mux"
)

type Env struct {
	urls models.URLModel
}

func (env *Env) shortenURL(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("url")
	url, err := env.urls.Shorten(longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urlResp := models.URL{LongURL: url.LongURL, ShortURL: url.ShortURL, ExpiryTime: url.ExpiryTime}
	js, err := json.Marshal(urlResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func (env *Env) unshortenURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("url")
	url, err := env.urls.Unshorten(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urlResp := models.URL{LongURL: url.LongURL, ShortURL: url.ShortURL, ExpiryTime: url.ExpiryTime}
	js, err := json.Marshal(urlResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("Database Error:%s\n", err)
	}

	env := &Env{urls: models.URLModel{DB: db}}

	router := mux.NewRouter()
	router.HandleFunc("/shorten", env.shortenURL).Methods("GET")
	router.HandleFunc("/unshorten", env.unshortenURL).Methods("GET")

	server := http.Server{
		Addr:        ":8080",
		Handler:     router,
		ReadTimeout: 5 * time.Second,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error listening %s\n", err)
		}
	}()

	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down the server%s\n", err)
	}
	log.Print("Server shutdown properly")
}
