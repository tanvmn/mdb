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

// UnmarshalJSON returns error if jsonStr isn't of format "<int32> mins" (double quotes included).
func (r *Runtime) UnmarshalJSON(jsonStr []byte) error {
	s, err := strconv.Unquote(string(jsonStr))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(s, " ")
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
