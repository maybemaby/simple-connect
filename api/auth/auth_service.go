package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthStore interface{}

type AuthService struct {
	pool *pgxpool.Pool
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{pool: pool}
}


