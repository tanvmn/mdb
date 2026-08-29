package data

import (
	"database/sql"
	"slices"

	"github.com/lib/pq"
)

const (
	PermCodeMoviesRead  = "movies:read"
	PermCodeMoviesWrite = "movies:write"
)

type Perm struct {
	ID   int64
	Code string
}

func (p *Perm) dest() []any {
	return []any{
		&p.ID,
		&p.Code,
	}
}

type Perms []*Perm

func (ps Perms) Include(code string) bool {
	return slices.ContainsFunc(ps, func(p *Perm) bool { return p.Code == code })
}

type PermRepo struct {
	db *sql.DB
}

func (r *PermRepo) GetAllForUser(userID int64) (Perms, error) {
	q := `select count(*) over(), p.id, p.code
		from users_permissions as up
		inner join permissions as p on p.id = up.permission_id
		inner join users as u on u.id = up.user_id
		where u.id=$1
	`
	args := []any{userID}
	newPerm := func() *Perm { return &Perm{} }
	perms, _, err := query(r.db, q, args, newPerm)

	return perms, err
}

func (r *PermRepo) AddForUser(userID int64, codes ...string) error {
	q := `insert into users_permissions
		select $1, id from permissions where code=any($2)
	`
	args := []any{userID, pq.Array(codes)}
	_, err := result(r.db, q, args)

	return err
}
