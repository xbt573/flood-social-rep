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

var (
	// attempts is a last users reaction attempts
	attempts = map[int64]time.Time{}
	// attemptsmux is a mutex for data race security
	attemptsmux = sync.RWMutex{}
)

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

		CREATE TABLE IF NOT EXISTS username(
		    user_id INTEGER NOT NULL,
		    username TEXT NOT NULL
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
		case "❤":
			fallthrough
		case "❤‍🔥":
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
		case "❤":
			fallthrough
		case "❤‍🔥":
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

	attemptsmux.RLock()

	attemptExists := false
	for id, attempt := range attempts {
		if userId != id {
			continue
		}

		attemptExists = true
		attempts[id] = time.Now()

		if time.Since(attempt) < time.Second*15 {
			attemptsmux.RUnlock()
			return nil
		}
	}

	attemptsmux.RUnlock()

	if !attemptExists {
		attemptsmux.Lock()
		attempts[userId] = time.Now()
		attemptsmux.Unlock()
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

// UpdateUsername is a function which adds username into database
// (used when getChatMember is fucked)
func UpdateUsername(userId int64, username string) error {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM username WHERE user_id=?", userId)

	// dummy
	var a, b any = nil, nil
	err = row.Scan(&a, &b)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		_, err = db.Exec("INSERT INTO username VALUES(?, ?)", userId, username)
		if err != nil {
			return err
		}

		return nil
	}

	_, err = db.Exec(
		"UPDATE username SET username=? WHERE user_id=?",
		username,
		userId,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetUsername is a function which gets username from database
// (used when getChatMember is fucked)
func GetUsername(userId int64) (string, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow("SELECT username FROM username WHERE user_id=?", userId)

	var username string
	err = row.Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

// GetReactions is a function which returns reactions set on messageId in chatId
func GetReactions(chatId, messageId int64) ([]models.Reaction, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return []models.Reaction{}, err
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT from_user_id, reaction FROM reactions WHERE chat_id=? AND message_id=?",
		chatId,
		messageId,
	)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return []models.Reaction{}, err
		}

		return []models.Reaction{}, nil
	}

	var reactions []models.Reaction

	for rows.Next() {
		var fromUserId int64
		var reaction string

		err := rows.Scan(&fromUserId, &reaction)
		if err != nil {
			return []models.Reaction{}, err
		}

		reactions = append(reactions, models.Reaction{
			UserId:   fromUserId,
			Reaction: reaction,
		})
	}

	err = rows.Err()
	if err != nil {
		return []models.Reaction{}, err
	}

	return reactions, nil
}
