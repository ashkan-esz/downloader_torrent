package service

import (
	"context"
	"downloader_torrent/configs"
	"downloader_torrent/internal/repository"
	"downloader_torrent/model"
	errorHandler "downloader_torrent/pkg/error"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/djherbis/times"
	torretParser "github.com/j-muller/go-torrent-parser"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITorrentService interface {
	DownloadQueueConsumer(wid int, queueItem QueueItem)
	DownloadFile(movieId string, torrentUrl string, source EnqueueSource) (*model.DownloadingFile, error)
	CancelDownload(fileName string) error
	RemoveDownload(fileName string) error
	GetTorrentStatus() *model.TorrentStatusRes
	GetDownloadingFiles() []*model.DownloadingFile
	GetLocalFiles() []*model.LocalFile
	UpdateDownloadingFiles(done <-chan bool)
	UpdateLocalFiles(done <-chan bool)
	CleanUp(done <-chan bool)
	SyncFilesAndDb(done <-chan bool)
	AutoDownloader(done <-chan bool)
	HandleAutoDownloader() error
	UpdateStats(done <-chan bool)
	GetDiskSpaceUsage() (int64, error)
	CleanUpSpace() error
	DownloadTorrentMetaFile(url string, location string) (string, error)
	RemoveTorrentMetaFile(metaFileName string) error
	RemoveExpiredLocalFiles() error
	RemoveInvalidLocalLinkFromDb() error
	RemoveLocalFilesNotFoundInDb() error
	RemoveIncompleteDownloadFiles() error
	RemoveOrphanTorrentMetaFiles() error
	CheckConcurrentServingLimit() bool
	CheckServingLocalFile(filename string) bool
	IncrementFileDownloadCount(filename string) error
	DecrementFileDownloadCount(filename string)
	IsTorrentFile(filename string, size int64) (bool, error)
	GetDownloadLink(filename string) string
	GetFileExpireTime(filename string, info os.FileInfo) time.Time
	GetTorrentFileExpireDelay(size int64) time.Duration
	ExtendLocalFileExpireTime(filename string) (time.Time, error)
}

type TorrentService struct {
	torrentRepo              repository.ITorrentRepository
	torrentClient            *torrent.Client
	downloadDir              string
	downloadingFiles         []*model.DownloadingFile
	downloadingFilesMux      *sync.Mutex
	localFiles               []*model.LocalFile
	localFilesMux            *sync.Mutex
	stats                    *model.Stats
	tasks                    *model.Tasks
	activeDownloadsCounts    int64
	activeDownloadsCountsMux *sync.Mutex
	downloadQueue            *DownloadQueue
}

var TorrentSvc *TorrentService

func NewTorrentService(torrentRepo repository.ITorrentRepository) *TorrentService {
	config := torrent.NewDefaultClientConfig()
	config.DataDir = "./downloads"
	config.Debug = false
	config.NoUpload = true
	torrentClient, _ := torrent.NewClient(config)
	//defer torrentClient.Close()

	service := &TorrentService{
		torrentRepo:         torrentRepo,
		torrentClient:       torrentClient,
		downloadDir:         "./downloads/",
		downloadingFiles:    make([]*model.DownloadingFile, 0),
		downloadingFilesMux: &sync.Mutex{},
		localFiles:          make([]*model.LocalFile, 0),
		localFilesMux:       &sync.Mutex{},
		stats: &model.Stats{
			Configs: &model.StatsConfigs{
				DownloadFileSizeLimitMb: 512,
				TorrentDownloadDisabled: true,
			},
			RemainingSpaceMb:          1024,
			TorrentDownloadTimeoutMin: 30,
		},
		tasks:                    &model.Tasks{},
		activeDownloadsCounts:    0,
		activeDownloadsCountsMux: &sync.Mutex{},
		downloadQueue:            nil,
	}

	done := make(chan bool)
	go service.UpdateStats(done)

	for service.stats.Configs.TorrentDownloadConcurrencyLimit == 0 {
		time.Sleep(time.Second)
	}

	count := int(service.stats.Configs.TorrentDownloadConcurrencyLimit)
	downloadQueue := NewDownloadQueue("download_queue.json", count, count*100, 30*time.Second, 10)

	downloadQueue.Start(service.DownloadQueueConsumer, 5*time.Second)
	//todo : add checkDequeue func

	service.downloadQueue = downloadQueue

	go service.UpdateDownloadingFiles(done)
	go service.UpdateLocalFiles(done)
	go service.CleanUp(done)
	go service.SyncFilesAndDb(done)
	go service.AutoDownloader(done)

	TorrentSvc = service

	return service
}

var torrentMetaFileRegex = regexp.MustCompile(`([.\-])torrent`)
var torrentFileRegex = regexp.MustCompile(`([.\-])torrent$`)

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) GetTorrentStatus() *model.TorrentStatusRes {
	return &model.TorrentStatusRes{
		DownloadingFiles:      m.downloadingFiles,
		LocalFiles:            m.localFiles,
		Stats:                 m.stats,
		Tasks:                 m.tasks,
		ActiveDownloadsCounts: m.activeDownloadsCounts,
		TorrentClientStats:    m.torrentClient.Stats(),
		DownloadQueueStats:    m.downloadQueue.GetStats(),
	}
}

