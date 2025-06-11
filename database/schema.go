package database

type Banned struct {
	id   string
	hash string
}

type ReqLimits struct {
	user     string
	requests int // incremental
}
