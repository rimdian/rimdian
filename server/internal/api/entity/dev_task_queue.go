package entity

import (
	"math/rand"
	"sync"

	"github.com/rimdian/rimdian/internal/common/dto"
)

// the dev dataLog queue is an inmemory object used to process async data imports
type DevDataImportQueue struct {
	List []*dto.DataLogInQueue
	mu   sync.Mutex
}

// picks a random dataLog from the queue and returns it
// many data imports contain user IDs that are also present in the next data import
// this produces user locks waiting for the next data import to be processed
// to avoid this, we pick a random dataLog from the queue and return it
func (queue *DevDataImportQueue) GetOne() (dataLogInQueue *dto.DataLogInQueue) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	if len(queue.List) == 0 {
		return nil
	}

	// get a data import from the queue at random index
	randomIndex := rand.Intn(len(queue.List))
	dataLogInQueue = queue.List[randomIndex]

	// remove the data import from the queue at given index
	queue.List = append(queue.List[:randomIndex], queue.List[randomIndex+1:]...)

	return dataLogInQueue
}

func (queue *DevDataImportQueue) Add(dataLog *dto.DataLogInQueue) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	queue.List = append(queue.List, dataLog)
}

func NewDevDataImportQueue() *DevDataImportQueue {
	return &DevDataImportQueue{
		List: []*dto.DataLogInQueue{},
	}
}