//------------------------------------------
//------------------------------------------

func (m *TorrentService) DownloadQueueConsumer(wid int, queueItem QueueItem) {
	m.WaitForDownloadConcurrencyFree()
	downloadFile, err := m.DownloadFile(queueItem.TitleId, queueItem.TorrentLink, queueItem.EnqueueSource)
	if err != nil {
		if errors.Is(err, model.ErrTorrentLinkNotFound) || errors.Is(err, model.ErrFileSizeExceeded) || errors.Is(err, model.ErrEmptyFile) {
			_ = m.RemoveAutoDownloaderLink(queueItem)
		}
		if !errors.Is(err, model.ErrMaximumDiskUsageExceeded) && !errors.Is(err, model.ErrFileSizeExceeded) {
			// don't save no-space-error
			errorHandler.SaveError(err.Error(), err)
		}
		m.WaitForDownloadConcurrencyFree()
		return
	}

	if downloadFile != nil {
		//fmt.Println(downloadFile.Name, err)
		//wait until download is done or got error
		for !downloadFile.Done {
			time.Sleep(1 * time.Second)
		}
		if downloadFile.Error != nil {
			errorHandler.SaveError(downloadFile.Error.Error(), downloadFile.Error)
			m.WaitForDownloadConcurrencyFree()
			return
		}
	}

	// doesn't need if downloaded successfully
	//_ = m.RemoveAutoDownloaderLink(queueItem)

	_ = m.HandleExpireTimeOfAutoDownloaded(downloadFile)

	m.WaitForDownloadConcurrencyFree()
}

func (m *TorrentService) RemoveAutoDownloaderLink(queueItem QueueItem) error {
	if queueItem.EnqueueSource == AutoDownloader {
		err := m.torrentRepo.UpdateTorrentAutoDownloaderPullDownloadLink(queueItem.TitleId, queueItem.TorrentLink)
		if err != nil {
			errorHandler.SaveError(err.Error(), err)
		}
		return err
	}
	return nil
}

func (m *TorrentService) WaitForDownloadConcurrencyFree() {
	for len(m.downloadingFiles) >= int(m.stats.Configs.TorrentDownloadConcurrencyLimit) {
		time.Sleep(2 * time.Second)
	}
}

