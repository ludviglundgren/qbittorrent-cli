package qbittorrent

import (
	"bufio"
	"crypto/sha1"
	"log"
	"os"
	"strings"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"

	"github.com/zeebo/bencode"
)

// Fastresume represents a qBittorrent fastresume file
type Fastresume struct {
	ActiveTime                int64                  `bencode:"active_time"`
	AddedTime                 int64                  `bencode:"added_time"`
	Allocation                string                 `bencode:"allocation"`
	ApplyIpFilter             int64                  `bencode:"apply_ip_filter"`
	AutoManaged               int64                  `bencode:"auto_managed"`
	CompletedTime             int64                  `bencode:"completed_time"`
	DisableDHT                int64                  `bencode:"disable_dht"`
	DisableLSD                int64                  `bencode:"disable_lsd"`
	DisablePEX                int64                  `bencode:"disable_pex"`
	DownloadRateLimit         int64                  `bencode:"download_rate_limit"`
	FileFormat                string                 `bencode:"file-format"`
	FileVersion               int64                  `bencode:"file-version"`
	FilePriority              []int                  `bencode:"file_priority"`
	FinishedTime              int64                  `bencode:"finished_time"`
	HttpSeeds                 []string               `bencode:"httpseeds"`
	InfoHash                  []byte                 `bencode:"info-hash"`
	LastDownload              int64                  `bencode:"last_download"`
	LastSeenComplete          int64                  `bencode:"last_seen_complete"`
	LastUpload                int64                  `bencode:"last_upload"`
	LibTorrentVersion         string                 `bencode:"libtorrent-version"`
	MaxConnections            int64                  `bencode:"max_connections"`
	MaxUploads                int64                  `bencode:"max_uploads"`
	NumComplete               int64                  `bencode:"num_complete"`
	NumDownloaded             int64                  `bencode:"num_downloaded"`
	NumIncomplete             int64                  `bencode:"num_incomplete"`
	Paused                    int64                  `bencode:"paused"`
	Pieces                    string                 `bencode:"pieces"`
	PiecePriority             []byte                 `bencode:"piece_priority"`
	Peers                     string                 `bencode:"peers"`
	Peers6                    string                 `bencode:"peers6"`
	QbtCategory               string                 `bencode:"qBt-category"`
	QbtContentLayout          string                 `bencode:"qBt-contentLayout"`
	QbtHasRootFolder          int64                  `bencode:"qBt-hasRootFolder"`
	QbtFirstLastPiecePriority int64                  `bencode:"qBt-firstLastPiecePriority"`
	QbtName                   string                 `bencode:"qBt-name"`
	QbtRatioLimit             int64                  `bencode:"qBt-ratioLimit"`
	QbtSavePath               string                 `bencode:"qBt-savePath"`
	QbtSeedStatus             int64                  `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit       int64                  `bencode:"qBt-seedingTimeLimit"`
	QbtTags                   []string               `bencode:"qBt-tags"`
	SavePath                  string                 `bencode:"save_path"`
	SeedMode                  int64                  `bencode:"seed_mode"`
	SeedingTime               int64                  `bencode:"seeding_time"`
	SequentialDownload        int64                  `bencode:"sequential_download"`
	ShareMode                 int64                  `bencode:"share_mode"`
	StopWhenReady             int64                  `bencode:"stop_when_ready"`
	SuperSeeding              int64                  `bencode:"super_seeding"`
	TotalDownloaded           int64                  `bencode:"total_downloaded"`
	TotalUploaded             int64                  `bencode:"total_uploaded"`
	Trackers                  [][]string             `bencode:"trackers"`
	UploadMode                int64                  `bencode:"upload_mode"`
	UploadRateLimit           int64                  `bencode:"upload_rate_limit"`
	UrlList                   []string               `bencode:"url-list"`
	Unfinished                *[]interface{}         `bencode:"unfinished,omitempty"`
	WithoutLabels             bool                   `bencode:"-"`
	WithoutTags               bool                   `bencode:"-"`
	HasFiles                  bool                   `bencode:"-"`
	TorrentFilePath           string                 `bencode:"-"`
	TorrentFile               map[string]interface{} `bencode:"-"`
	Path                      string                 `bencode:"-"`
	fileSizes                 int64                  `bencode:"-"`
	sizeAndPrio               [][]int64              `bencode:"-"`
	torrentFileList           []string               `bencode:"-"`
	NumPieces                 int64                  `bencode:"-"`
	PieceLength               int64                  `bencode:"-"`
	MappedFiles               []string               `bencode:"mapped_files,omitempty"`
}

// Encode qBittorrent fastresume file
func (fr *Fastresume) Encode(path string) error {
	_, err := os.Create(path)
	if err != nil {
		log.Printf("os create error: %v", err)
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("os open file error: %v", err)
		return err
	}

	defer file.Close()

	bufferedWriter := bufio.NewWriter(file)
	enc := bencode.NewEncoder(bufferedWriter)
	if err := enc.Encode(fr); err != nil {
		log.Printf("encode error: %v", err)
		return err
	}

	bufferedWriter.Flush()
	return nil
}

// ConvertFilePriority for each file set priority
func (fr *Fastresume) ConvertFilePriority(files []torrent.TorrentInfoFile) {
	var newPrioList []int

	/*
		File priority:
		0 Do not download
		1 Normal
		2 High
	*/
	for i := 0; i < len(files); i++ {
		newPrioList = append(newPrioList, 1)
	}

	fr.FilePriority = newPrioList
}

func (fr *Fastresume) FillPieces() {
	var pieces = make([]string, 0, fr.NumPieces)
	for i := int64(0); i < fr.NumPieces; i++ {
		pieces = append(pieces, "\x01")
	}
	fr.Pieces = strings.Join(pieces, "")
}

func (fr *Fastresume) GetInfoHashSHA1() (hash []byte) {
	torInfo, _ := bencode.EncodeString(fr.TorrentFile["info"].(map[string]interface{}))
	h := sha1.New()
	_, _ = h.Write([]byte(torInfo))

	ab := h.Sum(nil)
	return ab
}

//func (newstructure *NewTorrentStructure) GetTrackers(trackers interface{}) {
//	switch strct := trackers.(type) {
//	case []interface{}:
//		for _, st := range strct {
//			newstructure.GetTrackers(st)
//		}
//	case string:
//		for _, str := range strings.Fields(strct) {
//			newstructure.Trackers = append(newstructure.Trackers, []string{str})
//		}
//
//	}
//}
