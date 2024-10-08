package service

import (
	"downloader_torrent/model"
	errorHandler "downloader_torrent/pkg/error"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDownloadQueue interface {
	Enqueue(queueItem *QueueItem) (int, error)
	Dequeue() (*QueueItem, bool)
	GetIndex(torrentLink string) (int, bool)
	GetIndexAndReturn(torrentLink string) (*QueueItem, int, bool)
	GetItemOfUser(userId int64) ([]QueueItem, []int)
	periodicSaveQueue()
	checkSave()
	saveQueue()
	loadQueue()
	Start(consumerDequeueCheckFunc ConsumerDequeueCheckFunc, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration)
	worker(wid int, consumerDequeueCheckFunc ConsumerDequeueCheckFunc, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration)
	Close()
	GetStats() *model.DownloadQueueStats
}

type DownloadQueue struct {
	queue               []*QueueItem
	mutex               sync.Mutex
	queueFile           string
	capacity            int
	workers             int
	saveQueueInterval   time.Duration
	batchSize           int
	operationCount      int
	doneChan            chan bool
	done                bool
	wg                  *sync.WaitGroup
	enqueueCounter      int
	dequeueWorkersSleep bool
}

func NewDownloadQueue(queueFile string, workers int, capacity int, saveQueueInterval time.Duration, batchSize int) *DownloadQueue {
	dq := &DownloadQueue{
		queue:             make([]*QueueItem, 0, capacity),
		queueFile:         queueFile,
		capacity:          capacity,
		workers:           workers,
		saveQueueInterval: saveQueueInterval,
		batchSize:         batchSize,
		operationCount:    0,
		doneChan:          make(chan bool),
		done:              false,
		wg:                &sync.WaitGroup{},
	}

	dq.loadQueue()
	go dq.periodicSaveQueue()
	dq.enqueueCounter = 0

	return dq
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) GetStats() *model.DownloadQueueStats {
	return &model.DownloadQueueStats{
		Size:                len(dq.queue),
		EnqueueCounter:      dq.enqueueCounter,
		Capacity:            dq.capacity,
		Workers:             dq.workers,
		DequeueWorkersSleep: dq.dequeueWorkersSleep,
	}
}

//---------------------------------------
//---------------------------------------

type ConsumerFunc func(wid int, queueItem *QueueItem)
type ConsumerDequeueCheckFunc func(wid int) bool

type QueueItem struct {
	TitleId       primitive.ObjectID `json:"titleId"`
	TitleType     string             `json:"titleType"`
	TorrentLink   string             `json:"torrentLink"`
	EnqueueTime   time.Time          `json:"enqueueTime"`
	EnqueueSource EnqueueSource      `json:"enqueueSource"`
	UserId        int64              `json:"userId"`
	BotId         string             `json:"botId"`
	ChatId        string             `json:"chatId"`
	BotUsername   string             `json:"botUsername"`
}

type EnqueueSource string

const (
	Admin          EnqueueSource = "admin"
	User           EnqueueSource = "user"
	UserBot        EnqueueSource = "user-bot"
	AutoDownloader EnqueueSource = "auto-downloader"
)

var ErrOverflow = errors.New("overflow")

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) Enqueue(queueItem *QueueItem) (int, error) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	if len(dq.queue) >= dq.capacity {
		return -1, ErrOverflow
	}

	dq.queue = append(dq.queue, queueItem)

	dq.enqueueCounter++
	dq.checkSave()

	return len(dq.queue) - 1, nil
}

func (dq *DownloadQueue) Dequeue() (*QueueItem, bool) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	if len(dq.queue) == 0 {
		return &QueueItem{}, false
	}

	queueItem := dq.queue[0]
	dq.queue = dq.queue[1:]

	dq.checkSave()

	return queueItem, true
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) Start(consumerDequeueCheckFunc ConsumerDequeueCheckFunc, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration) {
	for i := 0; i < dq.workers; i++ {
		dq.wg.Add(1)
		go dq.worker(i, consumerDequeueCheckFunc, consumerFunc, emptyQueueSleep)
	}
}

func (dq *DownloadQueue) worker(wid int, consumerDequeueCheckFunc ConsumerDequeueCheckFunc, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration) {
	defer dq.wg.Done()

	for {
		if dq.done == true {
			//fmt.Println("consumer closed,done ", wid)
			return
		}

		if !consumerDequeueCheckFunc(wid) {
			dq.dequeueWorkersSleep = true
			time.Sleep(10 * time.Second)
			continue
		}
		dq.dequeueWorkersSleep = false

		item, exist := dq.Dequeue()

		//fmt.Printf("Consumer %d processing item: %v\n", wid, item)

		if !exist {
			// Queue is empty
			time.Sleep(emptyQueueSleep)
		} else {
			consumerFunc(wid, item)
		}
	}
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) GetIndex(torrentLink string) (int, bool) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	for i, info := range dq.queue {
		if info.TorrentLink == torrentLink {
			return i, true
		}
	}

	return 0, false
}

func (dq *DownloadQueue) GetIndexAndReturn(torrentLink string) (*QueueItem, int, bool) {
	//dq.mutex.Lock()
	//defer dq.mutex.Unlock()

	for i, info := range dq.queue {
		if info.TorrentLink == torrentLink {
			return info, i, true
		}
	}

	return nil, 0, false
}

func (dq *DownloadQueue) GetItemOfUser(userId int64) ([]*QueueItem, []int) {
	//unsafe because no lock is used

	res := []*QueueItem{}
	indexes := []int{}

	for i, info := range dq.queue {
		if info.UserId == userId {
			res = append(res, info)
			indexes = append(indexes, i)
		}
	}

	return res, indexes
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) periodicSaveQueue() {
	ticker := time.NewTicker(dq.saveQueueInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dq.doneChan:
			//fmt.Println("periodic save exit")
			return
		case <-ticker.C:
			if dq.operationCount > 0 {
				dq.mutex.Lock()
				dq.saveQueue()
				dq.operationCount = 0
				dq.mutex.Unlock()
			}
		}
	}
}

func (dq *DownloadQueue) checkSave() {
	dq.operationCount++
	if dq.operationCount >= dq.batchSize {
		dq.saveQueue()
		dq.operationCount = 0
	}
}

func (dq *DownloadQueue) saveQueue() {
	data, err := json.Marshal(dq.queue)
	if err != nil {
		errMsg := fmt.Sprintf("Error marshaling queue: %v", err)
		errorHandler.SaveError(errMsg, err)
		return
	}

	err = os.WriteFile(dq.queueFile, data, 0644)
	if err != nil {
		errMsg := fmt.Sprintf("Error saving queue: %v", err)
		errorHandler.SaveError(errMsg, err)
	}
}

func (dq *DownloadQueue) loadQueue() {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	data, err := os.ReadFile(dq.queueFile)
	if err != nil {
		if !os.IsNotExist(err) {
			errMsg := fmt.Sprintf("Error reading queue file: %v", err)
			errorHandler.SaveError(errMsg, err)
		}
		return
	}

	err = json.Unmarshal(data, &dq.queue)
	if err != nil {
		errMsg := fmt.Sprintf("Error unmarshaling queue: %v", err)
		errorHandler.SaveError(errMsg, err)
		return
	}
}

func (dq *DownloadQueue) Close() {
	dq.doneChan <- true
	dq.done = true
	dq.wg.Wait()
	dq.mutex.Lock()
	dq.saveQueue()
	dq.mutex.Unlock()
}