func (m *TorrentService) HandleExpireTimeOfAutoDownloaded(file *model.DownloadingFile) error {
	defer func() {
		if r := recover(); r != nil {
			// Convert the panic to an error
			fmt.Printf("recovered from panic: %v\n", r)
		}
	}()

	diffHour := m.stats.Configs.TorrentFilesExpireHour - configs.GetDbConfigs().DefaultTorrentDownloaderConfig.TorrentFilesExpireHour

	if diffHour == 0 {
		return nil
	}

	var downloadTime time.Time
	t, err := times.Stat(m.downloadDir + file.Name)
	if err == nil {
		downloadTime = t.BirthTime()
		if downloadTime.Before(t.AccessTime()) {
			downloadTime = t.AccessTime()
		}
	}
	if downloadTime.IsZero() {
		fileInfo, err := os.Stat(m.downloadDir + file.Name)
		if err == nil {
			downloadTime = fileInfo.ModTime()
		} else {
			downloadTime = time.Now()
		}
	}

	newAccessTime := downloadTime.Add(time.Duration(diffHour) * time.Hour)

	err = os.Chtimes(m.downloadDir+file.Name, newAccessTime, newAccessTime)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------
//------------------------------------------

func (m *TorrentService) DownloadFile(movieId string, torrentUrl string, source EnqueueSource) (d *model.DownloadingFile, err error) {
	locked := false
	defer func() {
		if locked {
			m.downloadingFilesMux.Unlock()
		}
		if r := recover(); r != nil {
			// Convert the panic to an error
			err = fmt.Errorf("recovered from panic: %v", r)
			errorHandler.SaveError(err.Error(), err)
		}
	}()

	if m.stats.Configs.TorrentDownloadDisabled {
		return nil, model.ErrTorrentDownloadDisabled
	}

	if m.stats.RemainingSpaceMb < m.stats.Configs.DownloadSpaceThresholdMb && m.stats.Configs.DownloadSpaceThresholdMb > 0 {
		return nil, model.ErrMaximumDiskUsageExceeded
	}

	checkResult, err := m.torrentRepo.CheckTorrentLinkExist(movieId, torrentUrl)
	if err != nil {
		return nil, err
	}
	if checkResult == nil {
		return nil, model.ErrTorrentLinkNotFound
	}

	m.downloadingFilesMux.Lock()
	locked = true
	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].TorrentUrl == torrentUrl {
			locked = false
			m.downloadingFilesMux.Unlock()
			return nil, model.ErrAlreadyDownloading
		}
	}

	// not needed, concurrency handled by queue-dequeue workers
	//if int64(len(m.downloadingFiles)) >= m.stats.Configs.TorrentDownloadConcurrencyLimit {
	//	locked = false
	//	m.downloadingFilesMux.Unlock()
	//	return nil, model.ErrTorrentDownloadConcurrencyLimit
	//}

	d = &model.DownloadingFile{
		State:          "started",
		Name:           "",
		MetaFileName:   "",
		Size:           0,
		DownloadedSize: 0,
		TorrentUrl:     torrentUrl,
		Torrent:        nil,
		TitleId:        movieId,
		TitleName:      checkResult.Title,
		TitleType:      checkResult.Type,
		StartTime:      time.Now(),
		Error:          nil,
		Done:           false,
	}

	m.downloadingFiles = append(m.downloadingFiles, d)
	locked = false
	m.downloadingFilesMux.Unlock()

	downloadDone := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.stats.TorrentDownloadTimeoutMin)*time.Minute)

	go func() {
		defer cancel()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		lastProgress := time.Now()

		for {
			select {
			case <-ctx.Done():
				//fmt.Println("Download timed out due to inactivity.")
				if d != nil {
					d.Torrent.Drop()
					d.Error = model.ErrTorrentDownloadTimeout
					d.Done = true
				}
				return
			case <-downloadDone:
				//fmt.Println("Downloading file stopped/completed")
				if d != nil {
					d.Done = true
				}
				return
			case <-ticker.C:
				if d != nil && d.Torrent != nil {
					downloadedBytes := d.Torrent.BytesCompleted()
					if downloadedBytes > d.DownloadedSize {
						lastProgress = time.Now()
					}

					if time.Since(lastProgress) > 2*time.Minute {
						//fmt.Println("No progress detected for 2 minute(s). Timing out.")
						d.Torrent.Drop()
						d.Error = model.ErrTorrentDownloadInactive
						d.Done = true
						return
					}

					d.DownloadedSize = downloadedBytes
					if d.Size > 0 && d.Size == d.DownloadedSize {
						if d.MetaFileName != "" {
							_ = m.RemoveTorrentMetaFile(d.MetaFileName)
						}

						// sample: download.movieTracker.site/downloads/ttt.mkv
						localUrl := m.GetDownloadLink(d.Name)
						expireTime := m.GetFileExpireTime(d.Name, nil).UnixMilli()
						err = m.torrentRepo.SaveTorrentLocalLink(movieId, checkResult.Type, torrentUrl, localUrl, expireTime)

						d.Done = true
						return
					}
				}
			}
		}
	}()

	defer func() {
		if err != nil {
			downloadDone <- true
			if d != nil {
				d.Error = err

				if !errors.Is(err, model.ErrFileAlreadyExist) && d.Name != "" {
					_ = m.removeTorrentFile(d.Name, false)
				}

				if d.MetaFileName != "" {
					_ = m.RemoveTorrentMetaFile(d.MetaFileName)
				}
			}
		}
	}()

	var t *torrent.Torrent
	if strings.HasPrefix(torrentUrl, "magent:?") {
		d.State = "adding magnet"
		t, err = m.torrentClient.AddMagnet(torrentUrl)
	} else {
		d.State = "downloading torrent meta file"
		d.MetaFileName, err = m.DownloadTorrentMetaFile(torrentUrl, m.downloadDir)
		if err != nil {
			return d, err
		}
		d.State = "adding torrent"
		t, err = m.torrentClient.AddTorrentFromFile(m.downloadDir + d.MetaFileName)
	}
	d.State = "getting info"
	if err != nil {
		return d, err
	}
	<-t.GotInfo()

	d.Name = t.Info().Name
	d.Size = t.Info().Length
	d.Torrent = t

	err = m.CheckFileExistAndSize(d)
	if err != nil {
		return d, err
	}

	d.State = "downloading"
	t.DownloadAll()

	return d, err
}

func (m *TorrentService) CheckFileExistAndSize(d *model.DownloadingFile) error {
	for _, lf := range m.localFiles {
		if lf.Name == d.Name {
			return model.ErrFileAlreadyExist
		}
	}

	if d.Size == 0 {
		return model.ErrEmptyFile
	}

	if d.Size > m.stats.Configs.DownloadFileSizeLimitMb*1024*1024 && m.stats.Configs.DownloadFileSizeLimitMb > 0 {
		//m := fmt.Sprintf("File size exceeds the limit (%vmb)", m.stats.Configs.DownloadFileSizeLimitMb)
		return model.ErrFileSizeExceeded
	}

	if d.Size > (m.stats.RemainingSpaceMb-200)*1024*1024 && m.stats.RemainingSpaceMb > 0 {
		//m := fmt.Sprintf("Not Enough space left (%vmb/%vmb)", d.Size/(1024*1024), m.stats.RemainingSpaceMb-200)
		return model.ErrNotEnoughSpace
	}

	return nil
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) CancelDownload(fileName string) error {
	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].Name == fileName {
			m.downloadingFiles[i].Torrent.Drop()
			_ = m.removeTorrentFile(m.downloadingFiles[i].Name, false)
			break
		}
	}

	m.downloadingFilesMux.Lock()
	defer m.downloadingFilesMux.Unlock()

	m.downloadingFiles = slices.DeleteFunc(m.downloadingFiles, func(d *model.DownloadingFile) bool {
		return d.Name == fileName
	})
	return nil
}

