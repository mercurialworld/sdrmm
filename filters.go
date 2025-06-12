package main

import (
	"database/sql"
	"time"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

type NoteLimits struct {
	minNJS float64
	maxNJS float64
	minNPS float64
	maxNPS float64
}

func isBanned(bsr string, db *sql.DB) bool {
	return database.FindBannedMap(bsr, db)
}

func isOlder(mapDate time.Time) bool {
	olderThan, err := time.Parse("2006-01-02", viper.GetString("bsr.newer-than"))
	if err != nil {
		olderThan, _ = time.Parse("2006-01-02", "2000-01-01")
	}

	return mapDate.Before(olderThan)
}

func isNewer(mapDate time.Time) bool {
	return time.Since(mapDate).Hours()/24 < float64(viper.GetInt("bsr.map-age"))
}

func isShorter(mapLength int, minLength int) bool {
	return mapLength < minLength
}

func isLonger(mapLength int, maxLength int) bool {
	return mapLength > maxLength
}

func userRequestedTooMuch(requestLimit int, username string, db *sql.DB) bool {
	userNumRequests := database.GetUserRequests(username, db)

	return userNumRequests == requestLimit
}

func checkNJSandNPS(diffs []drm.MapDifficultyData, limits NoteLimits) (bool, bool) {
	// doing NPS and NJS checks in one loop, screw it
	var passedNJSCheck = false
	var passedNPSCheck = false

	for _, diff := range diffs {
		njs := diff.NoteJumpSpeed
		nps := diff.NotesPerSecond

		if limits.maxNJS == 0 || limits.minNJS <= njs && njs <= limits.maxNJS {
			passedNJSCheck = true
		}

		if limits.maxNPS == 0 || limits.minNPS <= nps && nps <= limits.maxNPS {
			passedNPSCheck = true
		}
	}

	return passedNJSCheck, passedNPSCheck
}

func FilterMap(mapData drm.MapData, username string, db *sql.DB) (drm.MapData, error) {
	// is the map banned?
	if isBanned(mapData.BsrKey, db) {
		return mapData, &BannedMapError{id: mapData.BsrKey}
	}

	uploadTime := mapData.UploadTime.Local().UTC()

	// is the map older than a certain date?
	if isOlder(uploadTime) {
		return mapData, &OlderThanOldestDateError{date: mapData.UploadTime}
	}

	// has the map released bsr.map-age days ago?
	if isNewer(uploadTime) {
		return mapData, &NewMapError{date: mapData.UploadTime}
	}

	minLength := viper.GetInt("min-length")
	maxLength := viper.GetInt("max-length")
	duration := mapData.Duration

	// is the map too short?
	if minLength != 0 && isShorter(duration, minLength) {
		return mapData, &MapTooShortError{len: duration}
	}

	// is the map too long?
	if maxLength != 0 && isLonger(duration, maxLength) {
		return mapData, &MapTooLongError{len: duration}
	}

	// NJS and NPS checks
	noteLimits := NoteLimits{
		minNJS: viper.GetFloat64("njs.min"),
		maxNJS: viper.GetFloat64("njs.max"),
		minNPS: viper.GetFloat64("nps.min"),
		maxNPS: viper.GetFloat64("nps.max"),
	}

	passedNJSCheck, passedNPSCheck := checkNJSandNPS(mapData.Diffs, noteLimits)

	if !passedNJSCheck {
		return mapData, &NotInNJSRangeError{minNJS: noteLimits.minNJS, maxNJS: noteLimits.maxNJS}
	}

	if !passedNPSCheck {
		return mapData, &NotInNPSRangeError{minNPS: noteLimits.minNPS, maxNPS: noteLimits.maxNPS}
	}

	// did the user request enough maps this stream?
	requestLimit := viper.GetInt("request-limit")
	if requestLimit != 0 && userRequestedTooMuch(requestLimit, username, db) {
		return mapData, &UserMaxRequestsError{user: username}
	}

	// passed filters!
	return mapData, nil
}
