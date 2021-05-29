package handlers

import (
	"log"
	"net/http"
)

func internalServerError(w http.ResponseWriter, logMessage string) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Printf(logMessage)
	w.Write([]byte("something went terribly wrong"))
}

func badRequest(w http.ResponseWriter, logMessage string) {
	w.WriteHeader(http.StatusBadRequest)
	log.Printf(logMessage)
	w.Write([]byte("bad request"))
}

func notFound(w http.ResponseWriter, logMessage string) {
	w.WriteHeader(http.StatusNotFound)
	log.Printf(logMessage)
	w.Write([]byte("not found"))
}