func (m *TorrentService) RemoveDownload(fileName string) error {
	if _, err := os.Stat("./downloads/" + fileName); errors.Is(err, os.ErrNotExist) {
		return err
	}

	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].Name == fileName {
			m.downloadingFiles[i].Torrent.Drop()
			break
		}
	}

	m.downloadingFilesMux.Lock()
	defer m.downloadingFilesMux.Unlock()
	m.downloadingFiles = slices.DeleteFunc(m.downloadingFiles, func(d *model.DownloadingFile) bool {
		return d.Name == fileName
	})

	_ = m.removeTorrentFile(fileName, false)
	if strings.HasSuffix(fileName, ".mkv") {
		// remove converted file
		mp4File := strings.Replace(fileName, ".mkv", ".mp4", 1)
		if _, err := os.Stat("./downloads/" + fileName); err == nil {
			_ = m.removeTorrentFile(mp4File, false)
		}
	}

	return nil
}

func (m *TorrentService) ExtendLocalFileExpireTime(filename string) (time.Time, error) {
	if m.stats.Configs.TorrentFileExpireExtendHour == 0 {
		return time.Time{}, model.ErrExtendLocalFileExpireIsDisabled
	}

	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	defer func() {
		if r := recover(); r != nil {
			// Convert the panic to an error
			fmt.Printf("recovered from panic: %v\n", r)
		}
	}()

	for _, lf := range m.localFiles {
		if lf.Name == filename {

			var downloadTime time.Time
			t, err := times.Stat(m.downloadDir + filename)
			if err == nil {
				downloadTime = t.BirthTime()
				if downloadTime.Before(t.AccessTime()) {
					downloadTime = t.AccessTime()
				}
			}
			if downloadTime.IsZero() {
				fileInfo, err := os.Stat(m.downloadDir + filename)
				if err == nil {
					downloadTime = fileInfo.ModTime()
				} else {
					downloadTime = time.Now()
				}
			}

			newTime := lf.ExpireTime.Add(time.Duration(m.stats.Configs.TorrentFileExpireExtendHour) * time.Hour)
			newAccessTime := downloadTime.Add(time.Duration(m.stats.Configs.TorrentFileExpireExtendHour) * time.Hour)

			err = os.Chtimes(m.downloadDir+filename, newAccessTime, newAccessTime)
			if err != nil {
				return lf.ExpireTime, err
			}

			return newTime, nil
		}
	}

	return time.Time{}, model.ErrFileNotFound
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) UpdateDownloadingFiles(done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			m.GetDownloadingFiles()
		}
	}
}

func (m *TorrentService) UpdateLocalFiles(done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			m.GetLocalFiles()
		}
	}
}

func (m *TorrentService) CleanUp(done <-chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			_ = m.RemoveIncompleteDownloadFiles()
			_ = m.RemoveOrphanTorrentMetaFiles()
			_ = m.RemoveExpiredLocalFiles()
		}
	}
}

func (m *TorrentService) SyncFilesAndDb(done <-chan bool) {
	ticker := time.NewTicker(60 * time.Minute)
	defer ticker.Stop()

	time.Sleep(30 * time.Second)
	_ = m.RemoveInvalidLocalLinkFromDb()
	_ = m.RemoveLocalFilesNotFoundInDb()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			_ = m.RemoveInvalidLocalLinkFromDb()
			_ = m.RemoveLocalFilesNotFoundInDb()
		}
	}
}

