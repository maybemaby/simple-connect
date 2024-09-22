package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (*DBUser, error)
	CreateUser(ctx context.Context, email, password string) (string, error)
}

type AuthService struct {
	pool *pgxpool.Pool
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}

type DBUser struct {
	ID            string         `db:"id" json:"id"`
	PasswordHash  sql.NullString `db:"password_hash" json:"password_hash"`
	Email         *string        `db:"email" json:"email"`
	EmailVerified *time.Time     `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

type DBSession struct {
	Token     string    `db:"token" json:"token"`
	UserID    string    `db:"user_id" json:"user_id"`
	Expiry    time.Time `json:"expiry"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

const getUserByEmailQuery = `
SELECT * FROM users WHERE email = $1 LIMIT 1
`

func (as *AuthService) GetUserByEmail(ctx context.Context, email string) (*DBUser, error) {

	rows, err := as.pool.Query(ctx, getUserByEmailQuery, email)

	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[DBUser])

	return user, err
}

func (as *AuthService) CreateUser(ctx context.Context, email, password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	row := as.pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, string(hashed))

	if err != nil {
		return "", err
	}

	var id string

	err = row.Scan(&id)

	return id, err
}
