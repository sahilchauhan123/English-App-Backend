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
	DROP TABLE IF EXISTS users;

		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			full_name TEXT NOT NULL,
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			age TEXT NOT NULL,
			gender TEXT NOT NULL,
			interests TEXT[] NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			profile_pic TEXT NOT NULL,
			password TEXT NOT NULL,
			auth_type TEXT NOT NULL
		);`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to  create users table: %v", err)
	}

	createTableQuery = ` 
	DROP TABLE IF EXISTS refresh_tokens;

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INT PRIMARY KEY,
		refresh_token TEXT NOT NULL
	);`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create refersh_tokens table: %v", err)
	}
	createTableQuery = `
	DROP TABLE IF EXISTS call_sessions;
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
	// This function should check if the user exists in the database.
	// If the user exists, return userDetailsWithJWTtoken, otherwise return false.
	var user types.User
	fmt.Println("Checking user in database:", email)
	checkQuery := `SELECT id,full_name, username, email, age,gender,interests, profile_pic , created_at,password,auth_type FROM users WHERE email = $1;`
	err := p.Db.QueryRow(context.Background(), checkQuery, email).Scan(&user.Id, &user.FullName, &user.Username, &user.Email, &user.Age, &user.Gender, &user.Interests, &user.ProfilePic, &user.CreatedAt, &user.Password, &user.AuthType)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("User not found in database")
			return false, user, nil // User not found
		}
		fmt.Println("im in errorrrrrrrr", err)
		return false, user, nil
	}

	return true, user, nil
}

func (p *PostgreSQL) CheckUsernameIsAvailable(username string) bool {
	checkQuery := `SELECT email FROM users WHERE username = $1;`
	var output string
	p.Db.QueryRow(context.Background(), checkQuery, username).Scan(&output)

	if output == "" {
		fmt.Println("Username is available:", username)
		return true // Username is available
	}

	return false

}

func (p *PostgreSQL) SaveUserInDatabase(user *types.User) error {
	// This function should save the user in the database.
	// If the user is saved successfully, return nil, otherwise return an error.

	fmt.Println("Saving user in database:", user)
	hashedPassword, err := authservice.HashPassword(user.Password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword

	insertQuery := `INSERT INTO users (full_name, username, email, age,gender,interests, profile_pic,password, auth_type)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at;`
	// var id int
	// var created_at pgtype.Timestamptz

	err = p.Db.QueryRow(context.Background(), insertQuery, user.FullName, user.Username, user.Email, user.Age, user.Gender, user.Interests, user.ProfilePic, hashedPassword, user.AuthType).Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		fmt.Println("Error inserting user:", err)
		return fmt.Errorf("failed to insert user: %v", err)
	}
	// user.CreatedAt = created_at // Convert pgtype.Timestamptz to string
	fmt.Println("User saved successfully with ID:", user.Id)

	// Here you would typically execute an INSERT statement to save the user.
	// For now, we will just return nil to indicate success.

	return nil
}

func (p *PostgreSQL) StoreToken(user types.User, token string) error {
	query := `INSERT INTO refresh_tokens (id, refresh_token) VALUES ($1, $2);`
	_, err := p.Db.Exec(context.Background(), query, user.Id, token)
	if err != nil {
		fmt.Println("Error storing token:", err)
		return fmt.Errorf("failed to store token: %v", err)
	}
	return nil
}

func (p *PostgreSQL) DeleteToken(userid int64) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1;`
	_, err := p.Db.Exec(context.Background(), query, userid)
	if err != nil {
		// fmt.Println("Error deleting token:", err)
		// return fmt.Errorf("failed to delete token: %v", err)
		fmt.Println("token not found for user ID:", userid)
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
		fmt.Println("Error changing password:", err)
		return fmt.Errorf("failed to change password: %v", err)
	}
	fmt.Println("Password changed successfully for email:", email)
	return nil
}

func (p *PostgreSQL) StartCall(peer1, peer2 int64) (string, error) {
	var id string
	query := `INSERT INTO call_sessions (peer1 , peer2) VALUES ($1,$2) RETURNING id;`
	err := p.Db.QueryRow(context.Background(), query, peer1, peer2).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error in Insert Call ", err.Error())
	}
	return id, nil
}

// func (p *PostgreSQL) EndCall(peer1, peer2 string) error {
// 	// var id string
// 	// // // query := `UPDATE INTO call_sessions (peer1 , peer2) VALUES ($1,$2) RETURNING id;`
// 	// // err := p.Db.QueryRow(context.Background(), query, peer1, peer2).Scan(&id)
// 	// // if err != nil {
// 	// // 	return fmt.Errorf("error in Insert Call ", err.Error())
// 	// // }
// 	// return nil
// }
