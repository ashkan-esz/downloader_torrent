package service

import (
	"downloader_torrent/configs"
	"downloader_torrent/internal/repository"
	"downloader_torrent/model"
	errorHandler "downloader_torrent/pkg/error"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
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
	UpdateDiskInfo(done <-chan bool)
	GetDiskSpaceUsage() (int64, error)
	DownloadTorrentMetaFile(url string, location string) (string, error)
	RemoveTorrentMetaFile(metaFileName string) error
}

type TorrentService struct {
	torrentRepo         repository.ITorrentRepository
	torrentClient       *torrent.Client
	downloadDir         string
	downloadingFiles    []*model.DownloadingFile
	downloadingFilesMux *sync.Mutex
	localFiles          []*model.LocalFile
	localFilesMux       *sync.Mutex
	diskInfo            *model.DiskInfo
}

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
		diskInfo:            &model.DiskInfo{},
	}

	done := make(chan bool)
	go service.UpdateDownloadingFiles(done)
	go service.UpdateLocalFiles(done)
	go service.UpdateDiskInfo(done)

	return service
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) GetTorrentStatus() *model.TorrentStatusRes {
	return &model.TorrentStatusRes{
		DownloadingFiles: m.downloadingFiles,
		LocalFiles:       m.localFiles,
		DiskInfo:         m.diskInfo,
	}
}

//------------------------------------------
//------------------------------------------

func (m *TorrentService) DownloadFile(movieId string, torrentUrl string) (d *model.DownloadingFile, err error) {
	diskUsage, err := m.GetDiskSpaceUsage()
	if err != nil {
		return nil, err
	}
	if diskUsage >= int64(configs.GetConfigs().MaxDownloadSpaceGb*1024*1024*1024) {
		return nil, errors.New("maximum disk usage exceeded")
	}

	checkResult, err := m.torrentRepo.CheckTorrentLinkExist(movieId, torrentUrl)
	if err != nil {
		return nil, err
	}

	m.downloadingFilesMux.Lock()
	for i := range m.downloadingFiles {
		if m.downloadingFiles[i].TorrentUrl == torrentUrl {
			return nil, errors.New("already downloading")
		}
	}
	d = &model.DownloadingFile{
		State:          "started",
		Name:           "",
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
	defer func() {
		if err != nil {
			downloadDone <- true
			d.Error = err
			_ = m.removeTorrentFile(d.Name)
		}
	}()

	var t *torrent.Torrent
	var metaFileName string
	if strings.HasPrefix(torrentUrl, "magent:?") {
		d.State = "adding magnet"
		t, err = m.torrentClient.AddMagnet(torrentUrl)
	} else {
		d.State = "downloading torrent meta file"
		metaFileName, err = m.DownloadTorrentMetaFile(torrentUrl, m.downloadDir)
		if err != nil {
			return nil, err
		}
		d.State = "adding torrent"
		t, err = m.torrentClient.AddTorrentFromFile(m.downloadDir + metaFileName)
	}
	d.State = "getting info"
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()

	d.Name = t.Info().Name
	d.Size = t.Info().Length
	d.Torrent = t

	//---------------------------------------------
	dbconfig := configs.GetDbConfigs()
	if d.Size == 0 {
		return nil, errors.New("File is empty")
	}

	if d.Size > dbconfig.TorrentDownloadMaxFileSize*1024*1024 {
		m := fmt.Sprintf("File size exceeds the limit (%vmb)", dbconfig.TorrentDownloadMaxFileSize)
		return nil, errors.New(m)
	}
	//---------------------------------------------

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-downloadDone:
				//fmt.Println("Downloading file stopped/completed")
				return
			case <-ticker.C:
				d.DownloadedSize = d.Torrent.BytesCompleted()
				if d.Size == d.DownloadedSize {
					return
				}
			}
		}
	}()

	d.State = "downloading"
	t.DownloadAll()

	if metaFileName != "" {
		err = m.RemoveTorrentMetaFile(metaFileName)
		if err != nil {
			return nil, err
		}
	}

	// sample: download.movieTracker.site/downloads/ttt.mkv
	localUrl := "/downloads/" + d.Name
	err = m.torrentRepo.SaveTorrentLocalLink(movieId, checkResult.Type, torrentUrl, localUrl)
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

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) GetDownloadingFiles() []*model.DownloadingFile {
	m.downloadingFilesMux.Lock()
	defer m.downloadingFilesMux.Unlock()

	m.downloadingFiles = slices.DeleteFunc(m.downloadingFiles, func(df *model.DownloadingFile) bool {
		return df.Size == df.DownloadedSize || df.Error != nil
	})
	return m.downloadingFiles
}

func (m *TorrentService) GetLocalFiles() []*model.LocalFile {
	m.localFilesMux.Lock()
	defer m.localFilesMux.Unlock()

	dir, err := os.ReadDir(m.downloadDir)
	if err != nil {
		return make([]*model.LocalFile, 0)
	}
	localFiles := []*model.LocalFile{}
A:
	for i := range dir {
		if strings.Contains(dir[i].Name(), ".torrent.") || dir[i].IsDir() {
			continue
		}
		for i2 := range m.downloadingFiles {
			if m.downloadingFiles[i2].Name == dir[i].Name() {
				// still downloading
				continue A
			}
		}

		info, err := dir[i].Info()
		if err != nil {
			continue
		}
		f := &model.LocalFile{
			Name:         dir[i].Name(),
			Size:         info.Size(),
			DownloadLink: "/downloads/" + dir[i].Name(),
			StreamLink:   "/v1/stream/" + dir[i].Name(),
		}
		localFiles = append(localFiles, f)
	}

	m.localFiles = localFiles

	return m.localFiles
}

//-----------------------------------------
//-----------------------------------------

func (m *TorrentService) UpdateDiskInfo(done <-chan bool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			//fmt.Println("Periodic task stopped")
			return
		case <-ticker.C:
			totalFilesSize, _ := m.GetDiskSpaceUsage()

			m.diskInfo = &model.DiskInfo{
				TotalFilesSizeMb:      totalFilesSize / (1024 * 10224), //mb
				MaxDownloadSpaceMb:    int64(configs.GetConfigs().MaxDownloadSpaceGb * 1024),
				DownloadMaxFileSizeMb: configs.GetDbConfigs().TorrentDownloadMaxFileSize,
			}
		}
	}
}

func (m *TorrentService) GetDiskSpaceUsage() (int64, error) {
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
		errorMessage := fmt.Sprintf("Error on removing torrent meta file: %s", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

func (m *TorrentService) removeTorrentFile(filename string) error {
	err := os.Remove(m.downloadDir + filename)
	if err != nil {
		errorMessage := fmt.Sprintf("Error on removing torrent file: %s", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//-----------------------------------------
//-----------------------------------------
