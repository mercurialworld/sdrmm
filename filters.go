package main

import (
	"database/sql"
	"time"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
	"rustlang.pocha.moe/sdrmm/utils"
)

func FilterMap(mapData drm.MapData, db *sql.DB) (drm.MapData, error) {
	// is the map banned?
	if database.FindBannedMap(mapData.BsrKey, db) {
		return mapData, &BannedMapError{id: mapData.BsrKey}
	}

	// is the map older than a certain date?
	olderThan, err := time.Parse("2006-01-02", viper.GetString("bsr.newer-than"))
	utils.HandleError(err)

	if mapData.UploadTime.Local().UTC().Before(olderThan) {
		return mapData, &OlderThanOldestDateError{mapData.UploadTime}
	}

	// has the map released bsr.map-age days ago?
	if (time.Since(mapData.UploadTime.Time).Hours() / 24) < float64(viper.GetInt("bsr.map-age")) {
		return mapData, &NewMapError{date: mapData.UploadTime}
	}

	// doing NPS and NJS checks in one loop, screw it
	var passedNPSCheck = false
	var passedNJSCheck = false

	// NPS
	minNPS := viper.GetFloat64("nps.min")
	maxNPS := viper.GetFloat64("nps.max")

	// NJS
	minNJS := viper.GetFloat64("njs.min")
	maxNJS := viper.GetFloat64("njs.max")

	for _, diff := range mapData.Diffs {
		nps := diff.NotesPerSecond
		njs := diff.NoteJumpSpeed

		if maxNPS == 0 || minNPS <= nps && nps <= maxNPS {
			passedNPSCheck = true
		}

		if maxNJS == 0 || minNJS <= njs && njs <= maxNJS {
			passedNJSCheck = true
		}
	}

	if !passedNJSCheck {
		return mapData, &NotInNJSRangeError{minNJS: minNJS, maxNJS: maxNJS}
	}

	if !passedNPSCheck {
		return mapData, &NotInNPSRangeError{minNPS: minNPS, maxNPS: maxNPS}
	}

	// passed filters!
	return mapData, nil
}
