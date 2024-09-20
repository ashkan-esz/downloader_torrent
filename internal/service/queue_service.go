package service

import (
	errorHandler "downloader_torrent/pkg/error"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type IDownloadQueue interface {
	Enqueue(fileInfo FileInfo) (int, error)
	Dequeue() (FileInfo, bool)
	periodicSaveQueue()
	checkSave()
	saveQueue()
	loadQueue()
	Start(consumerFunc ConsumerFunc, emptyQueueSleep time.Duration)
	worker(wid int, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration)
	Close()
}

type DownloadQueue struct {
	queue             []FileInfo
	mutex             sync.Mutex
	queueFile         string
	capacity          int
	workers           int
	saveQueueInterval time.Duration
	batchSize         int
	operationCount    int
	doneChan          chan bool
	done              bool
	wg                *sync.WaitGroup
}

func NewDownloadQueue(queueFile string, workers int, capacity int, saveQueueInterval time.Duration, batchSize int) *DownloadQueue {
	dq := &DownloadQueue{
		queue:             make([]FileInfo, 0, capacity),
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

	return dq
}

//---------------------------------------
//---------------------------------------

type ConsumerFunc func(wid int, fileInfo FileInfo)

type FileInfo struct {
	URL      string
	FileName string
}

var ErrOverflow = errors.New("overflow")

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) Enqueue(fileInfo FileInfo) (int, error) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	if len(dq.queue) >= dq.capacity {
		return -1, ErrOverflow
	}

	dq.queue = append(dq.queue, fileInfo)

	dq.checkSave()

	return len(dq.queue) - 1, nil
}

func (dq *DownloadQueue) Dequeue() (FileInfo, bool) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	if len(dq.queue) == 0 {
		return FileInfo{}, false
	}

	fileInfo := dq.queue[0]
	dq.queue = dq.queue[1:]

	dq.checkSave()

	return fileInfo, true
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) Start(consumerFunc ConsumerFunc, emptyQueueSleep time.Duration) {
	for i := 0; i < dq.workers; i++ {
		dq.wg.Add(1)
		go dq.worker(i, consumerFunc, emptyQueueSleep)
	}
}

func (dq *DownloadQueue) worker(wid int, consumerFunc ConsumerFunc, emptyQueueSleep time.Duration) {
	defer dq.wg.Done()

	for {
		if dq.done == true {
			//fmt.Println("consumer closed,done ", wid)
			return
		}

		item, exist := dq.Dequeue()

		//fmt.Printf("Consumer %d processing item: %v\n", wid, item)

		if !exist {
			// Queue is empty
			time.Sleep(emptyQueueSleep)
			continue
		}

		// magic/logic happens here
		consumerFunc(wid, item)
	}
}

//---------------------------------------
//---------------------------------------

func (dq *DownloadQueue) GetIndex(url string) (int, bool) {
	dq.mutex.Lock()
	defer dq.mutex.Unlock()

	for i, info := range dq.queue {
		if info.URL == url {
			return i, true
		}
	}

	return 0, false
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