func (m *TorrentService) AutoDownloader(done <-chan bool) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	time.Sleep(1 * time.Minute)
	_ = m.HandleAutoDownloader()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			_ = m.HandleAutoDownloader()
		}
	}
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) HandleAutoDownloader() error {
	docs, err := m.torrentRepo.GetTorrentAutoDownloaderLinks()
	if err != nil {
		errorHandler.SaveError(err.Error(), err)
		return err
	}

	removedCounter := 0
	enqueueCounter := 0
	for _, doc := range docs {
		for _, removeLink := range doc.RemoveTorrentLinks {
			m.tasks.AutoDownloader = fmt.Sprintf("handling: %v removed, %v enqueued", removedCounter, enqueueCounter)
			_ = m.torrentRepo.UpdateTorrentAutoDownloaderPullRemoveLink(doc.Type, removeLink)
			removedCounter++
		}
	}

	for _, doc := range docs {
		for _, downloadLink := range doc.DownloadTorrentLinks {
			_, exist := m.downloadQueue.GetIndex(downloadLink)
			if !exist {
				qItem := QueueItem{
					TitleId:       doc.Id.Hex(),
					TitleType:     doc.Type,
					TorrentLink:   downloadLink,
					EnqueueTime:   time.Now(),
					EnqueueSource: AutoDownloader,
					UserInfo:      nil,
				}
				enqueueCounter++
				m.tasks.AutoDownloader = fmt.Sprintf("handling: %v removed, %v enqueued", removedCounter, enqueueCounter)
				_, err = m.downloadQueue.Enqueue(qItem)
				if errors.Is(err, ErrOverflow) {
					break
				}
			}
		}
	}

	m.tasks.AutoDownloader = fmt.Sprintf("%v removed, %v enqueued", removedCounter, enqueueCounter)

	if removedCounter > 0 {
		_ = m.RemoveLocalFilesNotFoundInDb()
	}

	return nil
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) GetDownloadingFiles() []*model.DownloadingFile {
	m.downloadingFilesMux.Lock()
	defer m.downloadingFilesMux.Unlock()

	m.downloadingFiles = slices.DeleteFunc(m.downloadingFiles, func(df *model.DownloadingFile) bool {
		return (df.Size > 0 && df.Size == df.DownloadedSize) || df.Error != nil
	})
	return m.downloadingFiles
}

func (m *TorrentService) GetLocalFiles() []*model.LocalFile {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	dirs, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return make([]*model.LocalFile, 0)
	}

A:
	for _, dir := range dirs {
		filename := dir.Name()
		if strings.Contains(filename, ".torrent.") || torrentFileRegex.MatchString(filename) || dir.IsDir() {
			continue
		}
		for i2 := range m.downloadingFiles {
			if m.downloadingFiles[i2].Name == filename {
				// still downloading
				continue A
			}
		}

		for i2 := range dirs {
			if strings.Contains(dirs[i2].Name(), fmt.Sprintf("-%v.torrent", filename)) {
				// incomplete download
				continue A
			}
		}

		info, err := dir.Info()
		if err != nil {
			continue
		}

		for i := range m.localFiles {
			if m.localFiles[i].Name == filename {
				// already existed
				m.localFiles[i].ExpireTime = m.GetFileExpireTime(filename, info) //just in case expire time changed
				continue A
			}
		}

		td := int64(0)
		ad := int64(0)
		// new file
		f := &model.LocalFile{
			Name:             filename,
			Size:             info.Size(),
			DownloadLink:     m.GetDownloadLink(filename),
			StreamLink:       "/v1/stream/" + filename,
			ExpireTime:       m.GetFileExpireTime(filename, info),
			TotalDownloads:   &td,
			ActiveDownloads:  &ad,
			LastDownloadTime: time.Time{},
		}
		m.localFiles = append(m.localFiles, f)
	}

	m.localFiles = slices.DeleteFunc(m.localFiles, func(lf *model.LocalFile) bool {
		for i := range dirs {
			if lf.Name == dirs[i].Name() {
				return false
			}
		}
		return true
	})

	return m.localFiles
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) UpdateStats(done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			totalFilesSize, _ := m.GetDiskSpaceUsage()

			dbConfigs := configs.GetDbConfigs()
			stats := &model.Stats{
				Configs: &model.StatsConfigs{
					DownloadSpaceLimitMb:                dbConfigs.TorrentDownloadMaxSpaceSize,
					DownloadFileSizeLimitMb:             dbConfigs.TorrentDownloadMaxFileSize,
					DownloadSpaceThresholdMb:            dbConfigs.TorrentDownloadSpaceThresholdSize,
					TorrentFilesExpireHour:              dbConfigs.TorrentFilesExpireHour,
					TorrentFilesServingConcurrencyLimit: dbConfigs.TorrentFilesServingConcurrencyLimit,
					TorrentDownloadConcurrencyLimit:     dbConfigs.TorrentDownloadConcurrencyLimit,
					TorrentFilesServingDisabled:         dbConfigs.TorrentFilesServingDisabled,
					TorrentDownloadDisabled:             dbConfigs.TorrentDownloadDisabled,
					TorrentFileExpireDelayFactor:        dbConfigs.TorrentFileExpireDelayFactor,
					TorrentFileExpireExtendHour:         dbConfigs.TorrentFileExpireExtendHour,
				},
				TotalFilesSizeMb:          totalFilesSize / (1024 * 1024), //mb
				TorrentDownloadTimeoutMin: dbConfigs.TorrentDownloadTimeoutMin,
			}

			localFilesSize := int64(0)
			for _, file := range m.localFiles {
				localFilesSize += file.Size / (1024 * 1024)
			}
			downloadFilesSize := int64(0)
			downloadFilesCurrentSize := int64(0)
			for _, file := range m.downloadingFiles {
				downloadFilesSize += file.Size / (1024 * 1024)
				downloadFilesCurrentSize += file.DownloadedSize / (1024 * 1024)
			}

			stats.LocalFilesSizeMb = localFilesSize
			stats.DownloadingFilesFinalSizeMb = downloadFilesSize
			stats.DownloadingFilesCurrentSizeMb = downloadFilesCurrentSize

			maxUsedSpace := maxInt64(stats.TotalFilesSizeMb, localFilesSize+downloadFilesSize)
			stats.RemainingSpaceMb = stats.Configs.DownloadSpaceLimitMb - maxUsedSpace

			m.stats = stats

			_ = m.CleanUpSpace()

		}
	}
}

