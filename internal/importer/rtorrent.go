package importer

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
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

type NewTorrentStructure struct {
	ActiveTime        int64  `bencode:"active_time"`
	AddedTime         int64  `bencode:"added_time"`
	Allocation        string `bencode:"allocation"`
	ApplyIpFilter     int64  `bencode:"apply_ip_filter"`
	AutoManaged       int64  `bencode:"auto_managed"`
	CompletedTime     int64  `bencode:"completed_time"`
	DisableDHT        int64  `bencode:"disable_dht"`
	DisableLSD        int64  `bencode:"disable_lsd"`
	DisablePEX        int64  `bencode:"disable_pex"`
	DownloadRateLimit int64  `bencode:"download_rate_limit"`
	FileFormat        string `bencode:"file-format"`
	FileVersion       int64  `bencode:"file-version"`
	//FilePriority        []int                  `bencode:"file_priority"`
	FinishedTime      int64    `bencode:"finished_time"`
	HttpSeeds         []string `bencode:"httpseeds"`
	InfoHash          string   `bencode:"info-hash"`
	LastDownload      int64    `bencode:"last_download"`
	LastSeenComplete  int64    `bencode:"last_seen_complete"`
	LastUpload        int64    `bencode:"last_upload"`
	LibTorrentVersion string   `bencode:"libtorrent-version"`
	MaxConnections    int64    `bencode:"max_connections"`
	MaxUploads        int64    `bencode:"max_uploads"`
	NumComplete       int64    `bencode:"num_complete"`
	NumDownloaded     int64    `bencode:"num_downloaded"`
	NumIncomplete     int64    `bencode:"num_incomplete"`
	//MappedFiles               []string               `bencode:"mapped_files,omitempty"`
	MappedFiles               []string
	Paused                    int64                  `bencode:"paused"`
	Pieces                    []byte                 `bencode:"pieces"`
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
	StopWhenReady             int64                  `bencode:"stop_when_ready"`
	SuperSeeding              int64                  `bencode:"super_seeding"`
	TotalDownloaded           int64                  `bencode:"total_downloaded"`
	TotalUploaded             int64                  `bencode:"total_uploaded"`
	Trackers                  [][]string             `bencode:"trackers"`
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
	//PiecePriority       []byte                 `bencode:"piece_priority"`
	//Replace             []replace.Replace      `bencode:"-"`
	Separator string `bencode:"-"`
}

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

	//if stat, err := os.Stat(torrentsSessionDir); err == nil && stat.IsDir() {
	//	// path is a directory
	//
	//	log.Println("path is dir")
	//}

	matches, _ := filepath.Glob(torrentsSessionDir + "*.torrent")

	totalJobs := len(matches)
	log.Printf("Total torrents to process: %v\n", totalJobs)

	positionNum := 0
	for _, match := range matches {
		positionNum++

		//var decodedVal NewFastResumeFile

		// TODO grab current directory and feed torrent via qbittorrent api to add fast resume data

		torrentID := getTorrentName(match)

		// TODO check if file exists, if true skip
		// If file already exists, skip
		//if _, err = os.Stat(opts.QbitDir + "/" + torrentID + ".torrent"); err == nil {
		//	log.Printf("Torrent already exists, skipping: %v", torrentID)
		//	continue
		//}

		torrentFile, err := decodeTorrentFile(match)
		if err != nil {
			log.Printf("Can't decode torrent file %v. Can't decode string %v. Continue", match, torrentID)
			continue
		}

		file, err := decodeTorrentMetaFile(match)
		if err != nil {
			return err
		}

		// check for FILE.torrent.libtorrent_resume
		resumeFile, err := decodeRTorrentLibTorrentResumeFile(match + ".libtorrent_resume")
		if err != nil {
			return err
		}

		// check for FILE.torrent.rtorrent
		rtorrentFile, err := decodeRTorrentFile(match + ".rtorrent")
		if err != nil {
			return err
		}

		// create new fast resume file
		//decodedVal.AddedTime = strToIntClean(rtorrentFile.Custom.AddTime)
		//decodedVal.SeedingTime = strToIntClean(rtorrentFile.Custom.SeedingTime)
		//decodedVal.CompletedTime = rtorrentFile.CompletedTime

		newFastResume := NewTorrentStructure{
			ActiveTime:                0,
			AddedTime:                 strToIntClean(rtorrentFile.Custom.AddTime),
			Allocation:                "sparse",
			ApplyIpFilter:             1,
			AutoManaged:               0,
			CompletedTime:             rtorrentFile.TimestampFinished,
			DownloadRateLimit:         -1,
			FileFormat:                "libtorrent resume file",
			FileVersion:               1,
			FinishedTime:              0,
			LastDownload:              rtorrentFile.TimestampFinished,
			LastSeenComplete:          0,
			LastUpload:                0,
			LibTorrentVersion:         "1.2.11.0",
			MaxConnections:            -1,
			MaxUploads:                -1,
			NumComplete:               0,
			NumDownloaded:             0,
			NumIncomplete:             0,
			Paused:                    1,
			QbtName:                   "",
			QbtContentLayout:          "Original",
			QbtFirstLastPiecePriority: 0,
			QbtRatioLimit:             -2000,
			QbtSavePath:               rtorrentFile.Directory,
			QbtSeedStatus:             1,
			QbtSeedingTimeLimit:       -2,
			QbtCategory:               "",
			QbtTags:                   []string{},
			SavePath:                  rtorrentFile.Directory,
			SeedMode:                  0,
			SeedingTime:               strToIntClean(rtorrentFile.Custom.SeedingTime),
			SequentialDownload:        0,
			StopWhenReady:             0,
			SuperSeeding:              0,
			TotalDownloaded:           rtorrentFile.TotalDownloaded,
			TotalUploaded:             rtorrentFile.TotalUploaded,
			UploadRateLimit:           -1,
			UrlList:                   file.UrlList,

			TorrentFile: torrentFile,
			Separator:   "/",
			Path:        rtorrentFile.Directory,

			//NumPieces: file.Info.Pieces,
			//Trackers:            file.AnnounceList,
		}
		newFastResume.TorrentFile = torrentFile

		if file.Info.Files != nil {
			newFastResume.HasFiles = true
		} else {
			newFastResume.HasFiles = false
		}

		//if value["path"].(string)[len(value["path"].(string))-1] == os.PathSeparator {
		//	newFastResume.Path = value["path"].(string)[:len(value["path"].(string))-1]
		//} else {
		//	newFastResume.Path = value["path"].(string)
		//}

		//newFastResume.ConvertPriority(newFastResume.TorrentFile["prio"].(string))

		// handle trackers
		newFastResume.Trackers = convertTrackers(*resumeFile)

		// todo do this as part of fill missing or other convert util
		newFastResume.NumPieces = int64(len(file.Info.Pieces)) / 20

		newFastResume.FillMissing()

		// handle rename
		newBaseName := newFastResume.GetHash()

		newFastResume.InfoHash = newFastResume.GetInfoHash()

		// copy torrent file
		if err = encodeNewFastResumeFile(opts.QbitDir+"/"+newBaseName+".fastresume", &newFastResume); err != nil {
			log.Printf("Can't create qBittorrent fastresume file %v error: %v", opts.QbitDir+torrentID+".fastresume", err)
			return err
		}

		if err = copyFile(match, opts.QbitDir+"/"+newBaseName+".torrent"); err != nil {
			log.Printf("Can't create qBittorrent torrent file %v error %v", opts.QbitDir+torrentID+".torrent", err)
			return err
		}

		//go processFiles(torrentID, decodedVal, opts, &torrentsSessionDir, positionNum, totalJobs)
		log.Printf("%v/%v Sucessfully imported: %v", positionNum, totalJobs, file.Info.Name)

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

func (newstructure *NewTorrentStructure) GetHash() (hash string) {
	torinfo, _ := bencode.EncodeString(newstructure.TorrentFile["info"].(map[string]interface{}))
	h := sha1.New()
	io.WriteString(h, torinfo)
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func (newstructure *NewTorrentStructure) GetInfoHash() (hash string) {
	torinfo, _ := bencode.EncodeString(newstructure.TorrentFile["info"].(map[string]interface{}))
	h := sha1.New()
	io.WriteString(h, torinfo)

	hash = fmt.Sprintf("<hex>% X</hex>", h.Sum(nil))
	return
}

func (newstructure *NewTorrentStructure) IfCompletedOn() {
	if newstructure.CompletedTime != 0 {
		newstructure.LastSeenComplete = time.Now().Unix()
	} else {
		newstructure.Unfinished = new([]interface{})
	}
}

func (newstructure *NewTorrentStructure) FillMissing() {
	newstructure.IfCompletedOn()

	newstructure.FillSizes()
	newstructure.FillSavePaths()

	newstructure.Pieces = newstructure.FillWholePieces("1")
	//if newstructure.Unfinished != nil {
	//	newstructure.Pieces = newstructure.FillWholePieces("0")
	//	if newstructure.HasFiles {
	//		newstructure.PiecePriority = newstructure.FillPiecesParted()
	//	} else {
	//		newstructure.PiecePriority = newstructure.FillWholePieces("1")
	//	}
	//} else {
	//	if newstructure.HasFiles {
	//		newstructure.Pieces = newstructure.FillPiecesParted()
	//	} else {
	//		newstructure.Pieces = newstructure.FillWholePieces("1")
	//	}
	//	newstructure.PiecePriority = newstructure.Pieces
	//}
}

func (newstructure *NewTorrentStructure) FillWholePieces(chr string) []byte {
	var newpieces = make([]byte, 0, newstructure.NumPieces)
	nchr, _ := strconv.Atoi(chr)
	for i := int64(0); i < newstructure.NumPieces; i++ {
		newpieces = append(newpieces, byte(nchr))
	}
	return newpieces
}

func (newstructure *NewTorrentStructure) FillPiecesParted() []byte {
	var newPieces = make([]byte, 0, newstructure.NumPieces)
	var allocation [][]int64
	chrOne, _ := strconv.Atoi("1")
	chrZero, _ := strconv.Atoi("0")
	offset := int64(0)
	for _, pair := range newstructure.sizeAndPrio {
		allocation = append(allocation, []int64{offset + 1, offset + pair[0], pair[1]})
		offset = offset + pair[0]
	}
	for i := int64(0); i < newstructure.NumPieces; i++ {
		belongs := false
		first, last := i*newstructure.PieceLength, (i+1)*newstructure.PieceLength
		for _, trio := range allocation {
			if (first >= trio[0]-newstructure.PieceLength && last <= trio[1]+newstructure.PieceLength) && trio[2] == 1 {
				belongs = true
			}
		}
		if belongs {
			newPieces = append(newPieces, byte(chrOne))
		} else {
			newPieces = append(newPieces, byte(chrZero))
		}
	}
	return newPieces
}

func (newstructure *NewTorrentStructure) FillSavePaths() {
	var torrentname string
	if name, ok := newstructure.TorrentFile["info"].(map[string]interface{})["name.utf-8"].(string); ok {
		torrentname = name
	} else {
		torrentname = newstructure.TorrentFile["info"].(map[string]interface{})["name"].(string)
	}
	origpath := newstructure.Path
	dirpaths := strings.Split(origpath, "\\")
	lastdirname := dirpaths[len(dirpaths)-1]
	if newstructure.HasFiles {
		if lastdirname == torrentname {
			// TODO handle new enums Original, Subfolder, NoSubfolder
			newstructure.QbtHasRootFolder = 1
			//newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbtHasRootFolder = 0
			//newstructure.QbtSavePath = newstructure.Path + newstructure.Separator
			newstructure.MappedFiles = newstructure.torrentFileList
		}
	} else {
		if lastdirname == torrentname {
			newstructure.QbtHasRootFolder = 0
			//newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		} else {
			newstructure.QbtHasRootFolder = 0
			newstructure.torrentFileList = append(newstructure.torrentFileList, lastdirname)
			newstructure.MappedFiles = newstructure.torrentFileList
			//newstructure.QbtSavePath = origpath[0 : len(origpath)-len(lastdirname)]
		}
	}
	//for _, pattern := range newstructure.Replace {
	//	newstructure.QbtSavePath = strings.ReplaceAll(newstructure.QbtSavePath, pattern.From, pattern.To)
	//}
	//var oldsep string
	//switch newstructure.Separator {
	//case "\\":
	//	oldsep = "/"
	//case "/":
	//	oldsep = "\\"
	//}
	//newstructure.QbtSavePath = strings.ReplaceAll(newstructure.QbtSavePath, oldsep, newstructure.Separator)
	//newstructure.SavePath = strings.ReplaceAll(newstructure.QbtSavePath, "\\", "/")
	//
	//for num, entry := range newstructure.MappedFiles {
	//	newentry := strings.ReplaceAll(entry, oldsep, newstructure.Separator)
	//	if entry != newentry {
	//		newstructure.MappedFiles[num] = newentry
	//	}
	//}
}

func (newstructure *NewTorrentStructure) FillSizes() {
	newstructure.fileSizes = 0
	if newstructure.HasFiles {
		var filelists [][]int64
		for _, file := range newstructure.TorrentFile["info"].(map[string]interface{})["files"].([]interface{}) {
			var length, mtime int64
			var filestrings []string
			if path, ok := file.(map[string]interface{})["path.utf-8"].([]interface{}); ok {
				for _, f := range path {
					filestrings = append(filestrings, f.(string))
				}
			} else {
				for _, f := range file.(map[string]interface{})["path"].([]interface{}) {
					filestrings = append(filestrings, f.(string))
				}
			}
			filename := strings.Join(filestrings, newstructure.Separator)
			newstructure.torrentFileList = append(newstructure.torrentFileList, filename)
			//fullpath := newstructure.Path + newstructure.Separator + filename
			newstructure.fileSizes += file.(map[string]interface{})["length"].(int64)

			length, mtime = 0, 0
			newstructure.sizeAndPrio = append(newstructure.sizeAndPrio,
				[]int64{file.(map[string]interface{})["length"].(int64), 0})

			//if n := newstructure.FilePriority[num]; n != 0 {
			//	length = file.(map[string]interface{})["length"].(int64)
			//	newstructure.sizeAndPrio = append(newstructure.sizeAndPrio, []int64{length, 1})
			//	mtime = fmTime(fullpath)
			//} else {
			//	length, mtime = 0, 0
			//	newstructure.sizeAndPrio = append(newstructure.sizeAndPrio,
			//		[]int64{file.(map[string]interface{})["length"].(int64), 0})
			//}
			flenmtime := []int64{length, mtime}
			filelists = append(filelists, flenmtime)
		}
	}
}

//func (newstructure *NewTorrentStructure) ConvertPriority(src string) {
//	var newPrio []int
//	for _, c := range []byte(src) {
//		if i := int(c); (i == 0) || (i == 128) { // if not selected
//			newPrio = append(newPrio, 0)
//		} else if (i >= 1) && (i <= 8) { // if low or normal prio
//			newPrio = append(newPrio, 1)
//		} else if (i > 8) && (i <= 15) { // if high prio
//			newPrio = append(newPrio, 6)
//		} else {
//			newPrio = append(newPrio, 0)
//		}
//	}
//	newstructure.FilePriority = newPrio
//}

func (newstructure *NewTorrentStructure) GetTrackers(trackers interface{}) {
	switch strct := trackers.(type) {
	case []interface{}:
		for _, st := range strct {
			newstructure.GetTrackers(st)
		}
	case string:
		for _, str := range strings.Fields(strct) {
			newstructure.Trackers = append(newstructure.Trackers, []string{str})
		}

	}
}

func convertTrackers(trackers RTorrentLibTorrentResumeFile) [][]string {
	var ret [][]string

	for t, _ := range trackers.Trackers {
		if t == "dht://" {
			continue
		}
		ret = append(ret, []string{t})
	}

	return ret
}

func fmTime(path string) (mtime int64) {
	if fi, err := os.Stat(path); err != nil {
		return 0
	} else {
		mtime = fi.ModTime().Unix()
		return
	}
}

func getTorrentName(file string) string {

	_, fileName := filepath.Split(file)
	//fmt.Printf("filename: %v\n", fileName)

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

func encodeNewFastResumeFile(path string, newStructure *NewTorrentStructure) error {

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
		Private     bool   `bencode:"private"`
		Files       []struct {
			Length int64    `bencode:"length"`
			Path   []string `bencode:"path"`
		} `bencode:"files"`
	} `bencode:"info"`
	UrlList []string `bencode:"url-list"`
}

type RTorrentLibTorrentResumeFile struct {
	Trackers map[string]map[string]int `bencode:"trackers"`
}

type RTorrentTorrentFile struct {
	Custom struct {
		AddTime     string `bencode:"addtime"`
		SeedingTime string `bencode:"seedingtime"`
		//XFileName   string `bencode:"x-filename"`
	} `bencode:"custom"`
	Directory         string `bencode:"directory"`
	SeedingTime       int64  `bencode:"seeding_time"`
	TotalDownloaded   int64  `bencode:"total_downloaded"`
	TotalUploaded     int64  `bencode:"total_uploaded"`
	TimestampFinished int64  `bencode:"timestamp.finished"`
	TimestampStarted  int64  `bencode:"timestamp.started"`
}
