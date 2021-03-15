package importer

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeebo/bencode"
)

type RTorrentImport struct{}

func NewRTorrentImporter() Importer {
	return &RTorrentImport{}
}

func (i *RTorrentImport) Import(opts Options) error {
	torrentsSessionDir := opts.RTorrentDir
	//dirPath := torrentsSessionDir

	//info, err := os.Stat("../../test/import")
	//myDir, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Print(myDir)

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

	//if stat, err := os.Stat(torrentsSessionDir); err == nil && stat.IsDir() {
	//	// path is a directory
	//
	//	log.Println("path is dir")
	//}

	matches, _ := filepath.Glob(torrentsSessionDir + "*.torrent")

	fmt.Println(len(matches))

	totalJobs := len(matches)

	positionNum := 0
	for _, match := range matches {
		fmt.Printf("file: %v\n", match)

		var decodedVal NewFastResumeFile

		// TODO grab current directory and feed torrent via qbittorrent api to add fast resume data

		//_, fileName := filepath.Split(match)
		//fmt.Printf("filename: %v\n", fileName)

		torrentID := getTorrentName(match)
		fmt.Printf("torrentID: %v\n", torrentID)

		// If file already exists, skip
		//if _, err = os.Stat(opts.QbitDir + "/" + torrentID + ".torrent"); err == nil {
		//	log.Printf("Torrent already exists, skipping: %v", torrentID)
		//	continue
		//}

		file, err := decodeTorrentMetaFile(match)
		if err != nil {
			return err
		}
		fmt.Println(file)

		// check for FILE.torrent.libtorrent_resume
		//resumeFile, err := decodeRTorrentLibTorrentResumeFile(match + ".libtorrent_resume")
		//if err != nil {
		//	return err
		//}
		//fmt.Println(resumeFile)
		// convert to qbit spec
		// only grab top level tracker keys

		// check for FILE.torrent.rtorrent
		rtorrentFile, err := decodeRTorrentFile(match + ".rtorrent")
		if err != nil {
			return err
		}
		fmt.Println(rtorrentFile)

		// create new fast resume file
		decodedVal.AddedTime = strToIntClean(rtorrentFile.Custom.AddTime)
		decodedVal.SeedingTime = strToIntClean(rtorrentFile.Custom.SeedingTime)
		//decodedVal.CompletedTime = rtorrentFile.CompletedTime

		decodedVal.TotalUploaded = rtorrentFile.TotalUploaded
		decodedVal.TotalDownloaded = rtorrentFile.TotalDownloaded

		decodedVal.SavePath = rtorrentFile.Directory
		decodedVal.Trackers = file.AnnounceList

		// handle trackers

		// copy torrent file

		go processFiles(torrentID, decodedVal, opts, &torrentsSessionDir, positionNum, totalJobs)
		time.Sleep(100 * time.Millisecond)
	}

	//fastResume.torrentFilePath = *torrentsPath + torrentID + ".torrent"
	//if _, err = os.Stat(fastResume.torrentFilePath); os.IsNotExist(err) {
	//	log.Printf("Can't find torrent file %v for %v", fastResume.torrentFilePath, torrentID)
	//	return err
	//}

	//resumeFilePath := torrentsSessionDir + "torrents.fastresume"
	//if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
	//	log.Println("Can't find deluge fastresume file")
	//	return err
	//}

	//fastresumeFile, err := decodeTorrentFile(resumeFilePath)
	//if err != nil {
	//	log.Println("Can't decode deluge fastresume file")
	//	return err
	//}

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

func getTorrentName(file string) string {

	_, fileName := filepath.Split(file)
	fmt.Printf("filename: %v\n", fileName)

	return strings.TrimSuffix(fileName, path.Ext(fileName))
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

func decodeTorrentMetaFile(path string) (*TorrentMetaFile, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var torrent TorrentMetaFile
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return &torrent, nil
}

func strToIntClean(line string) int64 {
	s := strings.TrimSuffix(line, "\n")
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

type TorrentMetaFile struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	CreatedBy    string     `bencode:"created by"`
	CreationDate int64      `bencode:"creation date"`
	Info         struct {
		Length      int64  `bencode:"length"`
		Name        string `bencode:"name"`
		PieceLength int64  `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	} `bencode:"info"`
	UrlList []string `bencode:"url-list"`
}

type RTorrentLibTorrentResumeFile struct {
	Trackers map[string]map[string]int
}

type RTorrentTorrentFile struct {
	//ActiveTime         int64     `bencode:"active_time"`
	//AddedTime          int64     `bencode:"added_time"`
	//AnnounceToDht      int64     `bencode:"announce_to_dht"`
	//AnnounceToLsd      int64     `bencode:"announce_to_lsd"`
	//AnnounceToTrackers int64     `bencode:"announce_to_trackers"`
	//AutoManaged        int64     `bencode:"auto_managed"`
	//BannedPeers        string    `bencode:"banned_peers"`
	//BannedPeers6       string    `bencode:"banned_peers6"`
	//BlockPerPiece      int64     `bencode:"blocks per piece"`
	//CompletedTime int64 `bencode:"completed_time"`
	//DownloadRateLimit  int64     `bencode:"download_rate_limit"`
	//FileSizes          [][]int64 `bencode:"file sizes"`
	//FileFormat         string    `bencode:"file-format"`
	//FileVersion        int64     `bencode:"file-version"`
	//FilePriority       []int     `bencode:"file_priority"`
	//FinishedTime       int64     `bencode:"finished_time"`
	//InfoHash           string    `bencode:"info-hash"`
	//LastSeenComplete   int64     `bencode:"last_seen_complete"`
	//LibtorrentVersion  string    `bencode:"libtorrent-version"`
	//MaxConnections     int64     `bencode:"max_connections"`
	//MaxUploads         int64     `bencode:"max_uploads"`
	//NumDownloaded      int64     `bencode:"num_downloaded"`
	//NumIncomplete      int64     `bencode:"num_incomplete"`
	//MappedFiles        []string  `bencode:"mapped_files,omitempty"`
	//Paused             int64     `bencode:"paused"`
	//Peers              string    `bencode:"peers"`
	//Peers6             string    `bencode:"peers6"`
	//Pieces             []byte    `bencode:"pieces"`

	Custom struct {
		AddTime     string `bencode:"addtime"`
		SeedingTime string `bencode:"seedingtime"`
		//XFileName   string `bencode:"x-filename"`
	} `bencode:"custom"`

	TotalUploaded int64  `bencode:"total_uploaded"`
	Directory     string `bencode:"directory"`

	//SavePath           string `bencode:"save_path"`
	//SeedMode           int64  `bencode:"seed_mode"`
	SeedingTime int64 `bencode:"seeding_time"`
	//SequentialDownload int64  `bencode:"sequential_download"`
	//SuperSeeding       int64  `bencode:"super_seeding"`
	TotalDownloaded int64 `bencode:"total_downloaded"`
	//TotalUploaded      int64          `bencode:"total_uploaded"`
	//Trackers        [][]string     `bencode:"trackers"`
	//UploadRateLimit int64          `bencode:"upload_rate_limit"`
	//Unfinished      *[]interface{} `bencode:"unfinished,omitempty"`
	//
	//hasFiles        bool
	//torrentFilePath string
	//torrentFile     map[string]interface{}
	//path            string
	//fileSizes       int64
	//sizeAndPrio     [][]int64
	//torrentFileList []string
	//nPieces         int64
	//pieceLength     int64
	//replace         []Replace
}
