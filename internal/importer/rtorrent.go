package importer

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"

	"github.com/anacrolix/torrent/metainfo"
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
	torrentsSessionDir := opts.SourceDir

	info, err := os.Stat(torrentsSessionDir)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "Directory does not exist: %s", torrentsSessionDir)
	}
	if err != nil {
		return errors.Wrapf(err, "Directory error: %s", torrentsSessionDir)

	}
	if !info.IsDir() {
		return errors.Errorf("Directory is a file, not a directory: %s", torrentsSessionDir)
	}

	matches, _ := filepath.Glob(path.Join(torrentsSessionDir, "*.torrent"))

	totalJobs := len(matches)
	log.Printf("Total torrents to process: %v", totalJobs)

	positionNum := 0
	for _, match := range matches {
		positionNum++

		torrentID := getTorrentFileName(match)

		torrentOutFile := path.Join(opts.QbitDir, torrentID+".torrent")

		// If file already exists, skip
		if _, err = os.Stat(torrentOutFile); err == nil {
			log.Printf("%d/%d %s Torrent already exists, skipping", positionNum, totalJobs, torrentOutFile)
			continue
		}

		torrentFile, err := torrent.OpenDecodeRaw(match)
		if err != nil {
			log.Printf("Could not decode torrent file %s. Could not decode string %s. Continue", match, torrentID)
			continue
		}

		file, err := metainfo.LoadFromFile(match)
		if err != nil {
			return err
		}
		metaInfo, err := file.UnmarshalInfo()
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
			NumPieces:                 int64(metaInfo.NumPieces()),
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
			SeedingTime:               0,
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

		//if file.Info.Files != nil {
		if metaInfo.Files != nil {
			newFastResume.HasFiles = true

			// valid QbtContentLayout = Original, Subfolder, NoSubfolder
			newFastResume.QbtContentLayout = "Original"
			// legacy and should be removed sometime with 4.3.X
			newFastResume.QbtHasRootFolder = 1

			// Fix savepath for torrents with subfolder
			// directory contains the whole torrent path, which gives error in qBit.
			// remove file.info.name from full path in id.rtorrent directory
			newPath := strings.ReplaceAll(rtorrentFile.Directory, metaInfo.Name, "")

			newFastResume.Path = newPath
			newFastResume.SavePath = newPath
			newFastResume.QbtSavePath = newPath
		} else {
			// if only single file then use NoSubfolder
			newFastResume.HasFiles = false

			newFastResume.QbtContentLayout = "NoSubfolder"
			newFastResume.QbtHasRootFolder = 0
		}

		// handle trackers
		newFastResume.Trackers = convertTrackers(*resumeFile)

		newFastResume.ConvertFilePriority(len(metaInfo.Files))

		// fill pieces to set as completed
		newFastResume.FillPieces()

		// Set 20 byte SHA1 hash
		newFastResume.InfoHash = newFastResume.GetInfoHashSHA1()

		// only run if not dry-run
		if opts.DryRun != true {
			// copy torrent file
			fastResumeOutFile := path.Join(opts.QbitDir, torrentID+".fastresume")
			if err = newFastResume.Encode(fastResumeOutFile); err != nil {
				log.Printf("Could not create qBittorrent fastresume file %s error: %q", fastResumeOutFile, err)
				return err
			}

			if err = torrent.CopyFile(match, torrentOutFile); err != nil {
				log.Printf("Could copy qBittorrent torrent file %v error %v", torrentOutFile, err)
				return err
			}
		}

		log.Printf("%d/%d %s Sucessfully imported: %s", positionNum, totalJobs, torrentID, metaInfo.Name)
	}

	return nil
}

// Takes id.rtorrent custom.seedingtime and converts to int64
func getActiveTime(t string) int64 {
	return int64(time.Since(time.Unix(strToIntClean(t), 0)).Seconds())
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
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var torrentResumeFile RTorrentLibTorrentResumeFile
	if err := bencode.DecodeBytes(dat, &torrentResumeFile); err != nil {
		return nil, err
	}

	return &torrentResumeFile, nil
}

func decodeRTorrentFile(path string) (*RTorrentTorrentFile, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var torrentFile RTorrentTorrentFile
	if err := bencode.DecodeBytes(dat, &torrentFile); err != nil {
		return nil, err
	}

	return &torrentFile, nil
}

// Clean and convert string to int from rtorrent.custom.addtime, seedingtime
func strToIntClean(line string) int64 {
	if line == "" {
		return 0
	}

	s := strings.TrimSuffix(line, "\n")
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
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
