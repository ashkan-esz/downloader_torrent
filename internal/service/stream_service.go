package service

import (
	"downloader_torrent/configs"
	"downloader_torrent/internal/repository"
	"downloader_torrent/model"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type IStreamService interface {
	GetConvertingFiles() []*model.ConvertingFile
	HandleFileConversion(filename string, noConversion bool, crf int) (string, string, int)
	ConvertMkvToMp4(filename string, crf int) error
}

type StreamService struct {
	movieRepo          repository.IMovieRepository
	downloadDir        string
	convertingFiles    []*model.ConvertingFile
	convertingFilesMux *sync.Mutex
}

func NewStreamService(movieRepo repository.IMovieRepository) *StreamService {
	return &StreamService{
		movieRepo:          movieRepo,
		downloadDir:        "./downloads/",
		convertingFiles:    make([]*model.ConvertingFile, 0),
		convertingFilesMux: &sync.Mutex{},
	}
}

//------------------------------------------
//------------------------------------------

func (m *StreamService) GetConvertingFiles() []*model.ConvertingFile {
	return m.convertingFiles
}

func (m *StreamService) HandleFileConversion(filename string, noConversion bool, crf int) (string, string, int) {
	if strings.HasSuffix(filename, ".mkv") && !noConversion && !configs.GetConfigs().DontConvertMkv {
		if len(m.convertingFiles) > 0 {
			if m.convertingFiles[0].Name == filename {
				errorMessage := fmt.Sprintf("Converting mkv to mp4: %v", m.convertingFiles[0].Progress)
				return filename, errorMessage, fiber.StatusConflict
			} else {
				errorMessage := fmt.Sprintf("Another file [%v] is Converting from mkv to mp4: %v", m.convertingFiles[0].Name, m.convertingFiles[0].Progress)
				return filename, errorMessage, fiber.StatusInternalServerError
			}
		}
		m.convertingFilesMux.Lock()
		convertedName := strings.Replace(filename, ".mkv", ".mp4", 1)
		if _, err := os.Stat("./downloads/" + convertedName); errors.Is(err, os.ErrNotExist) {
			err := m.ConvertMkvToMp4(filename, crf)
			if err != nil {
				m.convertingFilesMux.Unlock()
				errorMessage := fmt.Sprintf("Error on converting mkv to mp4: %v", err)
				return filename, errorMessage, fiber.StatusInternalServerError
			}
		}
		m.convertingFilesMux.Unlock()
		filename = convertedName
	}
	return filename, "", 0
}

func (m *StreamService) ConvertMkvToMp4(filename string, crf int) error {
	defer func() {
		m.convertingFiles = slices.DeleteFunc(m.convertingFiles, func(cf *model.ConvertingFile) bool {
			return cf.Name == filename
		})
	}()

	convFile := &model.ConvertingFile{
		Progress: "0",
		Name:     filename,
		Size:     0,
		Duration: 0,
	}
	m.convertingFiles = append(m.convertingFiles, convFile)

	inputFile := "./downloads/" + filename
	outputFile := "./downloads/" + strings.Replace(filename, ".mkv", ".mp4", 1)

	a, err := ffmpeg.Probe(inputFile)
	if err != nil {
		panic(err)
	}
	totalDuration, err := probeDuration(a)
	if err != nil {
		panic(err)
	}
	convFile.Duration = totalDuration

	err = ffmpeg.Input(inputFile).
		Output(outputFile, ffmpeg.KwArgs{
			"c:v":      "libx264",
			"c:a":      "copy",
			"preset":   "ultrafast",
			"movflags": "+faststart",
			"crf":      crf,
			"tune":     "animation",
		}).
		GlobalArgs("-progress", "unix://"+TempSock(totalDuration, convFile)).
		OverWriteOutput().
		Run()

	if err != nil {
		_ = os.Remove(outputFile)
	}

	return err
}

//------------------------------------------
//------------------------------------------

func TempSock(totalDuration float64, convFile *model.ConvertingFile) string {
	rand.Seed(time.Now().Unix())
	sockFileName := path.Join(os.TempDir(), fmt.Sprintf("%d_sock", rand.Int()))
	l, err := net.Listen("unix", sockFileName)
	if err != nil {
		panic(err)
	}

	go func() {
		re := regexp.MustCompile(`out_time_ms=(\d+)`)
		fd, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		buf := make([]byte, 16)
		data := ""
		progress := ""
		for {
			_, err := fd.Read(buf)
			if err != nil {
				return
			}
			data += string(buf)
			a := re.FindAllStringSubmatch(data, -1)
			cp := ""
			if len(a) > 0 && len(a[len(a)-1]) > 0 {
				c, _ := strconv.Atoi(a[len(a)-1][len(a[len(a)-1])-1])
				cp = fmt.Sprintf("%.2f", float64(c)/totalDuration/1000000)
			}
			if strings.Contains(data, "progress=end") {
				cp = "done"
			}
			if cp == "" {
				cp = ".0"
			}
			if cp != progress {
				progress = cp
				convFile.Progress = progress
			}
		}
	}()

	return sockFileName
}

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
}

func probeDuration(a string) (float64, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
