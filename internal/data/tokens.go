package data

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/tanvmn/mdb/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"

	columnsOfTokensAsT = " t.hash, t.user_id, t.expiry, t.scope "
	selectTokens       = "select " + columnsOfTokensAsT + " from tokens as t "
)

type Token struct {
	Plaintext string    `json:"token"` // 128-bit randomness, base32 encoded
	Hash      []byte    `json:"-"`     // SHA256 checksum
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func (t *Token) dest() []any {
	return []any{
		&t.Hash,
		&t.UserID,
		&t.Expiry,
		&t.Scope,
	}
}

func token(userID int64, lifetime time.Duration, scope string) *Token {
	t := &Token{
		Plaintext: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(lifetime),
		Scope:     scope,
	}
	// sha256 instead of bcrypt, cause t.Plaintext has 128-bit randomness and is base32 encoded
	hash := sha256.Sum256([]byte(t.Plaintext))
	t.Hash = hash[:]

	return t
}

// ValidateToken validates if plaintext is exactly 26 bytes long.
func ValidateToken(v *validator.Validator, plaintext string) {
	v.Validate(len(plaintext) == 26, "token", "must be 26 bytes long")
}

type TokenRepo struct {
	db *sql.DB
}

// Add adds t to tokens repo.
func (r *TokenRepo) Add(t *Token) error {
	q := "insert into tokens (hash, user_id, expiry, scope) values ($1,$2,$3,$4) "
	args := []any{t.Hash, t.UserID, t.Expiry, t.Scope}
	_, err := result(r.db, q, args)

	return err
}

// New creates a new token, adds it to tokens repo and returns it.
func (r *TokenRepo) New(userID int64, lifetime time.Duration, scope string) (*Token, error) {
	t := token(userID, lifetime, scope)
	err := r.Add(t)

	return t, err
}

// DeleteAllForUser deletes all tokens corresponded with userID and scope.
func (r *TokenRepo) DeleteAllForUser(scope string, userID int64) error {
	q := "delete from tokens where scope=$1 and user_id=$2 "
	args := []any{scope, userID}
	_, err := result(r.db, q, args)

	return err
}
