package main

import (
	"errors"
	"net/http"
	"strconv"
)

// id returns value of parameter "id" in the path pattern registered in http.ServeMux.
// id must be positive.
func id(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New(`invalid URL path parameter "id"`)
	}

	return id, nil
}
