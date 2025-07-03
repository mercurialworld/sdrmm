package config

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
	NoteLimits   NoteLimits
}
