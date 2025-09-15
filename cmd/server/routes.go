package main

import (
	"net/http"
)

func (app *application) NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.handleHome)
	return mux
}
