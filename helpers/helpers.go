package helpers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func ParseUintParam(r *http.Request, paramName string) (uint, error) {
	vars := mux.Vars(r)
	idParam := vars[paramName]
	parsed, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