func (m *TorrentService) GetDiskSpaceUsage() (int64, error) {
	// includes the current size of downloading files

	files, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				return 0, err
			}
			totalSize += info.Size()
		}
	}
	return totalSize, nil
}

func (m *TorrentService) CleanUpSpace() error {
	m.tasks.DiskSpaceCleaner = "handling"
	defer func() {
		m.tasks.DiskSpaceCleaner = ""
	}()

	lruActivationThreshold := minInt64(10*1024, 4*m.stats.Configs.DownloadSpaceThresholdMb)

	if m.stats.RemainingSpaceMb < lruActivationThreshold {
		m.localFilesMux.Lock()

		// sort by expire time
		slices.SortFunc(m.localFiles, func(a, b *model.LocalFile) int {
			compareExpire := a.ExpireTime.Compare(b.ExpireTime)
			if compareExpire != 0 {
				return compareExpire
			}
			return 0
		})

		freedSpace := int64(0)
		for i, lf := range m.localFiles {
			if i > len(m.localFiles) || (m.stats.RemainingSpaceMb+freedSpace >= lruActivationThreshold) {
				break
			}

			if *lf.TotalDownloads == 0 {
				freedSpace += lf.Size / (1024 * 1024)
				_ = m.removeTorrentFile(lf.Name, true)
			}
		}

		// sort by last download time
		slices.SortFunc(m.localFiles, func(a, b *model.LocalFile) int {
			compareExpire := a.LastDownloadTime.Compare(b.LastDownloadTime)
			if compareExpire != 0 {
				return compareExpire
			}
			if *a.TotalDownloads < *b.TotalDownloads {
				return -1
			}
			if *a.TotalDownloads > *b.TotalDownloads {
				return 1
			}
			return 0
		})

		for i, lf := range m.localFiles {
			if i > len(m.localFiles) || (m.stats.RemainingSpaceMb+freedSpace >= lruActivationThreshold) {
				break
			}

			if *lf.ActiveDownloads == 0 {
				err := m.removeTorrentFile(lf.Name, true)
				if err != nil {
					freedSpace += lf.Size
				}
			}
		}

		m.localFilesMux.Unlock()
	}

	return nil
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) DownloadTorrentMetaFile(url string, location string) (string, error) {
	metaFileName := strings.Replace(url, "https://", "", 1)
	metaFileName = strings.ReplaceAll(metaFileName, "/", "-")

	out, err := os.Create(location + metaFileName)
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		errorMessage := fmt.Sprintf("Error on downloading torrent meta: %s", err)
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Error on downloading torrent meta: %v", fmt.Errorf("bad status: %s", resp.Status))
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		errorMessage := fmt.Sprintf("Error on saving torrent meta file: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return "", err
	}

	return metaFileName, nil
}

func (m *TorrentService) RemoveTorrentMetaFile(metaFileName string) error {
	err := os.Remove(m.downloadDir + metaFileName)
	if err != nil {
		if !os.IsNotExist(err) {
			errorMessage := fmt.Sprintf("Error on removing torrent meta file: %s", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
	return err
}

func (m *TorrentService) removeTorrentFile(filename string, updateDb bool) error {
	err := os.Remove(m.downloadDir + filename)
	if err != nil {
		if !os.IsNotExist(err) {
			errorMessage := fmt.Sprintf("Error on removing torrent file: %s", err)
			errorHandler.SaveError(errorMessage, err)
		}
	} else if updateDb {
		_ = m.torrentRepo.RemoveTorrentLocalLink("serial", m.GetDownloadLink(filename))
	}
	return err
}

func (m *TorrentService) RemoveOrphanTorrentMetaFiles() error {
	dirs, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return err
	}

	removedCounter := int64(0)
A:
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}

		filename := dir.Name()
		if strings.Contains(filename, ".torrent.db") {
			continue
		}

		fileInfo, err := dir.Info()
		if err == nil {
			isTorrent, _ := m.IsTorrentFile(filename, fileInfo.Size())
			if isTorrent {
				elapsedMin := time.Now().Sub(fileInfo.ModTime()).Minutes()
				if elapsedMin > 5 {
					for _, df := range m.downloadingFiles {
						// check its downloading
						if df.MetaFileName == filename {
							continue A
						}
					}
					removedCounter++
					m.tasks.OrphanMetaFilesRemover = fmt.Sprintf("handling: %v removed", removedCounter)
					_ = m.RemoveTorrentMetaFile(filename)
				}
			}
		}
	}

	m.tasks.OrphanMetaFilesRemover = fmt.Sprintf("%v removed", removedCounter)

	return err
}

