package postgresql

import (
	"context"
	"fmt"
	authservice "github/english-app/internal/auth/service"
	"github/english-app/internal/types"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQL struct {
	Db *pgxpool.Pool
}

func New() (*PostgreSQL, error) {
	// This function should initialize the PostgreSQL connection.
	// For now, we will return a dummy PostgreSQL instance and nil error.
	// i will pgx import "github.com/jackc/pgx/v5"
	// and use it to connect to the database.
	// Replace the nil with actual DB connection logic.
	// Example:
	connstr := os.Getenv("POSTGRES_URL")
	conn, err := pgxpool.New(context.Background(), connstr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	createTableQuery := `

		CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		full_name TEXT NOT NULL,
		username TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		age INT NOT NULL,
		gender TEXT NOT NULL,
		profile_pic TEXT NOT NULL,
		auth_type TEXT NOT NULL,
		main_challenge TEXT NOT NULL,
		native_language TEXT NOT NULL,
		current_english_level TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
		);
`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to  create users table: %v", err)
	}

	createTableQuery = ` 

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INT PRIMARY KEY,
		refresh_token TEXT NOT NULL
	);`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create refersh_tokens table: %v", err)
	}
	createTableQuery = `
	CREATE TABLE IF NOT EXISTS call_sessions (
    	id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- random call ID
    	peer1_id BIGINT NOT NULL,
    	peer2_id BIGINT NOT NULL,
    	started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    	ended_at TIMESTAMPTZ, -- Made nullable, remove DEFAULT now()
    	status TEXT DEFAULT 'ongoing'  -- optional: 'ended', etc.
	);`
	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create call_sessions table: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	// db := &PostgreSQL{
	// 	Db: conn, // Replace with actual DB connection
	// }
	fmt.Println("Connected to PostgreSQL database successfully")
	return &PostgreSQL{Db: conn}, nil

}
func (p *PostgreSQL) CheckUserInDatabase(email string) (bool, types.User, error) {
	var user types.User
	fmt.Println("Checking user in database:", email)

	checkQuery := `SELECT 
		id, full_name, username, email, age, gender, profile_pic, 
		created_at, password, auth_type, main_challenge, native_language, current_english_level 
		FROM users WHERE email = $1;`

	err := p.Db.QueryRow(context.Background(), checkQuery, email).Scan(
		&user.Id,
		&user.FullName,
		&user.Username,
		&user.Email,
		&user.Age,
		&user.Gender,
		&user.ProfilePic,
		&user.CreatedAt,
		&user.Password,
		&user.AuthType,
		&user.MainChallenge,
		&user.NativeLanguage,
		&user.CurrentEnglishLevel,
	)

	if err == pgx.ErrNoRows {
		fmt.Println("User not found in database")
		return false, user, nil
	} else if err != nil {
		return false, user, fmt.Errorf("database error: %v", err)
	}

	return true, user, nil
}

func (p *PostgreSQL) CheckUsernameIsAvailable(username string) bool {
	checkQuery := `SELECT email FROM users WHERE username = $1;`
	var output string
	err := p.Db.QueryRow(context.Background(), checkQuery, username).Scan(&output)
	if err == pgx.ErrNoRows {
		return true
	}
	return false
}

func (p *PostgreSQL) SaveUserInDatabase(user *types.User) error {
	fmt.Println("Saving user in database:", user)
	hashedPassword, err := authservice.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	insertQuery := `INSERT INTO users (
		full_name, username, email, age, gender, profile_pic, password,
		auth_type, main_challenge, native_language, current_english_level
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11
	) RETURNING id, created_at;`

	err = p.Db.QueryRow(context.Background(), insertQuery,
		user.FullName,
		user.Username,
		user.Email,
		user.Age,
		user.Gender,
		user.ProfilePic,
		user.Password,
		user.AuthType,
		user.MainChallenge,
		user.NativeLanguage,
		user.CurrentEnglishLevel,
	).Scan(&user.Id, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}
	return nil
}

func (p *PostgreSQL) StoreToken(user types.User, token string) error {
	query := `
	INSERT INTO refresh_tokens (id, refresh_token)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET refresh_token = EXCLUDED.refresh_token;`

	_, err := p.Db.Exec(context.Background(), query, user.Id, token)
	if err != nil {
		return fmt.Errorf("failed to store token: %v", err)
	}
	return nil
}

func (p *PostgreSQL) DeleteToken(userid int64) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1;`
	_, err := p.Db.Exec(context.Background(), query, userid)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to delete token: %v", err)
	}
	return nil
}

func (p *PostgreSQL) ChangePassword(email string, newPassword string) error {
	password, err := authservice.HashPassword(newPassword)
	if err != nil {
		return err
	}
	query := `UPDATE users SET password = $1 WHERE email = $2;`
	_, err = p.Db.Exec(context.Background(), query, password, email)
	if err != nil {
		return fmt.Errorf("failed to change password: %v", err)
	}
	return nil
}

func (p *PostgreSQL) StartCall(peer1, peer2 int64) (string, error) {
	var id string
	query := `INSERT INTO call_sessions (peer1, peer2) VALUES ($1, $2) RETURNING id;`
	err := p.Db.QueryRow(context.Background(), query, peer1, peer2).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error starting call: %v", err)
	}
	return id, nil
}

func (p *PostgreSQL) CheckToken(token string) (bool, int64) {
	var id int64
	query := `SELECT id FROM refresh_tokens WHERE refresh_token = $1;`
	err := p.Db.QueryRow(context.Background(), query, token).Scan(&id)
	if err == pgx.ErrNoRows {
		return false, 0
	} else if err != nil {
		return false, 0
	}
	return true, id
}
