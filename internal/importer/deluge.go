package importer

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/zeebo/bencode"
)

type Options struct {
	SourceDir   string
	QbitDir     string
	RTorrentDir string
	DryRun      bool
}

type Importer interface {
	Import(opts Options) error
}

type DelugeImport struct{}

func NewDelugeImporter() Importer {
	return &DelugeImport{}
}

func (di *DelugeImport) Import(opts Options) error {
	torrentsStateDir := opts.SourceDir + "/state/"
	if _, err := os.Stat(torrentsStateDir); os.IsNotExist(err) {
		log.Println("Can't find deluge state directory")
		return err
	}

	resumeFilePath := torrentsStateDir + "torrents.fastresume"
	if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
		log.Println("Can't find deluge fastresume file")
		return err
	}

	fastresumeFile, err := decodeTorrentFile(resumeFilePath)
	if err != nil {
		log.Println("Can't decode deluge fastresume file")
		return err
	}

	totalJobs := len(fastresumeFile)
	log.Printf("Total torrents to process: %v\n", totalJobs)

	positionNum := 0
	for torrentID, value := range fastresumeFile {
		positionNum++
		var decodedVal NewFastResumeFile

		// If file already exists, skip
		if _, err = os.Stat(opts.QbitDir + "/" + torrentID + ".torrent"); err == nil {
			log.Printf("Torrent already exists, skipping: %v", torrentID)
			continue
		}

		// TODO check if torrent exist in fastresume data, but not on disk

		if err := bencode.DecodeString(value.(string), &decodedVal); err != nil {
			torrentFile := map[string]interface{}{}
			torrentFilePath := torrentsStateDir + torrentID + ".torrent"

			if _, err = os.Stat(torrentFilePath); os.IsNotExist(err) {
				log.Printf("Can't find torrent file %v. Can't decode string %v. Continue", torrentFilePath, torrentID)
				continue
			}
			torrentFile, err = decodeTorrentFile(torrentFilePath)
			if err != nil {
				log.Printf("Can't decode torrent file %v. Can't decode string %v. Continue", torrentFilePath, torrentID)
				continue
			}
			torrentName := torrentFile["info"].(map[string]interface{})["name"].(string)
			log.Printf("Can't decode row %v with torrent %v. Continue", torrentID, torrentName)
		}

		go processFiles(torrentID, decodedVal, opts, &torrentsStateDir, positionNum, totalJobs)
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func processFiles(torrentID string, fastResume NewFastResumeFile, opts Options, torrentsPath *string, position int, totalJobs int) error {
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

func decodeTorrentFile(path string) (map[string]interface{}, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var torrent map[string]interface{}
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}
	return torrent, nil
}

func encodeFastResumeFile(path string, newStructure *NewFastResumeFile) error {

	_, err2 := os.Create(path)
	if err2 != nil {
		log.Printf("os create error: %v", err2)
		return err2
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("os open file error: %v", err)
		return err
	}
	defer file.Close()
	bufferedWriter := bufio.NewWriter(file)
	enc := bencode.NewEncoder(bufferedWriter)
	if err := enc.Encode(newStructure); err != nil {
		log.Printf("encode error: %v", err)
		return err
	}
	bufferedWriter.Flush()
	return nil
}

func copyFile(src string, dst string) error {
	originalFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer originalFile.Close()
	newFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer newFile.Close()
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return err
	}
	if err := newFile.Sync(); err != nil {
		return err
	}
	return nil
}

type NewFastResumeFile struct {
	ActiveTime         int64     `bencode:"active_time"`
	AddedTime          int64     `bencode:"added_time"`
	AnnounceToDht      int64     `bencode:"announce_to_dht"`
	AnnounceToLsd      int64     `bencode:"announce_to_lsd"`
	AnnounceToTrackers int64     `bencode:"announce_to_trackers"`
	AutoManaged        int64     `bencode:"auto_managed"`
	BannedPeers        string    `bencode:"banned_peers"`
	BannedPeers6       string    `bencode:"banned_peers6"`
	BlockPerPiece      int64     `bencode:"blocks per piece"`
	CompletedTime      int64     `bencode:"completed_time"`
	DownloadRateLimit  int64     `bencode:"download_rate_limit"`
	FileSizes          [][]int64 `bencode:"file sizes"`
	FileFormat         string    `bencode:"file-format"`
	FileVersion        int64     `bencode:"file-version"`
	FilePriority       []int     `bencode:"file_priority"`
	FinishedTime       int64     `bencode:"finished_time"`
	InfoHash           string    `bencode:"info-hash"`
	LastSeenComplete   int64     `bencode:"last_seen_complete"`
	LibtorrentVersion  string    `bencode:"libtorrent-version"`
	MaxConnections     int64     `bencode:"max_connections"`
	MaxUploads         int64     `bencode:"max_uploads"`
	NumDownloaded      int64     `bencode:"num_downloaded"`
	NumIncomplete      int64     `bencode:"num_incomplete"`
	MappedFiles        []string  `bencode:"mapped_files,omitempty"`
	Paused             int64     `bencode:"paused"`
	Peers              string    `bencode:"peers"`
	Peers6             string    `bencode:"peers6"`
	Pieces             []byte    `bencode:"pieces"`

	QbtHasRootFolder    int64    `bencode:"qBt-hasRootFolder"`
	QbtCategory         string   `bencode:"qBt-category,omitempty"`
	QbtName             string   `bencode:"qBt-name"`
	QbtQueuePosition    int      `bencode:"qBt-queuePosition"`
	QbtRatioLimit       int64    `bencode:"qBt-ratioLimit"`
	QbtSavePath         string   `bencode:"qBt-savePath"`
	QbtSeedStatus       int64    `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit int64    `bencode:"qBt-seedingTimeLimit"`
	QbtTags             []string `bencode:"qBt-tags"`
	QbtTempPathDisabled int64    `bencode:"qBt-tempPathDisabled"`

	SavePath           string         `bencode:"save_path"`
	SeedMode           int64          `bencode:"seed_mode"`
	SeedingTime        int64          `bencode:"seeding_time"`
	SequentialDownload int64          `bencode:"sequential_download"`
	SuperSeeding       int64          `bencode:"super_seeding"`
	TotalDownloaded    int64          `bencode:"total_downloaded"`
	TotalUploaded      int64          `bencode:"total_uploaded"`
	Trackers           [][]string     `bencode:"trackers"`
	UploadRateLimit    int64          `bencode:"upload_rate_limit"`
	Unfinished         *[]interface{} `bencode:"unfinished,omitempty"`

	hasFiles        bool
	torrentFilePath string
	torrentFile     map[string]interface{}
	path            string
	fileSizes       int64
	sizeAndPrio     [][]int64
	torrentFileList []string
	nPieces         int64
	pieceLength     int64
	replace         []Replace
}

type Replace struct {
	from, to string
}
