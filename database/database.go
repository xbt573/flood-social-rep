// Package database is used to make database rating operations
package database

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3" // side-effect import
	"github.com/xbt573/flood-social-rep/models"
	"sync"
)

// mux is sync.Mutex which is locked where database operation is pending
var mux = sync.Mutex{}

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
		CREATE TABLE IF NOT EXISTS rating (
    		chat_id INTEGER NOT NULL,
    		user_id INTEGER NOT NULL,
    		rating INTEGER NOT NULL,

    		PRIMARY KEY ( chat_id, user_id )
		);

		CREATE TABLE IF NOT EXISTS set_reactions (
    		user_id INTEGER NOT NULL,
    		chat_id INTEGER NOT NULL,
    		message_id INTEGER NOT NULL
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	return nil
}

// TopRating is a function which returns top 10 users by rating, descending
func TopRating(chatId int64) ([]models.User, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return []models.User{}, err
	}
	defer db.Close()

	// Get top 10 users by rating, descending
	rows, err := db.Query(
		"SELECT * FROM rating WHERE chat_id=? ORDER BY rating DESC LIMIT 10",
		chatId,
	)
	if err != nil {
		return []models.User{}, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var userId int64
		var rating int

		err := rows.Scan(&chatId, &userId, &rating)
		if err != nil {
			return []models.User{}, err
		}

		users = append(users, models.User{
			UserId: userId,
			Rating: rating,
		})
	}

	err = rows.Err()
	if err != nil {
		return []models.User{}, err
	}

	return users, nil
}

// TopReverseRating is a function which returns top 10 users by rating, ascending
func TopReverseRating(chatId int64) ([]models.User, error) {
	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return []models.User{}, err
	}
	defer db.Close()

	// Get top 10 users by rating, ascending
	rows, err := db.Query(
		"SELECT * FROM rating WHERE chat_id=? ORDER BY rating ASC LIMIT 10",
		chatId,
	)
	if err != nil {
		return []models.User{}, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var userId int64
		var rating int

		err := rows.Scan(&chatId, &userId, &rating)
		if err != nil {
			return []models.User{}, err
		}

		users = append(users, models.User{
			UserId: userId,
			Rating: rating,
		})
	}

	err = rows.Err()
	if err != nil {
		return []models.User{}, err
	}

	return users, nil
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

	row := db.QueryRow(
		"SELECT * FROM rating WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	var rating int

	// Scan row and create zero-rating user if error is equal
	// to sql.ErrNoRows
	err = row.Scan(&chatId, &userId, &rating)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return models.User{}, err
		}

		_, err := db.Exec(
			"INSERT INTO rating VALUES(?, ?, ?);",
			chatId,
			userId,
			0,
		)

		if err != nil {
			return models.User{}, err
		}
	}

	return models.User{
		UserId: userId,
		Rating: rating,
	}, nil
}

// IncrementUserRating is a function which increments user rating
func IncrementUserRating(messageId int, chatId, fromUserId, userId int64) error {
	// No rating for you, buddy
	if userId == fromUserId {
		return nil
	}

	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT * FROM set_reactions WHERE chat_id=? AND user_id=? AND message_id=?",
		chatId,
		fromUserId,
		messageId,
	)

	// If error is equal to sql.ErrNoRows then allow increment reaction
	// a, b, c is a dummy values!
	var a, b, c any = nil, nil, nil
	err = row.Scan(&a, &b, &c)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		return nil
	}

	row = db.QueryRow(
		"SELECT * FROM rating WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	var rating int

	// Scan row and create zero-rating user if error is equal
	// to sql.ErrNoRows
	err = row.Scan(&chatId, &userId, &rating)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		_, err := db.Exec(
			"INSERT INTO rating VALUES(?, ?, ?);",
			chatId,
			userId,
			1,
		)

		if err != nil {
			return err
		}

		_, err = db.Exec(
			"INSERT INTO set_reactions VALUES(?, ?, ?)",
			fromUserId,
			chatId,
			messageId,
		)

		if err != nil {
			return err
		}

		return nil
	}

	_, err = db.Exec(
		"UPDATE rating SET rating=? WHERE user_id=? and chat_id=?;",
		rating+1,
		userId,
		chatId,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO set_reactions VALUES(?, ?, ?)",
		fromUserId,
		chatId,
		messageId,
	)
	if err != nil {
		return err
	}

	return nil
}

// DecrementUserRating is a function which decrements user rating
func DecrementUserRating(messageId int, chatId, fromUserId, userId int64) error {
	// No karma for you, buddy
	if userId == fromUserId {
		return nil
	}

	mux.Lock()
	defer mux.Unlock()

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT * FROM set_reactions WHERE chat_id=? AND user_id=? AND message_id=?",
		chatId,
		fromUserId,
		messageId,
	)

	// If errors is sql.ErrNoRows then allow decrementing rating
	// a, b, c is a dummy values!
	var a, b, c any = nil, nil, nil
	err = row.Scan(&a, &b, &c)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		return nil
	}

	row = db.QueryRow(
		"SELECT * FROM rating WHERE chat_id=? AND user_id=?",
		chatId,
		userId,
	)

	var rating int

	// Scan row and create zero-rating user if error is equal
	// to sql.ErrNoRows
	err = row.Scan(&chatId, &userId, &rating)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		_, err := db.Exec(
			"INSERT INTO rating VALUES(?, ?, ?);",
			chatId,
			userId,
			1,
		)

		if err != nil {
			return err
		}

		_, err = db.Exec(
			"INSERT INTO set_reactions VALUES(?, ?, ?)",
			fromUserId,
			chatId,
			messageId,
		)

		if err != nil {
			return err
		}

		return nil
	}

	_, err = db.Exec(
		"UPDATE rating SET rating=? WHERE user_id=? and chat_id=?;",
		rating-1,
		userId,
		chatId,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO set_reactions VALUES(?, ?, ?)",
		fromUserId,
		chatId,
		messageId,
	)
	if err != nil {
		return err
	}

	return nil
}
