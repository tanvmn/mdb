package data

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tanvmn/mdb/internal/validator"
)

type Filter struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string
}

func (f *Filter) sortValid() bool {
	return slices.Contains(f.SortList, f.Sort)
}

func ValidateFilter(v *validator.Validator, f *Filter) {
	v.Validate(f.Page > 0, "page", "must be more than 0")
	v.Validate(f.Page <= 10_000_000, "page", "must not be more than 10 millions")

	v.Validate(f.PageSize > 0, "page_size", "must be more than 0")
	v.Validate(f.PageSize <= 100, "page_size", "must not be more than 100")

	v.Validate(f.sortValid(), "sort", "invalid")
}

// sortColumn returns the valid sort column without the potential prefix "-".
// If f.Sort isn't valid, sortColumn panics.
func (f *Filter) sortColumn() string {
	if f.sortValid() {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("invalid sort parameter: " + f.Sort)
}

// sortDirection returns either DESC or ASC.
func (f *Filter) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// orderBy returns the SQL's ORDER BY clause.
func (f *Filter) orderBy() string {
	return fmt.Sprintf("order by %v %v, id asc ", f.sortColumn(), f.sortDirection())
}

// limit returns the max number of records to be returned by db.
func (f *Filter) limit() int {
	return f.PageSize
}

// offset tells db the number of records to skip before returning records.
// The risk of overflowing the return value is mitigated by setting f.Page & f.PageSize's limits in ValidateFilter func.
func (f *Filter) offset() int {
	return f.PageSize * (f.Page - 1)
}

// Metadata contains metadata about the returned records.
type Metadata struct {
	CurrentPage  int `json:"currentPage,omitzero"`
	PageSize     int `json:"pageSize,omitzero"`
	LastPage     int `json:"lastPage,omitzero"`
	TotalRecords int `json:"totalRecords,omitzero"`
}

func metadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}
