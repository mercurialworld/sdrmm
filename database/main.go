package database

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"rustlang.pocha.moe/sdrmm/utils"
)

func InitializeDB() *sql.DB {
	db, err := sql.Open("sqlite", "database.db")
	utils.PanicOnError(err)

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS banned(id TEXT PRIMARY KEY NOT NULL, hash TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS reqlimits(username TEXT PRIMARY KEY NOT NULL, requests INT NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS queueStatus(timestamp INT PRIMARY KEY NOT NULL, status BOOLEAN);
	`)
	utils.PanicOnError(err)

	return db
}

func NewSession(db *sql.DB) {
	// add new session to queue status table
	// set to... true?
}

func CloseDB(db *sql.DB) {
	err := db.Close()
	utils.PanicOnError(err)
}

func BanMap(id string, hash string, db *sql.DB) {
	_, err := db.Query(`
INSERT INTO banned(id, hash) VALUES(?, ?)
	`, id, hash)
	utils.PanicOnError(err)
}

func UnbanMap(id string, db *sql.DB) {
	_, err := db.Query(`
DELETE FROM banned WHERE id=?
	`, id)
	utils.PanicOnError(err)
}

func FindBannedMap(id string, db *sql.DB) bool {
	var theMap Banned

	// if it's in the database, it's banned
	if err := db.QueryRow(`
SELECT id, hash FROM banned WHERE id=?	
	`, id).Scan(&theMap.id, &theMap.hash); err != nil {
		return false
	}

	return true
}

func GetUserRequests(user string, db *sql.DB) int {
	userRow := ReqLimits{user: user, requests: 0}

	if err := db.QueryRow(`
SELECT (username, requests) FROM reqLimits WHERE username=?	
	`, user).Scan(&userRow.user, &userRow.requests); err != nil {
		if err == sql.ErrNoRows {
			newUser, err := db.Query(`
INSERT INTO reqLimits(username, requests) VALUES ?, ?	
			`, user, 0)

			newUser.Scan(&userRow.user, &userRow.requests)
			utils.PanicOnError(err)
		}
	}

	return userRow.requests
}

func SetUserRequests(user string, numReqs int, db *sql.DB) {
	_, err := db.Query(`
INSERT INTO reqLimits(username, requests) VALUES ?, ?	
	`, user, numReqs)
	utils.PanicOnError(err)
}

func ClearRequestLimits(db *sql.DB) {
	_, err := db.Query(`
DELETE * FROM reqLimits 
	`)
	utils.PanicOnError(err)
}

func ToggleQueue(status bool, db *sql.DB) {
	// order queueStatus by timestamp
	// set newest timestamp's status to status
}
