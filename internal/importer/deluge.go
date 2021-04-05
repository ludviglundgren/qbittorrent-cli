package importer

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"

	"github.com/anacrolix/torrent/metainfo"
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
	sourceDir := opts.SourceDir

	info, err := os.Stat(sourceDir)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "Directory does not exist: %v\n", sourceDir)
	}
	if err != nil {
		return errors.Wrapf(err, "Directory error: %v\n", sourceDir)

	}
	if !info.IsDir() {
		return errors.Errorf("Directory is a file, not a directory: %#v\n", sourceDir)
	}

	resumeFilePath := sourceDir + "torrents.fastresume"
	if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
		log.Println("Can't find deluge fastresume file")
		return err
	}

	fastresumeFile, err := decodeFastresumeFile(resumeFilePath)
	if err != nil {
		log.Println("Can't decode deluge fastresume file")
		return err
	}

	matches, _ := filepath.Glob(sourceDir + "*.torrent")

	totalJobs := len(matches)
	log.Printf("Total torrents to process: %v \n", totalJobs)

	positionNum := 0
	for torrentID, value := range fastresumeFile {

		// If a file exist in fastresume data but no .torrent file, skip
		if _, err = os.Stat(sourceDir + torrentID + ".torrent"); os.IsNotExist(err) {
			continue
		}

		positionNum++

		// If file already exists, skip
		if _, err = os.Stat(opts.QbitDir + "/" + torrentID + ".torrent"); err == nil {
			log.Printf("%v/%v %v Torrent already exists, skipping", positionNum, totalJobs, torrentID)
			continue
		}

		var fastResume qbittorrent.Fastresume

		if err := bencode.DecodeString(value.(string), &fastResume); err != nil {
			log.Printf("Can't decode row %v. Continue", torrentID)
			continue
		}

		fastResume.TorrentFilePath = sourceDir + torrentID + ".torrent"
		if _, err = os.Stat(fastResume.TorrentFilePath); os.IsNotExist(err) {
			log.Printf("Can't find torrent file %v for %v", fastResume.TorrentFilePath, torrentID)
			return err
		}

		file, err := metainfo.LoadFromFile(sourceDir + torrentID + ".torrent")
		if err != nil {
			return err
		}
		metaInfo, err := file.UnmarshalInfo()
		if err != nil {
			return err
		}

		if metaInfo.Files != nil {
			// valid QbtContentLayout = Original, Subfolder, NoSubfolder
			fastResume.QbtContentLayout = "Original"
			// legacy and should be removed sometime with 4.3.X
			fastResume.QbtHasRootFolder = 1
		} else {
			fastResume.QbtContentLayout = "NoSubfolder"
			fastResume.QbtHasRootFolder = 0
		}

		fastResume.QbtRatioLimit = -2000
		fastResume.QbtSeedStatus = 1
		fastResume.QbtSeedingTimeLimit = -2
		fastResume.QbtName = ""
		fastResume.QbtSavePath = fastResume.SavePath
		fastResume.QbtQueuePosition = positionNum

		fastResume.AutoManaged = 0
		fastResume.NumIncomplete = 0
		fastResume.Paused = 0

		fastResume.ConvertFilePriority(len(metaInfo.Files))

		// fill pieces to set as completed
		fastResume.NumPieces = int64(metaInfo.NumPieces())
		fastResume.FillPieces()

		// TODO handle replace paths

		if opts.DryRun != true {
			if err = fastResume.Encode(opts.QbitDir + "/" + torrentID + ".fastresume"); err != nil {
				log.Printf("Can't create qBittorrent fastresume file %v error: %v", opts.QbitDir+torrentID+".fastresume", err)
				return err
			}

			if err = copyFile(fastResume.TorrentFilePath, opts.QbitDir+"/"+torrentID+".torrent"); err != nil {
				log.Printf("Can't create qBittorrent torrent file %v error %v", opts.QbitDir+torrentID+".torrent", err)
				return err
			}
		}

		log.Printf("%v/%v %v Successfully imported: %v", positionNum, totalJobs, torrentID, metaInfo.Name)
	}

	return nil
}

func decodeFastresumeFile(path string) (map[string]interface{}, error) {
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
