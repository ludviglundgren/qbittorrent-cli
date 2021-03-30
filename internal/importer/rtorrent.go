package importer

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"

	"github.com/pkg/errors"
	"github.com/zeebo/bencode"
)

type RTorrentImport struct{}

func NewRTorrentImporter() Importer {
	return &RTorrentImport{}
}

var (
	stateFileExtension           = ".rtorrent"
	libtorrentStateFileExtension = ".libtorrent_resume"
)

func (i *RTorrentImport) Import(opts Options) error {
	torrentsSessionDir := opts.RTorrentDir

	info, err := os.Stat(torrentsSessionDir)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "Directory does not exist: %v\n", torrentsSessionDir)
	}
	if err != nil {
		return errors.Wrapf(err, "Directory error: %v\n", torrentsSessionDir)

	}
	if !info.IsDir() {
		return errors.Errorf("Directory is a file, not a directory: %#v\n", torrentsSessionDir)
	}

	matches, _ := filepath.Glob(torrentsSessionDir + "*.torrent")

	totalJobs := len(matches)
	log.Printf("Total torrents to process: %v\n", totalJobs)

	positionNum := 0
	for _, match := range matches {
		positionNum++

		torrentID := getTorrentFileName(match)

		// If file already exists, skip
		if _, err = os.Stat(opts.QbitDir + "/" + torrentID + ".torrent"); err == nil {
			log.Printf("%v/%v Torrent already exists, skipping: %v", positionNum, totalJobs, torrentID)
			continue
		}

		torrentFile, err := torrent.OpenDecodeRaw(match)
		if err != nil {
			log.Printf("Can't decode torrent file %v. Can't decode string %v. Continue", match, torrentID)
			continue
		}

		file, err := torrent.Decode(match)
		if err != nil {
			return err
		}

		// check for FILE.torrent.libtorrent_resume
		resumeFile, err := decodeRTorrentLibTorrentResumeFile(match + libtorrentStateFileExtension)
		if err != nil {
			return err
		}

		// check for FILE.torrent.rtorrent
		rtorrentFile, err := decodeRTorrentFile(match + stateFileExtension)
		if err != nil {
			return err
		}

		newFastResume := qbittorrent.Fastresume{
			ActiveTime:                getActiveTime(rtorrentFile.Custom.SeedingTime),
			AddedTime:                 strToIntClean(rtorrentFile.Custom.AddTime),
			Allocation:                "sparse",
			ApplyIpFilter:             1,
			AutoManaged:               0,
			CompletedTime:             rtorrentFile.TimestampFinished,
			DisableDHT:                0,
			DisableLSD:                0,
			DisablePEX:                0,
			DownloadRateLimit:         -1,
			FileFormat:                "libtorrent resume file",
			FileVersion:               1,
			FilePriority:              []int{},
			FinishedTime:              int64(time.Since(time.Unix(rtorrentFile.TimestampFinished, 0)).Minutes()),
			LastDownload:              0,
			LastSeenComplete:          rtorrentFile.TimestampFinished,
			LastUpload:                0,
			LibTorrentVersion:         "1.2.11.0",
			MaxConnections:            16777215,
			MaxUploads:                -1,
			NumComplete:               16777215,
			NumDownloaded:             16777215,
			NumIncomplete:             0,
			NumPieces:                 int64(len(file.Info.Pieces)) / 20,
			Paused:                    0,
			Peers:                     "",
			Peers6:                    "",
			QbtCategory:               "",
			QbtContentLayout:          "Original",
			QbtFirstLastPiecePriority: 0,
			QbtName:                   "",
			QbtRatioLimit:             -2000,
			QbtSavePath:               rtorrentFile.Directory,
			QbtSeedStatus:             1,
			QbtSeedingTimeLimit:       -2,
			QbtTags:                   []string{},
			SavePath:                  rtorrentFile.Directory,
			SeedMode:                  0,
			SeedingTime:               getActiveTime(rtorrentFile.Custom.SeedingTime),
			SequentialDownload:        0,
			ShareMode:                 0,
			StopWhenReady:             0,
			SuperSeeding:              0,
			TotalDownloaded:           rtorrentFile.TotalDownloaded,
			TotalUploaded:             rtorrentFile.TotalUploaded,
			UploadMode:                0,
			UploadRateLimit:           -1,
			UrlList:                   file.UrlList,

			TorrentFile: torrentFile,
			Path:        rtorrentFile.Directory,
		}

		if file.Info.Files != nil {
			newFastResume.HasFiles = true
		} else {
			newFastResume.HasFiles = false
		}

		// handle trackers
		newFastResume.Trackers = convertTrackers(*resumeFile)

		newFastResume.ConvertFilePriority(file.Info.Files)

		newFastResume.FillPieces()
		//newFastResume.PiecePriority = []byte(newFastResume.Pieces)

		//hash := torrent.CalculateInfoHash(torrentFile)
		newFastResume.InfoHash = newFastResume.GetInfoHashSHA1()

		// only run if not dry-run
		if opts.DryRun != true {
			// copy torrent file
			if err = newFastResume.Encode(opts.QbitDir + "/" + torrentID + ".fastresume"); err != nil {
				log.Printf("Can't create qBittorrent fastresume file %v error: %v", opts.QbitDir+torrentID+".fastresume", err)
				return err
			}

			if err = torrent.CopyFile(match, opts.QbitDir+"/"+torrentID+".torrent"); err != nil {
				log.Printf("Can't create qBittorrent torrent file %v error %v", opts.QbitDir+torrentID+".torrent", err)
				return err
			}
			//go processFiles(torrentID, decodedVal, opts, &torrentsSessionDir, positionNum, totalJobs)
		}

		log.Printf("%v/%v Sucessfully imported: %v %v", positionNum, totalJobs, torrentID, file.Info.Name)

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
func (i *RTorrentImport) processFiles(torrentID string, fastResume NewFastResumeFile, opts Options, torrentsPath *string, position int, totalJobs int) error {
	var err error

	fastResume.torrentFilePath = *torrentsPath + torrentID + ".torrent"
	if _, err = os.Stat(fastResume.torrentFilePath); os.IsNotExist(err) {
		log.Printf("Can't find torrent file %v for %v", fastResume.torrentFilePath, torrentID)
		return err
	}

	fastResume.torrentFile, err = decodeTorrentFile(fastResume.torrentFilePath)
	if err != nil {
		log.Printf("Can't find torrent file %v for %v", fastResume.torrentFilePath, torrentID)
		return err
	}

	if _, ok := fastResume.torrentFile["info"].(map[string]interface{})["files"]; ok {
		fastResume.QbtHasRootFolder = 1
	} else {
		fastResume.QbtHasRootFolder = 0
	}

	//fastResume.QbtQueuePosition = position
	fastResume.QbtQueuePosition = 1
	fastResume.QbtRatioLimit = -2000
	fastResume.QbtSeedStatus = 1
	fastResume.QbtSeedingTimeLimit = -2
	fastResume.QbtTempPathDisabled = 0
	fastResume.QbtName = ""
	fastResume.QbtHasRootFolder = 0

	fastResume.QbtSavePath = fastResume.SavePath
	// TODO handle replace paths

	if err = encodeFastResumeFile(opts.QbitDir+"/"+torrentID+".fastresume", &fastResume); err != nil {
		log.Printf("Can't create qBittorrent fastresume file %v error: %v", opts.QbitDir+torrentID+".fastresume", err)
		return err
	}

	if err = copyFile(fastResume.torrentFilePath, opts.QbitDir+"/"+torrentID+".torrent"); err != nil {
		log.Printf("Can't create qBittorrent torrent file %v error %v", opts.QbitDir+torrentID+".torrent", err)
		return err
	}

	log.Printf("%v/%v Sucessfully imported: %v", position, totalJobs, fastResume.torrentFile["info"].(map[string]interface{})["name"].(string))

	return nil
}

// Takes id.rtorrent custom.seedingtime and converts to int64
func getActiveTime(t string) int64 {
	return int64(time.Since(time.Unix(strToIntClean(t), 0)).Minutes())
}

// convertTrackers from rtorrent file spec to qBittorrent fastresume
func convertTrackers(trackers RTorrentLibTorrentResumeFile) [][]string {
	var ret [][]string

	for url, status := range trackers.Trackers {
		// skip if dht
		if url == "dht://" {
			continue
		}

		if status["enabled"] == 1 {
			ret = append(ret, []string{url})
		}
	}

	return ret
}

// getTorrentFileName from file. Removes file extension
func getTorrentFileName(file string) string {
	_, fileName := filepath.Split(file)
	trimmed := strings.TrimSuffix(fileName, path.Ext(fileName))
	toLower := strings.ToLower(trimmed)

	return toLower
}

func decodeRTorrentLibTorrentResumeFile(path string) (*RTorrentLibTorrentResumeFile, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//var torrent map[string]interface{}
	var torrent RTorrentLibTorrentResumeFile
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return &torrent, nil
}

func decodeRTorrentFile(path string) (*RTorrentTorrentFile, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//var torrent map[string]interface{}
	var torrent RTorrentTorrentFile
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return &torrent, nil
}

// Clean and convert string to int from rtorrent.custom.addtime, seedingtime
func strToIntClean(line string) int64 {
	s := strings.TrimSuffix(line, "\n")
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

type RTorrentLibTorrentResumeFile struct {
	Trackers map[string]map[string]int `bencode:"trackers"`
}

type RTorrentTorrentFile struct {
	Custom struct {
		AddTime     string `bencode:"addtime"`
		SeedingTime string `bencode:"seedingtime"`
	} `bencode:"custom"`
	Directory         string `bencode:"directory"`
	TotalDownloaded   int64  `bencode:"total_downloaded"`
	TotalUploaded     int64  `bencode:"total_uploaded"`
	TimestampFinished int64  `bencode:"timestamp.finished"`
	TimestampStarted  int64  `bencode:"timestamp.started"`
}
