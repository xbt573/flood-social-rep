// Package database is used to make database rating operations
package database

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3" // side-effect import
	"github.com/xbt573/flood-social-rep/models"
	"golang.org/x/exp/maps"
	"strings"
	"sync"
	"time"
)

// mux is sync.Mutex which is locked where database operation is pending
var mux = sync.Mutex{}

// attempts is a last users reaction attempts
var attempts []models.Attempt

// Blacklist errors
var (
	ErrAlreadyBlacklisted = errors.New("user already in blacklist")
	ErrNotInBlacklist     = errors.New("user is not in blacklist")
)

// Init is a function which initializes database for first time use
// (if was not initialized before). Returns non-nil error if
// something goes wrong!
func Init() error {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS reactions(
		    chat_id INTEGER NOT NULL,
		    from_user_id INTEGER NOT NULL,
		    user_id INTEGER NOT NULL,
		    message_id INTEGER NOT NULL,
		    reaction TEXT NOT NULL,
		    
		    PRIMARY KEY ( chat_id, from_user_id, user_id, message_id, reaction )
		);

		CREATE TABLE IF NOT EXISTS blacklist(
			chat_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL  
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

// TopRating is a function which returns users rating
func TopRating(chatId int64) ([]models.User, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return []models.User{}, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT user_id, reaction FROM reactions WHERE chat_id=?`,
		chatId,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return []models.User{}, err
		}

		return []models.User{}, nil
	}

	usermap := map[int64]models.User{}

	for rows.Next() {
		var userId int64
		var reaction string

		err := rows.Scan(&userId, &reaction)
		if err != nil {
			return []models.User{}, err
		}

		if _, exists := usermap[userId]; !exists {
			usermap[userId] = models.User{
				UserId: userId,
			}
		}

		switch reaction {
		case "👍": // Positive reactions
			fallthrough
		case "🔥":
			fallthrough
		case "❤️":
			fallthrough
		case "👏":
			fallthrough
		case "💯":
			tmp := usermap[userId]
			tmp.Likes++

			usermap[userId] = tmp
		case "🤡": // Negative reactions
			fallthrough
		case "💩":
			fallthrough
		case "🤮":
			fallthrough
		case "👎":
			tmp := usermap[userId]
			tmp.Dislikes++

			usermap[userId] = tmp

		case "🐳": // whale bruh
			tmp := usermap[userId]
			tmp.Whales++

			usermap[userId] = tmp
		}
	}

	err = rows.Err()
	if err != nil {
		return []models.User{}, err
	}

	return maps.Values(usermap), nil
}

// GetUserRating is a function which returns rating for specific user.
func GetUserRating(chatId, userId int64) (models.User, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return models.User{}, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT reaction FROM reactions WHERE chat_id=? AND user_id=?`,
		chatId,
		userId,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return models.User{}, err
		}

		return models.User{
			UserId: userId,
		}, nil
	}

	var user models.User

	for rows.Next() {
		var reaction string

		err := rows.Scan(&reaction)
		if err != nil {
			return models.User{}, err
		}

		switch reaction {
		case "👍": // Positive reactions
			fallthrough
		case "🔥":
			fallthrough
		case "❤️":
			fallthrough
		case "👏":
			fallthrough
		case "💯":
			user.Likes++

		case "🤡": // Negative reactions
			fallthrough
		case "💩":
			fallthrough
		case "🤮":
			fallthrough
		case "👎":
			user.Dislikes++

		case "🐳": // whale bruh
			user.Whales++
		}
	}

	err = rows.Err()
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// AddReaction is a function which adds reaction to database
func AddReaction(chatId, fromUserId, userId, messageId int64, reaction string) error {
	// No karma for you, buddy
	if userId == fromUserId {
		return nil
	}

	var attemptExists bool
	for idx, attempt := range attempts {
		if attempt.UserId != fromUserId {
			continue
		}

		attemptExists = true
		attempts[idx].Time = time.Now()

		if time.Since(attempt.Time) <= 15*time.Second {
			return nil
		}

		break
	}

	if !attemptExists {
		attempts = append(attempts, models.Attempt{
			UserId: fromUserId,
			Time:   time.Now(),
		})
	}

	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT * FROM blacklist WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	// dummy values
	var a, b any = nil, nil
	err = row.Scan(&a, &b)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		// blacklist clause
		return nil
	}

	_, err = db.Exec(
		`INSERT INTO reactions VALUES(?, ?, ?, ?, ?)`,
		chatId,
		fromUserId,
		userId,
		messageId,
		reaction,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// ignore constraint error 🐳
			return nil
		}

		return err
	}

	return nil
}

// AddBlacklist is a function which adds user into blacklist
func AddBlacklist(chatId, userId int64) error {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT * FROM blacklist WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	// dummy values
	var a, b any = nil, nil
	err = row.Scan(&a, &b)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		return ErrAlreadyBlacklisted
	}

	_, err = db.Exec(`INSERT INTO blacklist VALUES(?, ?)`, chatId, userId)
	if err != nil {
		return err
	}

	return nil
}

// RemoveBlacklist is a function which removes user from blacklist
func RemoveBlacklist(chatId, userId int64) error {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT * FROM blacklist WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	// dummy values
	var a, b any = nil, nil
	err = row.Scan(&a, &b)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		return ErrNotInBlacklist
	}

	_, err = db.Exec(
		"DELETE FROM blacklist WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}
