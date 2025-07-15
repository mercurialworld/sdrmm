package main

import (
	"time"

	"rustlang.pocha.moe/sdrmm/config"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
)

type Filterer struct {
	config   config.BSRConfig
	database database.DRMDatabase
}

func (f Filterer) isBanned(bsr string) bool {
	return f.database.FindBannedMap(bsr)
}

func (f Filterer) isOlder(mapDate time.Time) bool {
	return mapDate.Before(f.config.NewerThan)
}

func (f Filterer) isNewer(mapDate time.Time) bool {
	return time.Since(mapDate).Hours()/24 < float64(f.config.MapAge)
}

func (f Filterer) isShorter(mapLength int) bool {
	return mapLength < f.config.MinLength
}

func (f Filterer) isLonger(mapLength int) bool {
	return mapLength > f.config.MaxLength
}

func (f Filterer) userRequestedTooMuch(userNumRequests int) bool {
	return userNumRequests == f.config.RequestLimit
}

func (f Filterer) userTooMuchInQueue(userReqsInQueue int) bool {
	return userReqsInQueue == f.config.QueueRequestLimit
}

func (f Filterer) checkNJSandNPS(diffs []drm.MapDifficultyData) (bool, bool) {
	// doing NPS and NJS checks in one loop, screw it
	var passedNJSCheck = false
	var passedNPSCheck = false

	for _, diff := range diffs {
		njs := diff.NoteJumpSpeed
		nps := diff.NotesPerSecond

		if f.config.NoteLimits.MaxNJS == 0 || f.config.NoteLimits.MinNJS <= njs && njs <= f.config.NoteLimits.MaxNJS {
			passedNJSCheck = true
		}

		if f.config.NoteLimits.MaxNPS == 0 || f.config.NoteLimits.MinNPS <= nps && nps <= f.config.NoteLimits.MaxNPS {
			passedNPSCheck = true
		}
	}

	return passedNJSCheck, passedNPSCheck
}

func (f Filterer) isClosed() bool {
	return !f.database.GetQueueStatus()
}

func (f Filterer) FilterMap(mapData drm.MapData, username string, modadd bool) (drm.MapData, error) {
	if modadd {
		return mapData, nil
	}
	duration := mapData.Duration
	userRequests := f.database.GetUserRequests(username)
	userRequestsInQueue := f.database.GetUserRequestsInQueue(username)

	// is the queue closed?
	if f.isClosed() {
		return mapData, &QueueIsClosedError{}
	}

	// is the map banned?
	if f.isBanned(mapData.BsrKey) {
		return mapData, &BannedMapError{id: mapData.BsrKey}
	}

	uploadTime := mapData.UploadTime.Local().UTC()

	// is the map older than a certain date?
	if f.isOlder(uploadTime) {
		return mapData, &OlderThanOldestDateError{date: mapData.UploadTime}
	}

	// has the map released bsr.map-age days ago?
	if f.isNewer(uploadTime) {
		return mapData, &NewMapError{date: mapData.UploadTime}
	}
	// is the map too short?
	if f.config.MinLength != 0 && f.isShorter(duration) {
		return mapData, &MapTooShortError{len: duration}
	}

	// is the map too long?
	if f.config.MaxLength != 0 && f.isLonger(duration) {
		return mapData, &MapTooLongError{len: duration}
	}

	// NJS and NPS checks
	passedNJSCheck, passedNPSCheck := f.checkNJSandNPS(mapData.Diffs)

	if !passedNJSCheck {
		return mapData, &NotInNJSRangeError{minNJS: f.config.NoteLimits.MinNJS, maxNJS: f.config.NoteLimits.MaxNJS}
	}

	if !passedNPSCheck {
		return mapData, &NotInNPSRangeError{minNPS: f.config.NoteLimits.MinNPS, maxNPS: f.config.NoteLimits.MinNPS}
	}

	// did the user request enough maps this stream?
	if f.config.RequestLimit != 0 && f.userRequestedTooMuch(userRequests) {
		return mapData, &UserMaxRequestsError{user: username, maxRequests: f.config.RequestLimit}
	}

	// does the user have too many requests currently in queue?
	if f.config.QueueRequestLimit != 0 && f.userTooMuchInQueue(userRequestsInQueue) {
		return mapData, &UserQueueMaxRequestsError{user: username, maxInQueue: f.config.QueueRequestLimit}
	}

	// passed filters!
	return mapData, nil
}
