package importer

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/zeebo/bencode"
)

type DelugeImporter interface {
	Import(resumeFilePath string) error
}

type DelugeImport struct{}

func NewDelugeImporter() DelugeImporter {
	return &DelugeImport{}
}

func (di *DelugeImport) Import(resumeFilePath string) error {
	/*
		torrentspath := flags.delugedir + "state" + sep
		if _, err := os.Stat(torrentspath); os.IsNotExist(err) {
			log.Println("Can't find deluge state directory")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
		resumefilepath := flags.delugedir + "state" + sep + "torrents.fastresume"
		if _, err := os.Stat(resumefilepath); os.IsNotExist(err) {
			log.Println("Can't find deluge fastresume file")
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}
	*/

	//fastresumeFile, err := decodetorrentfile("./test/import/deluge/torrents.fastresume")
	fastresumeFile, err := decodetorrentfile(resumeFilePath)
	if err != nil {
		log.Println("Can't decode deluge fastresume file")
		time.Sleep(30 * time.Second)
		os.Exit(1)
	}

	totalJobs := len(fastresumeFile)
	log.Printf("total jobs: %v\n", totalJobs)

	for _, value := range fastresumeFile {
		var decodedval NewTorrentStructure

		if err := bencode.DecodeString(value.(string), &decodedval); err != nil {
			//torrentfile := map[string]interface{}{}
			log.Printf("could not decode file: %+v", decodedval)
		}
		//log.Printf("file: %+v", decodedval)
	}

	return nil
}

func decodetorrentfile(path string) (map[string]interface{}, error) {
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

type NewTorrentStructure struct {
	ActiveTime          int64          `bencode:"active_time"`
	AddedTime           int64          `bencode:"added_time"`
	AnnounceToDht       int64          `bencode:"announce_to_dht"`
	AnnounceToLsd       int64          `bencode:"announce_to_lsd"`
	AnnounceToTrackers  int64          `bencode:"announce_to_trackers"`
	AutoManaged         int64          `bencode:"auto_managed"`
	BannedPeers         string         `bencode:"banned_peers"`
	BannedPeers6        string         `bencode:"banned_peers6"`
	BlockPerPiece       int64          `bencode:"blocks per piece"`
	CompletedTime       int64          `bencode:"completed_time"`
	DownloadRateLimit   int64          `bencode:"download_rate_limit"`
	FileSizes           [][]int64      `bencode:"file sizes"`
	FileFormat          string         `bencode:"file-format"`
	FileVersion         int64          `bencode:"file-version"`
	FilePriority        []int          `bencode:"file_priority"`
	FinishedTime        int64          `bencode:"finished_time"`
	InfoHash            string         `bencode:"info-hash"`
	LastSeenComplete    int64          `bencode:"last_seen_complete"`
	LibtorrentVersion   string         `bencode:"libtorrent-version"`
	MaxConnections      int64          `bencode:"max_connections"`
	MaxUploads          int64          `bencode:"max_uploads"`
	NumDownloaded       int64          `bencode:"num_downloaded"`
	NumIncomplete       int64          `bencode:"num_incomplete"`
	MappedFiles         []string       `bencode:"mapped_files,omitempty"`
	Paused              int64          `bencode:"paused"`
	Peers               string         `bencode:"peers"`
	Peers6              string         `bencode:"peers6"`
	Pieces              []byte         `bencode:"pieces"`
	QbthasRootFolder    int64          `bencode:"qBt-hasRootFolder"`
	Qbtcategory         string         `bencode:"qBt-category,omitempty"`
	Qbtname             string         `bencode:"qBt-name"`
	QbtqueuePosition    int            `bencode:"qBt-queuePosition"`
	QbtratioLimit       int64          `bencode:"qBt-ratioLimit"`
	QbtsavePath         string         `bencode:"qBt-savePath"`
	QbtseedStatus       int64          `bencode:"qBt-seedStatus"`
	QbtseedingTimeLimit int64          `bencode:"qBt-seedingTimeLimit"`
	Qbttags             []string       `bencode:"qBt-tags"`
	QbttempPathDisabled int64          `bencode:"qBt-tempPathDisabled"`
	SavePath            string         `bencode:"save_path"`
	SeedMode            int64          `bencode:"seed_mode"`
	SeedingTime         int64          `bencode:"seeding_time"`
	SequentialDownload  int64          `bencode:"sequential_download"`
	SuperSeeding        int64          `bencode:"super_seeding"`
	TotalDownloaded     int64          `bencode:"total_downloaded"`
	TotalUploaded       int64          `bencode:"total_uploaded"`
	Trackers            [][]string     `bencode:"trackers"`
	UploadRateLimit     int64          `bencode:"upload_rate_limit"`
	Unfinished          *[]interface{} `bencode:"unfinished,omitempty"`
	hasFiles            bool
	torrentFilePath     string
	torrentFile         map[string]interface{}
	path                string
	fileSizes           int64
	sizeAndPrio         [][]int64
	torrentFileList     []string
	nPieces             int64
	pieceLenght         int64
	replace             []Replace
}

type Replace struct {
	from, to string
}
