package main

import (
	"database/sql"
	"time"

	"rustlang.pocha.moe/sdrmm/config"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

func isBanned(bsr string, db *sql.DB) bool {
	return database.FindBannedMap(bsr, db)
}

func isOlder(mapDate time.Time, minDate time.Time) bool {
	return mapDate.Before(minDate)
}

func isNewer(mapDate time.Time, mapAge int) bool {
	return time.Since(mapDate).Hours()/24 < float64(mapAge)
}

func isShorter(mapLength int, minLength int) bool {
	return mapLength < minLength
}

func isLonger(mapLength int, maxLength int) bool {
	return mapLength > maxLength
}

func userRequestedTooMuch(requestLimit, userNumRequests int) bool {
	return userNumRequests == requestLimit
}

func checkNJSandNPS(diffs []drm.MapDifficultyData, limits config.NoteLimits) (bool, bool) {
	// doing NPS and NJS checks in one loop, screw it
	var passedNJSCheck = false
	var passedNPSCheck = false

	for _, diff := range diffs {
		njs := diff.NoteJumpSpeed
		nps := diff.NotesPerSecond

		if limits.MaxNJS == 0 || limits.MinNJS <= njs && njs <= limits.MaxNJS {
			passedNJSCheck = true
		}

		if limits.MaxNPS == 0 || limits.MinNPS <= nps && nps <= limits.MaxNPS {
			passedNPSCheck = true
		}
	}

	return passedNJSCheck, passedNPSCheck
}

func isClosed(db *sql.DB) bool {
	return !database.GetQueueStatus(db)
}

func FilterMap(mapData drm.MapData, username string, numRequests int, modadd bool, config config.BSRConfig, db *sql.DB) (drm.MapData, error) {
	if modadd {
		return mapData, nil
	}
	duration := mapData.Duration

	// is the queue closed?
	if isClosed(db) {
		return mapData, &QueueIsClosedError{}
	}

	// is the map banned?
	if isBanned(mapData.BsrKey, db) {
		return mapData, &BannedMapError{id: mapData.BsrKey}
	}

	uploadTime := mapData.UploadTime.Local().UTC()

	// is the map older than a certain date?
	if isOlder(uploadTime, config.NewerThan) {
		return mapData, &OlderThanOldestDateError{date: mapData.UploadTime}
	}

	// has the map released bsr.map-age days ago?
	if isNewer(uploadTime, config.MapAge) {
		return mapData, &NewMapError{date: mapData.UploadTime}
	}
	// is the map too short?
	if config.MinLength != 0 && isShorter(duration, config.MinLength) {
		return mapData, &MapTooShortError{len: duration}
	}

	// is the map too long?
	if config.MaxLength != 0 && isLonger(duration, config.MaxLength) {
		return mapData, &MapTooLongError{len: duration}
	}

	// NJS and NPS checks
	passedNJSCheck, passedNPSCheck := checkNJSandNPS(mapData.Diffs, config.NoteLimits)

	if !passedNJSCheck {
		return mapData, &NotInNJSRangeError{minNJS: config.NoteLimits.MinNJS, maxNJS: config.NoteLimits.MaxNJS}
	}

	if !passedNPSCheck {
		return mapData, &NotInNPSRangeError{minNPS: config.NoteLimits.MinNPS, maxNPS: config.NoteLimits.MinNPS}
	}

	// did the user request enough maps this stream?
	if config.RequestLimit != 0 && userRequestedTooMuch(config.RequestLimit, numRequests) {
		return mapData, &UserMaxRequestsError{user: username}
	}

	// passed filters!
	return mapData, nil
}
