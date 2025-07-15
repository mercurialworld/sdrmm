package database

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"rustlang.pocha.moe/sdrmm/utils"
)

type DRMDatabase struct {
	DB *sql.DB
}

func InitializeDB() *sql.DB {
	db, err := sql.Open("sqlite", "database.db")
	utils.PanicOnError(err)

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS banned(id TEXT PRIMARY KEY NOT NULL, hash TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS reqlimits(username TEXT PRIMARY KEY NOT NULL, successfulRequests INT NOT NULL DEFAULT 0, requestsInQueue INT NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS queueStatus(timestamp INT PRIMARY KEY NOT NULL, status BOOLEAN);
	`)
	utils.PanicOnError(err)

	return db
}

func (db DRMDatabase) NewSession(timestamp time.Time) {
	// add new session to queue status table, with status set to false
	_, err := db.DB.Query(`
INSERT INTO queueStatus(timestamp, status) VALUES (?, ?)
	`, timestamp.Unix(), false)
	utils.PanicOnError(err)
}

func (db DRMDatabase) CloseDB() {
	err := db.DB.Close()
	utils.PanicOnError(err)
}

func (db DRMDatabase) BanMap(id string, hash string) {
	_, err := db.DB.Query(`
INSERT INTO banned(id, hash) VALUES(?, ?)
	`, id, hash)
	utils.PanicOnError(err)
}

func (db DRMDatabase) UnbanMap(id string) {
	_, err := db.DB.Query(`
DELETE FROM banned WHERE id=?
	`, id)
	utils.PanicOnError(err)
}

func (db DRMDatabase) FindBannedMap(id string) bool {
	var theMap Banned

	// if it's in the database, it's banned
	if err := db.DB.QueryRow(`
SELECT id, hash FROM banned WHERE id=?	
	`, id).Scan(&theMap.id, &theMap.hash); err != nil {
		return false
	}

	return true
}

func (db DRMDatabase) GetUserRequests(user string) int {
	userRow := UserRequests{user: user, successfulRequests: 0}

	if err := db.DB.QueryRow(`
SELECT username, successfulRequests FROM reqLimits WHERE username=?	
	`, user).Scan(&userRow.user, &userRow.successfulRequests); err != nil {
		if err == sql.ErrNoRows {
			newUser, err := db.DB.Query(`
INSERT INTO reqLimits(username, successfulRequests) VALUES (?, ?)	
			`, user, 0)

			newUser.Scan(&userRow.user, &userRow.successfulRequests)
			utils.PanicOnError(err)
		}
	}

	return userRow.successfulRequests
}

func (db DRMDatabase) SetUserRequests(user string, numReqs int) {
	_, err := db.DB.Query(`
INSERT INTO reqLimits(username, successfulRequests) VALUES (?, ?)
	`, user, numReqs)
	utils.PanicOnError(err)
}

func (db DRMDatabase) ClearRequestLimits() {
	_, err := db.DB.Query(`
DELETE FROM reqLimits 
	`)
	utils.PanicOnError(err)
}

func (db DRMDatabase) SetQueueStatus(status bool) {
	// order queueStatus by timestamp
	queueStatus := QueueStatus{timestamp: 0, status: false}

	// get first row
	err := db.DB.QueryRow(`
SELECT timestamp, status FROM queueStatus ORDER BY timestamp DESC LIMIT 1
	`).Scan(&queueStatus.timestamp, &queueStatus.status)
	// something is wrong if there are no rows
	utils.PanicOnError(err)

	// set newest timestamp's status to status
	queueStatus.status = status

	// insert it back
	_, err = db.DB.Query(`
UPDATE queueStatus SET status=? WHERE timestamp=?	
	`, queueStatus.status, queueStatus.timestamp)
	utils.PanicOnError(err)
}

func (db DRMDatabase) GetQueueStatus() bool {
	queueStatus := QueueStatus{timestamp: 0, status: false}

	// get first row
	err := db.DB.QueryRow(`
SELECT timestamp, status FROM queueStatus ORDER BY timestamp DESC
	`).Scan(&queueStatus.timestamp, &queueStatus.status)
	// something is wrong if there are no rows
	utils.PanicOnError(err)

	// return the status
	return queueStatus.status
}

func (db DRMDatabase) GetUserRequestsInQueue(user string) int {
	userRow := UserRequests{user: user, requestsInQueue: 0}

	if err := db.DB.QueryRow(`
SELECT username, requestsInQueue FROM reqLimits WHERE username=?	
	`, user).Scan(&userRow.user, &userRow.requestsInQueue); err != nil {
		if err == sql.ErrNoRows {
			newUser, err := db.DB.Query(`
INSERT INTO reqLimits(username, requestsInQueue) VALUES (?, ?)	
			`, user, 0)

			newUser.Scan(&userRow.user, &userRow.requestsInQueue)
			utils.PanicOnError(err)
		}
	}

	return userRow.requestsInQueue
}

func (db DRMDatabase) SetUserRequestsInQueue(user string, numReqs int) {
	_, err := db.DB.Query(`
INSERT INTO reqLimits(username, requestsInQueue) VALUES (?, ?)
	`, user, numReqs)
	utils.PanicOnError(err)
}