func (m *TorrentService) RemoveIncompleteDownloadFiles() error {
	dirs, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return err
	}

	removedCounter := int64(0)
A:
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}

		filename := dir.Name()
		if strings.Contains(filename, ".torrent.db") {
			continue
		}

		fileInfo, err := dir.Info()
		if err == nil {
			isTorrent, _ := m.IsTorrentFile(filename, fileInfo.Size())
			if !isTorrent {
				for _, df := range m.downloadingFiles {
					// check its downloading
					if df.Name == filename {
						continue A
					}
				}
				for _, dir2 := range dirs {
					if strings.Contains(dir2.Name(), fmt.Sprintf("-%v.torrent", filename)) {
						// torrent-meta-file exist --> incomplete download
						removedCounter++
						m.tasks.InCompleteDownloadsRemover = fmt.Sprintf("handling: %v removed", removedCounter)
						err = m.removeTorrentFile(filename, false)
						if err == nil {
							err = m.RemoveTorrentMetaFile(dir2.Name())
						}
					}
				}
			}
		}
	}

	m.tasks.InCompleteDownloadsRemover = fmt.Sprintf("%v removed", removedCounter)

	return err
}

func (m *TorrentService) RemoveExpiredLocalFiles() error {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	removedCounter := int64(0)
	for _, lf := range m.localFiles {
		if time.Now().After(lf.ExpireTime) && *lf.ActiveDownloads == 0 && time.Now().After(lf.LastDownloadTime.Add(m.GetTorrentFileExpireDelay(lf.Size))) {
			// file is expired
			removedCounter++
			_ = m.removeTorrentFile(lf.Name, true)
		}
		m.tasks.ExpiredFilesRemover = fmt.Sprintf("handling: %v removed", removedCounter)
	}
	m.tasks.ExpiredFilesRemover = fmt.Sprintf("%v removed", removedCounter)

	return nil
}

func (m *TorrentService) RemoveInvalidLocalLinkFromDb() error {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	m.tasks.DbsInvalidLocalLinksRemover = "handling serials: started"
	serialsCursor, err := m.torrentRepo.GetAllSerialTorrentLocalLinks()
	defer func() {
		if serialsCursor != nil {
			_ = serialsCursor.Close(context.TODO())
		}
		m.tasks.DbsInvalidLocalLinksRemover = ""
	}()

	if err != nil {
		errorHandler.SaveError(err.Error(), err)
		return err
	}

	_ = m.HandleTorrentLinksRemoveCursor(serialsCursor, "serial")
	m.tasks.DbsInvalidLocalLinksRemover = "handling serials: ended"

	//-----------------------------------

	m.tasks.DbsInvalidLocalLinksRemover = "handling movies: started"
	moviesCursor, err := m.torrentRepo.GetAllMovieTorrentLocalLinks()
	defer func() {
		if moviesCursor != nil {
			_ = moviesCursor.Close(context.TODO())
		}
	}()

	if err != nil {
		errorHandler.SaveError(err.Error(), err)
		return err
	}

	_ = m.HandleTorrentLinksRemoveCursor(moviesCursor, "movie")
	m.tasks.DbsInvalidLocalLinksRemover = "handling movies: ended"

	return nil
}

func (m *TorrentService) RemoveLocalFilesNotFoundInDb() error {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	defer func() {
		m.tasks.InvalidLocalFilesRemover = ""
	}()

	arr := []string{}
	checkedCounter := int64(0)
	removedCounter := int64(0)
	for i, lf := range m.localFiles {
		checkedCounter++
		arr = append(arr, lf.DownloadLink)
		if len(arr) > 19 || i == len(m.localFiles)-1 {
			serialLocalLinks, err := m.torrentRepo.FindSerialTorrentLinks(arr)
			if err != nil {
				errorHandler.SaveError(err.Error(), err)
				return err
			}

			movieLocalLinks, err := m.torrentRepo.FindMovieTorrentLinks(arr)
			if err != nil {
				errorHandler.SaveError(err.Error(), err)
				return err
			}

			serialLocalLinks = append(serialLocalLinks, movieLocalLinks...)

			for _, l := range arr {
				if !slices.Contains(serialLocalLinks, l) {
					temp := strings.Split(l, "/")
					filename := temp[len(temp)-1]
					removedCounter++
					_ = m.removeTorrentFile(filename, false)
				}
			}

			m.tasks.InvalidLocalFilesRemover = fmt.Sprintf("handling: %v checked, %v removed", checkedCounter, removedCounter)

			arr = []string{}
		}
	}

	return nil
}

