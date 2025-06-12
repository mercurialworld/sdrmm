package main

import (
	"fmt"

	"rustlang.pocha.moe/sdrmm/utils"
)

// Maps outside NJS range
type NotInNJSRangeError struct {
	minNJS float64
	maxNJS float64
}

func (e *NotInNJSRangeError) Error() string {
	return fmt.Sprintf("map does not have a difficulty between %f and %f NJS", e.minNJS, e.maxNJS)
}

// Maps outside NPS range
type NotInNPSRangeError struct {
	minNPS float64
	maxNPS float64
}

func (e *NotInNPSRangeError) Error() string {
	return fmt.Sprintf("map does not have a difficulty between %f and %f NPS", e.minNPS, e.maxNPS)

}

// Maps uploaded before bsr.newer-than
type OlderThanOldestDateError struct {
	date utils.UnixTime
}

func (e *OlderThanOldestDateError) Error() string {
	// what the fuck is this time formatting
	// fuck off google https://go.dev/src/time/format.go
	return fmt.Sprintf("map is too old, was uploaded on %s", e.date.Format("2006 January 2"))
}

// Maps younger than bsr.map-age days
type NewMapError struct {
	date utils.UnixTime
}

func (e *NewMapError) Error() string {
	return fmt.Sprintf("map is too new, was uploaded on %s", e.date.Format("2006 January 2"))
}

// Banned map
type BannedMapError struct {
	id string
}

func (e *BannedMapError) Error() string {
	return fmt.Sprintf("%s is banned from being requested", e.id)
}

// Map too short
type MapTooShortError struct {
	len int
}

func (e *MapTooShortError) Error() string {
	return fmt.Sprintf("map is too short (length %d:%d)", (e.len/60)%60, e.len%60)
}

// Map too long
type MapTooLongError struct {
	len int
}

func (e *MapTooLongError) Error() string {
	return fmt.Sprintf("map is too long (length %d:%d)", (e.len/60)%60, e.len%60)
}
