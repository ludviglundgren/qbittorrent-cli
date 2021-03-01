package importer

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/zeebo/bencode"
)

type Options struct {
	DelugeDir string
	QbitDir   string
	StateDir  string
}

type DelugeImporter interface {
	Import(opts Options) error
}

type DelugeImport struct{}

func NewDelugeImporter() DelugeImporter {
	return &DelugeImport{}
}

func (di *DelugeImport) Import(opts Options) error {
	torrentsPath := opts.DelugeDir + "/state/"
	if _, err := os.Stat(torrentsPath); os.IsNotExist(err) {
		log.Println("Can't find deluge state directory")
		return err
	}

	resumeFilePath := opts.DelugeDir + "/state/torrents.fastresume"
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
	log.Printf("total jobs: %v\n", totalJobs)

	positionNum := 0
	for key, value := range fastresumeFile {
		positionNum++
		var decodedVal NewTorrentStructure

		if err := bencode.DecodeString(value.(string), &decodedVal); err != nil {
			torrentFile := map[string]interface{}{}
			torrentFilePath := opts.DelugeDir + "/state/" + key + ".torrent"

			if _, err = os.Stat(torrentFilePath); os.IsNotExist(err) {
				log.Printf("Can't find torrent file %v. Can't decode string %v. Continue", torrentFilePath, key)
				continue
			}
			torrentFile, err = decodeTorrentFile(torrentFilePath)
			if err != nil {
				log.Printf("Can't decode torrent file %v. Can't decode string %v. Continue", torrentFilePath, key)
				continue
			}
			torrentName := torrentFile["info"].(map[string]interface{})["name"].(string)
			log.Printf("Can't decode row %v with torrent %v. Continue", key, torrentName)
		}

		processFile(key, decodedVal, opts, &torrentsPath, positionNum)
	}

	return nil
}

func processFile(key string, newStructure NewTorrentStructure, opts Options, torrentsPath *string, position int) error {
	var err error

	newStructure.torrentFilePath = *torrentsPath + key + ".torrent"
	if _, err = os.Stat(newStructure.torrentFilePath); os.IsNotExist(err) {
		log.Printf("Can't find torrent file %v for %v", newStructure.torrentFilePath, key)
		return err
	}

	newStructure.torrentFile, err = decodeTorrentFile(newStructure.torrentFilePath)
	if err != nil {
		log.Printf("Can't find torrent file %v for %v", newStructure.torrentFilePath, key)
		return err
	}
	newStructure.QbtQueuePosition = position
	newStructure.QbtQueuePosition = 1
	newStructure.QbtRatioLimit = -2000
	newStructure.QbtSeedStatus = 1
	newStructure.QbtSeedingTimeLimit = -2
	newStructure.QbtTempPathDisabled = 0
	newStructure.QbtName = ""
	newStructure.QbtHasRootFolder = 0

	newStructure.QbtSavePath = newStructure.SavePath

	if err = encodeTorrentFile(opts.QbitDir+key+".fastresume", &newStructure); err != nil {
		log.Printf("Can't create qBittorrent fastresume file %v", opts.QbitDir+key+".fastresume")
		return err
	}

	if err = copyFile(newStructure.torrentFilePath, opts.QbitDir+key+".torrent"); err != nil {
		log.Printf("Can't create qBittorrent torrent file %v", opts.QbitDir+key+".torrent")
		return err
	}

	log.Printf("Sucessfully imported %v", newStructure.torrentFile["info"].(map[string]interface{})["name"].(string))

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

func encodeTorrentFile(path string, newStructure *NewTorrentStructure) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Create(path)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferedWriter := bufio.NewWriter(file)
	enc := bencode.NewEncoder(bufferedWriter)
	if err := enc.Encode(newStructure); err != nil {
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

type NewTorrentStructure struct {
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
	pieceLenght     int64
	replace         []Replace
}

type Replace struct {
	from, to string
}
