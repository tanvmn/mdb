package validator

import "regexp"

// EmailRX is an email regular expression from https://html.spec.whatwg.org/#valid-e-mail-address
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string // maps key to error message.
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true of v has no errors.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddErr adds an error to v.Errors map if key doesn't exist.
func (v *Validator) AddErr(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Validate calls v.AddErr(key,message) if !ok.
func (v *Validator) Validate(ok bool, key, message string) {
	if !ok {
		v.AddErr(key, message)
	}
}

// Matches returns true if s contains any match of regular expression rx.
func Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// Unique returns true if all of vs are unique or vs == nil.
func Unique[T comparable](vs []T) bool {
	uniques := make(map[T]struct{})
	for _, v := range vs {
		uniques[v] = struct{}{}
	}

	return len(uniques) == len(vs)
}
