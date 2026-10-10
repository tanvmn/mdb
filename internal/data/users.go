package data

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/tanvmn/mdb/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	selectUsers   = "select u.id, u.created_at, u.name, u.email, u.password_hash, u.activated, u.version from users as u "
	returningUser = " returning id, created_at, name, email, password_hash, activated, version "
)

var (
	ErrDuplicateEmail     = errors.New("data: duplicate email")
	ErrMismatchedPassword = errors.New("data: mismatched password")
	ErrNotHavePerm        = errors.New("data: user doesn't have the permission")

	AnonymousUser = &User{}
)

type password struct {
	plaintext *string
	hash      []byte
}

// Set stores plaintext and its bcrypt hash in p.
func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

// Matches returns nil if plaintext matches p.hash.
// Matches could return ErrMismatchedPassword or another error occured while comparing plaintext to its hash.
func (p *password) Matches(plaintext string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrMismatchedPassword
		}
		return err
	}

	return nil
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (u *User) dest() []any {
	return []any{
		&u.ID,
		&u.CreatedAt,
		&u.Name,
		&u.Email,
		&u.Password.hash,
		&u.Activated,
		&u.Version,
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Validate(email != "", "email", "must not be empty")
	v.Validate(validator.Matches(email, validator.EmailRX), "email", "invalid format")
}

// ValidatePassword validates if plaintext's length is between 8 and 72 bytes.
func ValidatePassword(v *validator.Validator, plaintext string) {
	l := len(plaintext)
	v.Validate(l >= 8, "password", "must be at least 8 bytes")
	v.Validate(l <= 72, "password", "must not be larger than 72 bytes")
}

func ValidateUser(v *validator.Validator, u *User) {
	v.Validate(u.Name != "", "name", "must not be empty")
	v.Validate(len(u.Name) <= 500, "name", "must not be larger than 500 bytes")
	ValidateEmail(v, u.Email)

	if u.Password.plaintext != nil {
		ValidatePassword(v, *u.Password.plaintext)
	}
	if u.Password.hash == nil {
		panic("missing user's password hash")
	}
}

type UserRepo struct {
	db *sql.DB
}

// read queries an user and reads that into u.
// ErrRecordNotFound is returned if no user is found.
func (r *UserRepo) read(query string, args []any, u *User) error {
	return queryRow(r.db, query, args, u.dest())
}

// get queries an user.
// ErrRecordNotFound is returned if no user is found.
func (r *UserRepo) get(query string, args []any) (*User, error) {
	u := &User{}
	err := r.read(query, args, u)

	return u, err
}

func emailIsDuplicate(err error) bool {
	s := err.Error()

	return strings.Contains(s, "duplicate key") && strings.Contains(s, `"users_email_key"`)
}

func (r *UserRepo) Add(u *User) error {
	q := "insert into users (name, email, password_hash, activated) values ($1,$2,$3,$4) " + returningUser
	args := []any{u.Name, u.Email, u.Password.hash, u.Activated}

	if err := r.read(q, args, u); err != nil {
		if emailIsDuplicate(err) {
			return ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (r *UserRepo) GetByEmail(email string) (*User, error) {
	q := selectUsers + " where email=$1 "
	args := []any{email}

	return r.get(q, args)
}

func (r *UserRepo) Set(u *User) error {
	q := "update users set name=$1, email=$2, password_hash=$3, activated=$4, version=version+1 where version=$5 and id=$6 " + returningUser
	args := []any{u.Name, u.Email, u.Password.hash, u.Activated, u.Version, u.ID}

	if err := r.read(q, args, u); err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			return ErrEditConflict
		case emailIsDuplicate(err):
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByToken returns an user associated with an existing, unexpired token with scope.
func (r *UserRepo) GetByToken(scope, plaintext string) (*User, error) {
	// postgres doesn't compute t.hash=$1 in constant time,
	// so attacker can send thousands of requests, track the response times
	// and analyze the differences to get the estimate of the hash value.
	// But even if the hash value is found, attackers still have to guess the plaintext token
	// in order to use the API, and that's is infeasible with current tech.
	q := selectUsers + `
		inner join tokens as t
		on t.user_id = u.id
		where t.hash=$1 --<-- vulnerable
		and t.scope=$2
		and t.expiry > $3
	`
	hash := sha256.Sum256([]byte(plaintext))
	// note the [:], cause pq doesn't support converting Go array.
	args := []any{hash[:], scope, time.Now()}

	return r.get(q, args)
}

// HasPerm returns nil if user with id has perm,
// returns ErrUserNotHavePerm if user doesn't,
// else returns the error occured while computing.
func (r *UserRepo) HasPerm(id int64, perm string) error {
	q := `select exists (
		select true
		from users_permissions as up
		inner join users as u on u.id = up.user_id
		inner join permissions as p on p.id = up.permission_id
		where u.id=$1 and p.code=$2
	)
	`
	args := []any{id, perm}
	var has bool
	if err := queryRow(r.db, q, args, []any{&has}); err != nil {
		return err
	}
	if !has {
		return ErrNotHavePerm
	}

	return nil
}
