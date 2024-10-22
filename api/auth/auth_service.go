package auth

import (
	"context"
	"database/sql"
	"errors"
	"simple-connect/api/data/gen/gopg/public/model"
	. "simple-connect/api/data/gen/gopg/public/table"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (*DBUser, error)
	GetUserByID(ctx context.Context, id int64) (*model.Users, error)
	CreateUser(ctx context.Context, email, password string) (string, error)
	GetUserAccount(ctx context.Context, email string, provider string) (UserAccount, error)
	UpdateAccountTokens(ctx context.Context, userId int64, provider string, providerId string, data UpdateAccountTokensData) error
	CreateAccount(ctx context.Context, data CreateAccountData) (int64, error)
}

type AuthService struct {
	pool *pgxpool.Pool
	db   *sql.DB
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	db := stdlib.OpenDBFromPool(pool)
	return &AuthService{pool: pool, db: db}
}

type DBUser struct {
	ID            int64          `db:"id" json:"id"`
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

func (as *AuthService) GetUserByID(ctx context.Context, id int64) (*model.Users, error) {

	stmt := SELECT(Users.AllColumns.Except(Users.PasswordHash)).FROM(Users).WHERE(Users.ID.EQ(Int(id))).LIMIT(1)

	var user model.Users

	err := stmt.Query(as.db, &user)

	return &user, err
}

func (as *AuthService) CreateUser(ctx context.Context, email, password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	row := as.pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, string(hashed))

	var id string

	err = row.Scan(&id)

	return id, err
}

type UserAccount struct {
	Email      string `db:"email"`
	Provider   string `db:"provider"`
	ProviderId string `db:"provider_id"`
	UserId     int64  `db:"user_id"`
}

func (as *AuthService) GetUserAccount(ctx context.Context, email string, provider string) (UserAccount, error) {

	rows, err := as.pool.Query(ctx,
		"SELECT u.email, a.provider, a.provider_id, a.user_id FROM users u LEFT JOIN accounts a ON u.id = a.user_id WHERE u.email = $1 AND a.provider = $2 LIMIT 1",
		email, provider)

	if err != nil {
		return UserAccount{}, err
	}

	data, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[UserAccount])

	return data, err
}

type UpdateAccountTokensData struct {
	AccessToken       string       `db:"access_token"`
	RefreshToken      string       `db:"refresh_token"`
	AccessTokenExpiry sql.NullTime `db:"access_token_expires_at"`
}

func (as *AuthService) UpdateAccountTokens(ctx context.Context, userId int64, provider string, providerId string, data UpdateAccountTokensData) error {
	tx, err := as.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx,
		`UPDATE accounts SET
			access_token = @access_token,
			refresh_token = @refresh_token,
			access_token_expires_at = @access_token_expires_at
			WHERE user_id = @user_id AND provider = @provider AND provider_id = @provider_id
			`, pgx.NamedArgs{
			"access_token":            data.AccessToken,
			"refresh_token":           data.RefreshToken,
			"access_token_expires_at": data.AccessTokenExpiry,
			"user_id":                 userId,
			"provider":                provider,
			"provider_id":             providerId,
		},
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() != 1 {
		return errors.New("more than one row affected")
	}

	err = tx.Commit(ctx)

	return err
}

type CreateAccountData struct {
	Email             string
	Provider          string
	ProviderId        string       `db:"provider_id"`
	AccessToken       string       `db:"access_token"`
	RefreshToken      string       `db:"refresh_token"`
	AccessTokenExpiry sql.NullTime `db:"access_token_expires_at"`
}

func (as *AuthService) CreateAccount(ctx context.Context, data CreateAccountData) (int64, error) {

	tx, err := as.pool.Begin(ctx)

	if err != nil {
		return 0, err
	}

	defer tx.Rollback(ctx)

	var userId int64
	userRow := tx.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", data.Email)

	err = userRow.Scan(&userId)

	if err != nil {
		return 0, err
	}

	accountRow := tx.QueryRow(ctx,
		"INSERT into accounts (provider, provider_id, user_id, access_token, refresh_token, access_token_expires_at) VALUES (@provider, @provider_id, @user_id, @access_token, @refresh_token, @access_token_expires_at) RETURNING provider_id",
		pgx.NamedArgs{
			"provider":                data.Provider,
			"provider_id":             data.ProviderId,
			"user_id":                 userId,
			"access_token":            data.AccessToken,
			"refresh_token":           data.RefreshToken,
			"access_token_expires_at": data.AccessTokenExpiry,
		},
	)

	var accountId string

	err = accountRow.Scan(&accountId)

	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)

	return userId, err
}
