package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator"
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

func readRequestDatatest(outType interface{}, r *http.Request) (*interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&outType)
	if err != nil {
		return nil, fmt.Errorf("malformed json data: %w", err)
	}
	validate := validator.New()
	err = validate.Struct(outType)
	if err != nil {
		return nil, fmt.Errorf("invalid json data: %w", err)
	}

	return &outType, nil
}
