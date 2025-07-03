package config

import "time"

type NoteLimits struct {
	MinNJS float64
	MaxNJS float64
	MinNPS float64
	MaxNPS float64
}

type BSRConfig struct {
	MinLength    int
	MaxLength    int
	RequestLimit int
	NewerThan    time.Time
	MapAge       int
	NoteLimits   NoteLimits
}
