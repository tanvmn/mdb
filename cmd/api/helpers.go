package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
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

// json encodes data to JSON, set "Content-Type: application/json", writes status, then writes the encoded data to w.
// If encoding fails, json calls a.serverErr.
// If writing to w fails, json serves status 500 without a body.
func (a *app) json(w http.ResponseWriter, r *http.Request, status int, data any) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
	bytes = append(bytes, '\n')

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

// created calls a.json with status 201.
func (a *app) created(w http.ResponseWriter, r *http.Request, data any) {
	a.json(w, r, http.StatusCreated, data)
}

// accepted calls a.json with status 202.
func (a *app) accepted(w http.ResponseWriter, r *http.Request, data any) {
	a.json(w, r, http.StatusAccepted, data)
}

// succeeded calls a.json with status 200 and wrap{"message":"succeeded"} as data.
func (a *app) succeeded(w http.ResponseWriter, r *http.Request) {
	a.ok(w, r, wrap{"message": "succeeded"})
}

// decodeJSON decodes a single JSON value from rc into the value pointed to by dst.
// dst must be a non-nil pointer.
func decodeJSON(w http.ResponseWriter, rc io.ReadCloser, dst any) error {
	rc = http.MaxBytesReader(w, rc, maxJSONSize)
	dec := json.NewDecoder(rc)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		se, malformed := errors.AsType[*json.SyntaxError](err)
		ute, wrongType := errors.AsType[*json.UnmarshalTypeError](err)
		_, invalidDest := errors.AsType[*json.InvalidUnmarshalError](err)
		mbe, oversized := errors.AsType[*http.MaxBytesError](err)

		switch {
		case malformed:
			return fmt.Errorf("JSON is malformed at byte number %v", se.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("JSON is malformed")
		case wrongType: // this could happen if JSON number value overflows or underflows Go numeric type.
			if ute.Field != "" {
				return fmt.Errorf("wrong JSON type for %q", ute.Field)
			}
			return fmt.Errorf("wrong JSON type at byte number %v", ute.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			k := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("JSON contains unknown key %v", k)
		case oversized:
			return fmt.Errorf("body must not be larger than %v bytes", mbe.Limit)
		case errors.Is(err, data.ErrInvalidRuntimeFormat):
			return errors.New(strings.TrimPrefix(err.Error(), "data: "))
		case invalidDest:
			// this is developer's error, so panic instead of returning error.
			panic(fmt.Sprintf("decodeJSON: decode destination is not a non-nil pointer: %s", err)) // err instead of iue cause iue is only a node in err's tree.
		default:
			return err
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// parseString either returns the value corresponded wih key as a string, or returns defaultVal if key is not found.
func parseString(query url.Values, key, defaultVal string) string {
	s := query.Get(key)
	if s == "" {
		return defaultVal
	}

	return s
}

// parseCSV either returns the csv corresponded wih key as a []string, or returns defaultVal if key is not found.
func parseCSV(query url.Values, key string, defaultVal []string) []string {
	csv := query.Get(key)
	if csv == "" {
		return defaultVal
	}

	return strings.Split(csv, ",")
}

// parseInt parses the value corresponded with key to int.
// defaultVal is returned if key is not found.
// An error is added to v if parsing fails.
func parseInt(query url.Values, key string, defaultVal int, v *validator.Validator) int {
	s := query.Get(key)
	if s == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddErr(key, "must be sensible integer")
		return defaultVal
	}

	return i
}

// do defers recovering a potential panic, then calls f().
func (a *app) do(f func()) {
	defer func() {
		if err := recover(); err != nil {
			a.log.Error(fmt.Sprintf("%v", err))
		}
	}()

	f()
}

// godo starts a new goroutine to call a.do(f).
func (a *app) godo(f func()) {
	go a.do(f)
}

// wgo increments a.wg by 1, then starts a new goroutine to: defer a.wg.Done(), then calls a.do(f).
func (a *app) wgo(f func()) {
	a.wg.Go(func() { a.do(f) })
}
