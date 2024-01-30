package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	// Set a custom header (optional)
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	fmt.Fprintf(w, "OK")
	// fmt.Fprintln(w, "status:available")
	// fmt.Fprintf(w, "Enviroment:%s\n", app.config.env)
	// fmt.Fprintf(w, "Version: %s\n", version)
}
