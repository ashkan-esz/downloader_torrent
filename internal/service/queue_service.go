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
	done              chan bool
}

func NewDownloadQueue(queueFile string, workers int, capacity int, saveQueueInterval time.Duration, batchSize int) *DownloadQueue {
	dq := &DownloadQueue{
		queue: make([]FileInfo, 0),
		//queue:             make([]FileInfo, 0, capacity),
		queueFile:         queueFile,
		capacity:          capacity,
		workers:           workers,
		saveQueueInterval: saveQueueInterval,
		batchSize:         batchSize,
		operationCount:    0,
		done:              make(chan bool),
	}

	dq.loadQueue()
	go dq.periodicSaveQueue()
	dq.Start()

	return dq
}

//---------------------------------------
//---------------------------------------

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

func (dq *DownloadQueue) Start() {
	// todo :
	//for i := 0; i < q.workers; i++ {
	//	wid := i
	//	go q.worker(wid)
	//}
}

func (dq *DownloadQueue) worker(wid int) {
	// todo :
	//for qItem := range q.queue {
	//	qItem()
	//}
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
		case <-dq.done:
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
	dq.done <- true
	dq.mutex.Lock()
	dq.saveQueue()
	dq.mutex.Unlock()
}
