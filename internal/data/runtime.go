package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r *Runtime) MarshalJSON() ([]byte, error) {
	jsonStr := strconv.Quote(fmt.Sprintf("%v mins", *r))

	return []byte(jsonStr), nil
}
