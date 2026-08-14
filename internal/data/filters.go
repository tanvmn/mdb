package data

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tanvmn/mdb/internal/validator"
)

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string
}

func (f *Filters) sortIsValid() bool {
	return slices.Contains(f.SortList, f.Sort)
}

func ValidateFilters(v *validator.Validator, f *Filters) {
	v.Validate(f.Page > 0, "page", "must be more than 0")
	v.Validate(f.Page <= 10_000_000, "page", "must not be more than 10 millions")

	v.Validate(f.PageSize > 0, "page_size", "must be more than 0")
	v.Validate(f.PageSize <= 100, "page_size", "must not be more than 100")

	v.Validate(f.sortIsValid(), "sort", "invalid")
}

// sortColumn returns the valid sort column without the potential prefix "-".
// If f.Sort isn't valid, sortColumn panics.
func (f *Filters) sortColumn() string {
	if f.sortIsValid() {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("invalid sort parameter:" + f.Sort)
}

// sortDirection returns either DESC or ASC.
func (f *Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f *Filters) orderBy() string {
	return fmt.Sprintf("order by %v %v, id asc", f.sortColumn(), f.sortDirection())
}
