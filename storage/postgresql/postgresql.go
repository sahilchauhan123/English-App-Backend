package postgresql

import (
	"context"
	"fmt"
	authservice "github/english-app/internal/auth/service"
	"github/english-app/internal/types"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		created_at TIMESTAMPTZ DEFAULT NOW(),
    	pictures TEXT[] DEFAULT '{}',
		is_active BOOl DEFAULT TRUE
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
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		peer1_id BIGINT NOT NULL,
		peer2_id BIGINT NOT NULL,
		peer1_name TEXT NOT NULL,
		peer2_name TEXT NOT NULL,
		peer1_pic TEXT,
		peer2_pic TEXT,
		started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		ended_at TIMESTAMPTZ,
		status TEXT DEFAULT 'ongoing',
		FOREIGN KEY (peer1_id) REFERENCES users(id),
		FOREIGN KEY (peer2_id) REFERENCES users(id)
	);`
	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create call_sessions table: %v", err)
	}

	createTableQuery = `
	CREATE TABLE IF NOT EXISTS blocks (
		user_id INT NOT NULL,
		blocked_by INT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (user_id, blocked_by),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (blocked_by) REFERENCES users(id)
	);`
	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create blocks table: %v", err)
	}

	createTableQuery = `
		CREATE TABLE IF NOT EXISTS feedback (
		id SERIAL PRIMARY KEY,
		feedback_to INT NOT NULL,
		feedback_from INT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		comments TEXT,
		rating INT,
		callID UUID,
		FOREIGN KEY (feedback_to) REFERENCES users(id),
		FOREIGN KEY (feedback_from) REFERENCES users(id)
	);`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create Feedback table: %v", err)
	}

	createTableQuery = `
		CREATE TABLE IF NOT EXISTS leaderboard (
			user_id BIGINT NOT NULL,
			period_type TEXT NOT NULL,       -- 'daily', 'weekly', 'monthly', or 'alltime'
			period_start DATE NOT NULL,      -- date representing the start of that period
			total_duration BIGINT DEFAULT 0, -- total call duration in seconds
			updated_at TIMESTAMPTZ DEFAULT now(),
			PRIMARY KEY (user_id, period_type, period_start)
		);
		`

	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create leaderboard table: %v", err)
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
		created_at, password, auth_type, main_challenge, native_language, current_english_level,pictures
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
		&user.Pictures,
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

	var output string
	checkQuery := `SELECT email FROM users WHERE username = $1;`
	username = strings.TrimSpace(strings.ToLower(username))

	err := p.Db.QueryRow(context.Background(), checkQuery, username).Scan(&output)
	fmt.Println("Username check query result:", output, "Error:", err)
	if err == pgx.ErrNoRows {
		fmt.Println("Username is available:", username)
		return true
	}
	fmt.Println("Username is taken:", username)
	return false
}

func (p *PostgreSQL) SaveUserInDatabase(user *types.User) error {
	fmt.Println("Saving user in database:", user)
	hashedPassword, err := authservice.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = hashedPassword
	user.Username = strings.TrimSpace(strings.ToLower(user.Username))

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

func (p *PostgreSQL) StartCall(peer1, peer2 types.User) (string, error) {
	var id string
	query := `
        INSERT INTO call_sessions (
            peer1_id, peer2_id, 
            peer1_name, peer2_name,
            peer1_pic, peer2_pic
        ) VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id;`

	err := p.Db.QueryRow(
		context.Background(),
		query,
		peer1.Id, peer2.Id,
		peer1.FullName, peer2.FullName,
		peer1.ProfilePic, peer2.ProfilePic,
	).Scan(&id)

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

func (p *PostgreSQL) EndCall(id string) error {
	var durationSeconds float64
	var peer1ID, peer2ID int64

	query := `
		UPDATE call_sessions
		SET status = 'ended', ended_at = NOW()
		WHERE id = $1
		RETURNING 
			peer1_id,
			peer2_id,
			EXTRACT(EPOCH FROM (ended_at - started_at)) AS duration_seconds;`

	err := p.Db.QueryRow(context.Background(), query, id).Scan(&peer1ID, &peer2ID, &durationSeconds)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil // No ongoing call to end, so just return nil
		}
		return fmt.Errorf("error ending call: %v", err)
	}
	// Update leaderboard for both peers
	err = p.UpdateLeaderboard(peer1ID, durationSeconds)
	if err != nil {
		return fmt.Errorf("error updating leaderboard for peer1: %v", err)
	}
	err = p.UpdateLeaderboard(peer2ID, durationSeconds)
	if err != nil {
		return fmt.Errorf("error updating leaderboard for peer2: %v", err)
	}
	return nil
}

func (p *PostgreSQL) InsertPicture(id int64, imageUrl string) error {
	query := `UPDATE users SET pictures = array_append(pictures, $1) WHERE id = $2;`
	_, err := p.Db.Exec(context.Background(), query, imageUrl, id)
	if err != nil {
		return fmt.Errorf("error inserting picture: %v", err)
	}
	return nil
}

func (p *PostgreSQL) CheckPictureLength(id int64) (int, error) {
	var pictures []string
	checkQuery := `SELECT pictures FROM users WHERE id = $1;`
	err := p.Db.QueryRow(context.Background(), checkQuery, id).Scan(&pictures)
	if err != nil {
		return 0, fmt.Errorf("error checking pictures length: %v", err)
	}

	fmt.Println("Pictures array length:", len(pictures))
	return len(pictures), nil
}

func (p *PostgreSQL) GetProfile(userID int64) (types.User, error) {
	var profile types.User
	query := `SELECT 
		id, full_name, username, email, age, gender, profile_pic, 
		created_at, main_challenge, native_language, current_english_level,pictures
		FROM users WHERE id = $1;`
	err := p.Db.QueryRow(context.Background(), query, userID).Scan(
		&profile.Id,
		&profile.FullName,
		&profile.Username,
		&profile.Email,
		&profile.Age,
		&profile.Gender,
		&profile.ProfilePic,
		&profile.CreatedAt,
		&profile.MainChallenge,
		&profile.NativeLanguage,
		&profile.CurrentEnglishLevel,
		&profile.Pictures,
	)
	if err != nil {
		return types.User{}, fmt.Errorf("error fetching profile: %v", err)
	}
	return profile, nil
}

func (p *PostgreSQL) GetCallHistory(userId int64, timestamp time.Time) ([]types.CallRecord, error) {
	var query string
	var rows pgx.Rows
	var err error

	if !timestamp.IsZero() {
		query = `SELECT * FROM call_sessions
		WHERE (peer1_id = $1 OR peer2_id = $1) AND started_at >= $2
		ORDER BY started_at DESC
		LIMIT 25;`
		fmt.Println("the timestamp is : ", timestamp)
		rows, err = p.Db.Query(context.Background(), query, userId, timestamp)
	} else {
		query = `SELECT * FROM call_sessions
		WHERE peer1_id = $1 OR peer2_id = $1
		ORDER BY started_at DESC
		LIMIT 25;`
		rows, err = p.Db.Query(context.Background(), query, userId)
	}

	if err != nil {
		return nil, fmt.Errorf("error fetching call history: %v", err)
	}

	var history []types.CallHistory
	for rows.Next() {
		var record types.CallHistory
		var endedAt pgtype.Timestamptz
		var startedAt pgtype.Timestamptz
		var durationSeconds float64

		err := rows.Scan(
			&record.CallId,
			&record.PeerID1,
			&record.PeerID2,
			&record.PeerName1,
			&record.PeerName2,
			&record.PeerPic1,
			&record.PeerPic2,
			&startedAt,
			&endedAt,
			&record.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning call history row: %v", err)
		}

		if endedAt.Valid {
			record.CallEnd = endedAt.Time.UTC().Format("2006-01-02 15:04:05")
			record.CallStart = startedAt.Time.UTC().Format("2006-01-02 15:04:05")
			durationSeconds = endedAt.Time.Sub(startedAt.Time).Seconds()
		} else {
			record.CallEnd = "ongoing"
		}

		minutes := int(durationSeconds) / 60
		seconds := int(durationSeconds) % 60
		record.DurationInMin = fmt.Sprintf("%02d:%02d", minutes, seconds)

		history = append(history, record)

	}

	formattedRecord := FormatCallRecord(history, userId)

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating call history rows: %v", rows.Err())
	}

	return formattedRecord, nil
}

func (p *PostgreSQL) DeleteAccount(id int64) error {
	query := `UPDATE users SET is_active = FALSE WHERE id = $1;`
	_, err := p.Db.Query(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("error fetching call history: %v", err)
	}
	return nil

}

func (p *PostgreSQL) DeletePicture(userId int64, imageUrl string) error {
	query := `UPDATE users SET pictures = array_remove(pictures, $1) WHERE id = $2;`
	_, err := p.Db.Exec(context.Background(), query, imageUrl, userId)
	if err != nil {
		return fmt.Errorf("error deleting picture: %v", err)
	}
	return nil
}

func (p *PostgreSQL) BlockUser(userId int64, blockUserId int64) error {

	query := `INSERT INTO blocks (user_id, blocked_by)
		VALUES ($1,$2)
		ON CONFLICT (user_id, blocked_by) DO NOTHING;
	`
	_, err := p.Db.Exec(context.Background(), query, userId, blockUserId)

	if err != nil {
		return err
	}

	return nil
}
func (p *PostgreSQL) GetLeaderboard(periodType string) ([]types.LeaderboardEntry, error) {
	// Decide the date truncation dynamically based on the period type
	var periodStartExpr string

	switch periodType {
	case "daily":
		periodStartExpr = "date_trunc('day', NOW())"
	case "weekly":
		periodStartExpr = "date_trunc('week', NOW())"
	case "monthly":
		periodStartExpr = "date_trunc('month', NOW())"
	case "alltime":
		// All-time uses a fixed period_start
		periodStartExpr = "date '1970-01-01'"
	default:
		return nil, fmt.Errorf("invalid period type: %s", periodType)
	}

	query := fmt.Sprintf(`
		SELECT 
			RANK() OVER (ORDER BY l.total_duration DESC) AS rank,
			u.id AS user_id,
			u.username,
			u.full_name,
			u.profile_pic,
			u.native_language,
			l.total_duration
		FROM leaderboard l
		JOIN users u ON u.id = l.user_id
		WHERE l.period_type = $1
		  AND l.period_start = %s
		ORDER BY l.total_duration DESC
		LIMIT 10;
	`, periodStartExpr)

	rows, err := p.Db.Query(context.Background(), query, periodType)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %v", err)
	}
	defer rows.Close()

	var leaderboard []types.LeaderboardEntry
	for rows.Next() {
		var entry types.LeaderboardEntry
		entry.PeriodStart = periodType
		err := rows.Scan(
			&entry.Rank,
			&entry.UserData.Id,
			&entry.UserData.Username,
			&entry.UserData.FullName,
			&entry.UserData.ProfilePic,
			&entry.UserData.NativeLanguage,
			&entry.TotalDuration,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %v", err)
		}
		leaderboard = append(leaderboard, entry)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating leaderboard rows: %v", rows.Err())
	}

	return leaderboard, nil
}

func (p *PostgreSQL) UpdateLeaderboard(userID int64, duration float64) error {
	now := time.Now()

	periods := []struct {
		Type  string
		Start time.Time
	}{
		{"daily", time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())},
		{"weekly", now.AddDate(0, 0, -int(now.Weekday()))}, // start of week
		{"monthly", time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())},
		{"alltime", time.Date(1970, 1, 1, 0, 0, 0, 0, now.Location())},
	}

	for _, pInfo := range periods {
		query := `
			INSERT INTO leaderboard (user_id, period_type, period_start, total_duration)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, period_type, period_start)
			DO UPDATE SET total_duration = leaderboard.total_duration + EXCLUDED.total_duration;
		`
		_, err := p.Db.Exec(context.Background(), query, userID, pInfo.Type, pInfo.Start, duration)
		if err != nil {
			return fmt.Errorf("failed to update %s leaderboard: %v", pInfo.Type, err)
		}
	}

	return nil
}

func FormatCallRecord(records []types.CallHistory, id int64) []types.CallRecord {
	var formattedRecords []types.CallRecord
	var formattedRecord types.CallRecord
	for _, record := range records {
		if record.PeerID1 == id {
			formattedRecord.PeerId = record.PeerID2
			formattedRecord.PeerPic = record.PeerPic2
			formattedRecord.PeerName = record.PeerName2
		} else {
			formattedRecord.PeerId = record.PeerID1
			formattedRecord.PeerPic = record.PeerPic1
			formattedRecord.PeerName = record.PeerName1
		}
		formattedRecord.CallEnd = record.CallEnd
		formattedRecord.CallId = record.CallId
		formattedRecord.CallStart = record.CallStart
		formattedRecord.DurationInMin = record.DurationInMin
		formattedRecord.Status = record.Status

		formattedRecords = append(formattedRecords, formattedRecord)
	}

	return formattedRecords
}
