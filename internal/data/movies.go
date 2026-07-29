package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // don't embed Runtime, cause (*)Movie will satisfies json.Marshaler, then JSON of (*)Movie will be JSON of (*)Runtime.
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}
