package queue

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"frauddetection/internal/domain"

	"github.com/beanstalkd/go-beanstalk"
)

type BeanstalkQueue struct {
	mu   sync.Mutex
	tube *beanstalk.Tube
}

func NewBeanstalkQueue(conn *beanstalk.Conn, tubeName string) *BeanstalkQueue {
	return &BeanstalkQueue{
		tube: &beanstalk.Tube{Conn: conn, Name: tubeName},
	}
}

func (q *BeanstalkQueue) Publish(
	ctx context.Context,
	job domain.Job,
) error {
	_, err := q.tube.Put(job.Payload, 1, 0, 30*time.Second)
	if err != nil {
		return fmt.Errorf(
			"infrastructure: failed to publish job: %w",
			err,
		)
	}
	return nil
}

func (q *BeanstalkQueue) Reserve(ctx context.Context) (*domain.Job, error) {
	q.mu.Lock()
	id, body, err := q.tube.Conn.Reserve(5 * time.Second)
	q.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("infrastructure: reserve failed: %w", err)
	}

	return &domain.Job{
		ID:      strconv.FormatUint(id, 10),
		Payload: body,
	}, nil
}

// Bury puts a job into a "buried" state where it will not be processed
// until it is manually "kicked" back into the ready queue.
func (q *BeanstalkQueue) Bury(ctx context.Context, id string) error {
	jobID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("infrastructure: invalid job id: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Priority 0 is the highest priority in Beanstalkd
	err = q.tube.Conn.Bury(jobID, 0)
	if err != nil {
		return fmt.Errorf(
			"infrastructure: failed to bury job %d: %w",
			jobID,
			err,
		)
	}
	return nil
}

func (q *BeanstalkQueue) Delete(ctx context.Context, id string) error {
	jobID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	return q.tube.Conn.Delete(jobID)
}
