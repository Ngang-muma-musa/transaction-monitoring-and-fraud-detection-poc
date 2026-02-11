package domain

// Job represents the standard payload format pushed to the queue.
type Job struct {
	ID      string
	Payload []byte
}