func (m *TorrentService) HandleTorrentLinksRemoveCursor(cursor *mongo.Cursor, movieType string) error {
	if cursor == nil {
		return nil
	}

	var doc map[string]string
	checkCounter := int64(0)
	removeCounter := int64(0)
	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&doc)
		if err != nil {
			errorHandler.SaveError(err.Error(), err)
			continue
		}
		checkCounter++

		temp := strings.Split(doc["localLink"], "/")
		filename := temp[len(temp)-1]

		fileExist := false
		for _, lf := range m.localFiles {
			if lf.Name == filename {
				fileExist = true
				break
			}
		}
		if !fileExist {
			removeCounter++
			_ = m.torrentRepo.RemoveTorrentLocalLink(movieType, doc["localLink"])
		}

		m.tasks.DbsInvalidLocalLinksRemover = fmt.Sprintf("handling %vs: %v checked, %v removed", movieType, checkCounter, removeCounter)
	}

	return nil
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) CheckConcurrentServingLimit() bool {
	m.activeDownloadsCountsMux.Lock()
	defer m.activeDownloadsCountsMux.Unlock()

	limit := m.stats.Configs.TorrentFilesServingConcurrencyLimit

	return m.activeDownloadsCounts < limit || limit == 0
}

func (m *TorrentService) IncrementFileDownloadCount(filename string) error {
	m.activeDownloadsCountsMux.Lock()
	m.activeDownloadsCounts++
	defer m.activeDownloadsCountsMux.Unlock()

	for _, lf := range m.localFiles {
		if lf.Name == filename {
			go func() {
				_ = m.torrentRepo.IncrementTorrentLinkDownload("serial", m.GetDownloadLink(filename))
			}()

			td := atomic.AddInt64(lf.TotalDownloads, 1)
			ad := atomic.AddInt64(lf.ActiveDownloads, 1)
			lf.TotalDownloads = &td
			lf.ActiveDownloads = &ad
			lf.LastDownloadTime = time.Now()
			return nil
		}
	}
	return model.ErrFileNotFound
}

func (m *TorrentService) DecrementFileDownloadCount(filename string) {
	m.activeDownloadsCountsMux.Lock()
	m.activeDownloadsCounts--
	defer m.activeDownloadsCountsMux.Unlock()

	for _, lf := range m.localFiles {
		if lf.Name == filename {
			ad := atomic.AddInt64(lf.ActiveDownloads, -1)
			lf.ActiveDownloads = &ad
		}
	}
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) CheckServingLocalFile(filename string) bool {
	return !m.stats.Configs.TorrentFilesServingDisabled
}

func (m *TorrentService) GetDownloadLink(filename string) string {
	return configs.GetConfigs().ServerAddress + "/partial_download/" + filename
}

func (m *TorrentService) GetFileExpireTime(filename string, info os.FileInfo) time.Time {
	defer func() {
		if r := recover(); r != nil {
			// Convert the panic to an error
			fmt.Printf("recovered from panic: %v\n", r)
		}
	}()

	expireHour := m.stats.Configs.TorrentFilesExpireHour
	if expireHour == 0 {
		expireHour = 48
	}

	var downloadTime time.Time
	t, err := times.Stat(m.downloadDir + filename)
	if err == nil {
		downloadTime = t.BirthTime()
		if downloadTime.Before(t.AccessTime()) {
			downloadTime = t.AccessTime()
		}
	}
	if downloadTime.IsZero() {
		if info != nil {
			downloadTime = info.ModTime()
		} else {
			fileInfo, err := os.Stat(m.downloadDir + filename)
			if err == nil {
				downloadTime = fileInfo.ModTime()
			}
		}
	}

	return downloadTime.Add(time.Duration(expireHour) * time.Hour)
}

func (m *TorrentService) GetTorrentFileExpireDelay(size int64) time.Duration {
	return time.Duration(float32(size/(1024*1024))*m.stats.Configs.TorrentFileExpireDelayFactor) * time.Second
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) IsTorrentFile(filename string, size int64) (bool, error) {
	if torrentMetaFileRegex.MatchString(filename) {
		return true, nil
	}

	// > 512kb
	if size > 512*1024 {
		return false, nil
	}

	// Open the file
	//file, err := os.Open(m.downloadDir + filename)
	//if err != nil {
	//	return false, err
	//}
	//defer file.Close()

	//metaInfo, err := torretParser.ParseFromFile(m.downloadDir + filename)
	_, err := torretParser.ParseFromFile(m.downloadDir + filename)
	if err != nil {
		return false, err
	}

	// Check for required keys
	//if metaInfo.Announce == "" || len(metaInfo.Files) == 0 {
	//	return false, nil // Missing required keys
	//}

	return true, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
