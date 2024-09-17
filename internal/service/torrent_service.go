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
)

type ITorrentService interface {
	DownloadFile(movieId string, torrentUrl string) (*model.DownloadingFile, error)
	CancelDownload(fileName string) error
	RemoveDownload(fileName string) error
	GetTorrentStatus() *model.TorrentStatusRes
	GetDownloadingFiles() []*model.DownloadingFile
	GetLocalFiles() []*model.LocalFile
	UpdateDownloadingFiles(done <-chan bool)
	UpdateLocalFiles(done <-chan bool)
	CleanUp(done <-chan bool)
	UpdateDiskInfo(done <-chan bool)
	GetDiskSpaceUsage() (int64, error)
	DownloadTorrentMetaFile(url string, location string) (string, error)
	RemoveTorrentMetaFile(metaFileName string) error
	RemoveExpiredLocalFiles() error
	RemoveIncompleteDownloadFiles() error
	RemoveOrphanTorrentMetaFiles() error
	CheckConcurrentServingLimit() bool
	CheckServingLocalFile(filename string) bool
	IncrementFileDownloadCount(filename string) error
	DecrementFileDownloadCount(filename string)
	IsTorrentFile(filename string, size int64) (bool, error)
	GetDownloadLink(filename string) string
	GetFileExpireTime(filename string, info os.FileInfo) time.Time
}

type TorrentService struct {
	torrentRepo              repository.ITorrentRepository
	torrentClient            *torrent.Client
	downloadDir              string
	downloadingFiles         []*model.DownloadingFile
	downloadingFilesMux      *sync.Mutex
	localFiles               []*model.LocalFile
	localFilesMux            *sync.Mutex
	diskInfo                 *model.DiskInfo
	activeDownloadsCounts    int64
	activeDownloadsCountsMux *sync.Mutex
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
		diskInfo: &model.DiskInfo{
			Configs: &model.DiskInfoConfigs{
				DownloadFileSizeLimitMb: 512,
				TorrentDownloadDisabled: true,
			},
			RemainingSpaceMb:          1024,
			TorrentDownloadTimeoutMin: 30,
		},
		activeDownloadsCounts:    0,
		activeDownloadsCountsMux: &sync.Mutex{},
	}

	done := make(chan bool)
	go service.UpdateDiskInfo(done)
	go service.UpdateDownloadingFiles(done)
	go service.UpdateLocalFiles(done)
	go service.CleanUp(done)

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
		DiskInfo:              m.diskInfo,
		ActiveDownloadsCounts: m.activeDownloadsCounts,
		TorrentClientStats:    m.torrentClient.Stats(),
	}
}

//------------------------------------------
//------------------------------------------

func (m *TorrentService) DownloadFile(movieId string, torrentUrl string) (d *model.DownloadingFile, err error) {
	if m.diskInfo.Configs.TorrentDownloadDisabled {
		return nil, model.ErrTorrentDownloadDisabled
	}

	if m.diskInfo.RemainingSpaceMb < m.diskInfo.Configs.DownloadSpaceThresholdMb && m.diskInfo.Configs.DownloadSpaceThresholdMb > 0 {
		return nil, errors.New("maximum disk usage exceeded")
	}

	checkResult, err := m.torrentRepo.CheckTorrentLinkExist(movieId, torrentUrl)
	if err != nil {
		return nil, err
	}

	m.downloadingFilesMux.Lock()
	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].TorrentUrl == torrentUrl {
			m.downloadingFilesMux.Unlock()
			return nil, errors.New("already downloading")
		}
	}

	if int64(len(m.downloadingFiles)) >= m.diskInfo.Configs.TorrentDownloadConcurrencyLimit {
		m.downloadingFilesMux.Unlock()
		return nil, model.ErrTorrentDownloadConcurrencyLimit
	}

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
	}
	m.downloadingFiles = append(m.downloadingFiles, d)
	m.downloadingFilesMux.Unlock()

	downloadDone := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.diskInfo.TorrentDownloadTimeoutMin)*time.Minute)

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
				}
				return
			case <-downloadDone:
				//fmt.Println("Downloading file stopped/completed")
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

				if !errors.Is(err, model.ErrFileAlreadyExist) {
					_ = m.removeTorrentFile(d.Name)
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

	//---------------------------------------------

	for _, lf := range m.localFiles {
		if lf.Name == d.Name {
			return d, model.ErrFileAlreadyExist
		}
	}

	//---------------------------------------------
	if d.Size == 0 {
		return d, errors.New("File is empty")
	}

	if d.Size > m.diskInfo.Configs.DownloadFileSizeLimitMb*1024*1024 && m.diskInfo.Configs.DownloadFileSizeLimitMb > 0 {
		m := fmt.Sprintf("File size exceeds the limit (%vmb)", m.diskInfo.Configs.DownloadFileSizeLimitMb)
		return d, errors.New(m)
	}

	if d.Size > (m.diskInfo.RemainingSpaceMb-200)*1024*1024 && m.diskInfo.RemainingSpaceMb > 0 {
		m := fmt.Sprintf("Not Enough space left (%vmb/%vmb)", d.Size/(1024*1024), m.diskInfo.RemainingSpaceMb-200)
		return d, errors.New(m)
	}
	//---------------------------------------------

	d.State = "downloading"
	t.DownloadAll()

	return d, err
}

