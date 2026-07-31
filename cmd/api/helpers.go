package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const maxJSONSize = 1_048_576 // bytes

type wrap map[string]any

// id returns the positive integer value of parameter "id" in the path pattern registered in http.ServeMux.
func id(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id path parameter")
	}

	return id, nil
}

// json encodes data to JSON, set Content-Type: application/json, writes status, then writes the encoded data to w.
// If encoding fails, json calls a.serverErr.
// If writing to w fails, json serves status 500 without a body.
func (a *app) json(w http.ResponseWriter, r *http.Request, status int, data any) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(bytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ok calls a.json with status 200.
func (a *app) ok(w http.ResponseWriter, r *http.Request, data any) {
	a.json(w, r, http.StatusOK, data)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONSize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var se *json.SyntaxError
		var ute *json.UnmarshalTypeError
		var iue *json.InvalidUnmarshalError
		var mbe *http.MaxBytesError

		switch {
		case errors.As(err, &se):
			return fmt.Errorf("JSON is malformed at byte number %v", se.Offset)
		// this happens in some syntax error cases.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("JSON is malformed")
		case errors.As(err, &ute):
			if ute.Field != "" {
				return fmt.Errorf("wrong JSON type for value of %q", ute.Field)
			}
			return fmt.Errorf("wrong JSON type at byte number %v", ute.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			k := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("JSON contains unknown key %v", k)
		case errors.As(err, &mbe):
			return fmt.Errorf("body must not be larger than %v bytes", mbe.Limit)
		case errors.As(err, &iue):
			// this is developer's error, so panic instead of returning error.
			panic(fmt.Errorf("decodeJSON: decode destination is not a non-nil pointer: %w", err)) // err instead of iue cause iue is only node in err's tree.
		default:
			return err
		}
	}

	if err = dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
