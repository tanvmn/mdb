package data

import (
	"fmt"
	"time"

	"github.com/tanvmn/mdb/internal/validator"
)

const MinMovieYear = 1888

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // don't embed Runtime, cause (*)Movie will satisfies json.Marshaler, then JSON of (*)Movie will be JSON of (*)Runtime.
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, m *Movie) {
	v.Validate(m.Title != "", "title", "must be provided and not an empty string")
	v.Validate(len(m.Title) <= 500, "title", "must not be longer than 500 bytes")

	v.Validate(m.Year != 0, "year", fmt.Sprintf("must be provided and not before %v", MinMovieYear))
	v.Validate(m.Year >= MinMovieYear, "year", fmt.Sprintf("must not be before %v", MinMovieYear))
	v.Validate(m.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Validate(m.Runtime != 0, "runtime", "must be provided and a positive integer")
	v.Validate(m.Runtime > 0, "runtime", "must be a positive integer")

	genresCount := len(m.Genres)
	v.Validate(m.Genres != nil, "genres", "must be provided")
	v.Validate(genresCount >= 1, "genres", "must have at least 1")
	v.Validate(genresCount <= 5, "genres", "must not have more than 5")
	v.Validate(validator.Unique(m.Genres), "genres", "must not have duplicates")
}
