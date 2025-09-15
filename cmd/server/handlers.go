package main

import (
	"fmt"
	"net/http"
)

func (app *application) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome home")
}