func (m *TorrentService) CancelDownload(fileName string) error {
	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].Name == fileName {
			m.downloadingFiles[i].Torrent.Drop()
			_ = m.removeTorrentFile(m.downloadingFiles[i].Name)
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

	_ = m.removeTorrentFile(fileName)
	if strings.HasSuffix(fileName, ".mkv") {
		// remove converted file
		mp4File := strings.Replace(fileName, ".mkv", ".mp4", 1)
		if _, err := os.Stat("./downloads/" + fileName); err == nil {
			_ = m.removeTorrentFile(mp4File)
		}
	}

	localUrl := "/downloads/" + fileName
	// don't know the type
	_ = m.torrentRepo.RemoveTorrentLocalLink("movie", localUrl)
	_ = m.torrentRepo.RemoveTorrentLocalLink("serial", localUrl)

	return nil
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
			Name:            filename,
			Size:            info.Size(),
			DownloadLink:    m.GetDownloadLink(filename),
			StreamLink:      "/v1/stream/" + filename,
			ExpireTime:      m.GetFileExpireTime(filename, info),
			TotalDownloads:  &td,
			ActiveDownloads: &ad,
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

func (m *TorrentService) UpdateDiskInfo(done <-chan bool) {
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
			info := &model.DiskInfo{
				Configs: &model.DiskInfoConfigs{
					DownloadSpaceLimitMb:                dbConfigs.TorrentDownloadMaxSpaceSize,
					DownloadFileSizeLimitMb:             dbConfigs.TorrentDownloadMaxFileSize,
					DownloadSpaceThresholdMb:            dbConfigs.TorrentDownloadSpaceThresholdSize,
					TorrentFilesExpireHour:              dbConfigs.TorrentFilesExpireHour,
					TorrentFilesServingConcurrencyLimit: dbConfigs.TorrentFilesServingConcurrencyLimit,
					TorrentDownloadConcurrencyLimit:     dbConfigs.TorrentDownloadConcurrencyLimit,
					TorrentFilesServingDisabled:         dbConfigs.TorrentFilesServingDisabled,
					TorrentDownloadDisabled:             dbConfigs.TorrentDownloadDisabled,
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

			info.LocalFilesSizeMb = localFilesSize
			info.DownloadingFilesFinalSizeMb = downloadFilesSize
			info.DownloadingFilesCurrentSizeMb = downloadFilesCurrentSize

			maxUsedSpace := maxInt64(info.TotalFilesSizeMb, localFilesSize+downloadFilesSize)
			info.RemainingSpaceMb = info.Configs.DownloadSpaceLimitMb - maxUsedSpace

			m.diskInfo = info

			//todo : handle when theres no space and running downloads
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

func (m *TorrentService) removeTorrentFile(filename string) error {
	err := os.Remove(m.downloadDir + filename)
	if err != nil {
		if !os.IsNotExist(err) {
			errorMessage := fmt.Sprintf("Error on removing torrent file: %s", err)
			errorHandler.SaveError(errorMessage, err)
		}
	}
	return err
}

func (m *TorrentService) RemoveOrphanTorrentMetaFiles() error {
	dirs, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return err
	}

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
					_ = m.RemoveTorrentMetaFile(filename)
				}
			}
		}
	}

	return err
}

func (m *TorrentService) RemoveIncompleteDownloadFiles() error {
	dirs, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return err
	}

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
						err = m.removeTorrentFile(filename)
						if err == nil {
							err = m.RemoveTorrentMetaFile(dir2.Name())
						}
					}
				}
			}
		}
	}

	return err
}

func (m *TorrentService) RemoveExpiredLocalFiles() error {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	for _, lf := range m.localFiles {
		if time.Now().After(lf.ExpireTime) && *lf.ActiveDownloads == 0 {
			// file is expired
			_ = m.removeTorrentFile(lf.Name)
		}
	}

	return nil
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) CheckConcurrentServingLimit() bool {
	m.activeDownloadsCountsMux.Lock()
	defer m.activeDownloadsCountsMux.Unlock()

	limit := m.diskInfo.Configs.TorrentFilesServingConcurrencyLimit

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
			return nil
		}
	}
	return errors.New("file not found")
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
	return !m.diskInfo.Configs.TorrentFilesServingDisabled
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

	expireHour := m.diskInfo.Configs.TorrentFilesExpireHour
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
