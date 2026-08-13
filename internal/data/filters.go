package data

import (
	"slices"

	"github.com/tanvmn/mdb/internal/validator"
)

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string
}

func ValidateFilters(v *validator.Validator, f *Filters) {
	v.Validate(f.Page > 0, "page", "must be more than 0")
	v.Validate(f.Page <= 10_000_000, "page", "must not be more than 10 millions")

	v.Validate(f.PageSize > 0, "page_size", "must be more than 0")
	v.Validate(f.PageSize <= 100, "page_size", "must not be more than 100")

	v.Validate(slices.Contains(f.SortList, f.Sort), "sort", "invalid sort value")
}
