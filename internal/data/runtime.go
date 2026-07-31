package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("data: invalid runtime format")

type Runtime int32

func (r *Runtime) MarshalJSON() ([]byte, error) {
	jsonStr := strconv.Quote(fmt.Sprintf("%v mins", *r))

	return []byte(jsonStr), nil
}

func (r *Runtime) UnmarshalJSON(jsonVal []byte) error {
	unquoted, err := strconv.Unquote(string(jsonVal))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(unquoted, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(i)

	return nil
}
