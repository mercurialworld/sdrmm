package database

type Banned struct {
	id   string
	hash string
}

type UserRequests struct {
	user               string
	successfulRequests int // incremental
	requestsInQueue    int
}

type QueueStatus struct {
	status    bool
	timestamp int
}
