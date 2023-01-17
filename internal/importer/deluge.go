package importer

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/ludviglundgren/qbittorrent-cli/pkg/qbittorrent"
	"github.com/ludviglundgren/qbittorrent-cli/pkg/torrent"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"
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
		return errors.Wrapf(err, "Directory does not exist: %s\n", sourceDir)
	}
	if err != nil {
		return errors.Wrapf(err, "Directory error: %s\n", sourceDir)

	}
	if !info.IsDir() {
		return errors.Errorf("Directory is a file, not a directory: %s\n", sourceDir)
	}

	resumeFilePath := path.Join(sourceDir, "torrents.fastresume")
	if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
		log.Println("Could not find deluge fastresume file")
		return err
	}

	fastresumeFile, err := decodeFastresumeFile(resumeFilePath)
	if err != nil {
		log.Println("Could not decode deluge fastresume file")
		return err
	}

	matches, _ := filepath.Glob(path.Join(sourceDir, "*.torrent"))

	totalJobs := len(matches)
	log.Printf("Total torrents to process: %d\n", totalJobs)

	positionNum := 0
	for torrentID, value := range fastresumeFile {
		torrentNamePath := path.Join(sourceDir, torrentID+".torrent")

		// If a file exist in fastresume data but no .torrent file, skip
		if _, err = os.Stat(torrentNamePath); os.IsNotExist(err) {
			continue
		}

		positionNum++

		torrentOutFile := path.Join(opts.QbitDir, torrentID+".torrent")

		// If file already exists, skip
		if _, err = os.Stat(torrentOutFile); err == nil {
			log.Printf("%d/%d %s Torrent already exists, skipping", positionNum, totalJobs, torrentID)
			continue
		}

		var fastResume qbittorrent.Fastresume

		if err := bencode.DecodeString(value.(string), &fastResume); err != nil {
			log.Printf("Could not decode row %s. Continue", torrentID)
			continue
		}

		fastResume.TorrentFilePath = torrentNamePath
		if _, err = os.Stat(fastResume.TorrentFilePath); os.IsNotExist(err) {
			log.Printf("Could not find torrent file %s for %s", fastResume.TorrentFilePath, torrentID)
			return err
		}

		file, err := metainfo.LoadFromFile(torrentNamePath)
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
			fastResumeOutFile := path.Join(opts.QbitDir, torrentID+".fastresume")
			if err = fastResume.Encode(fastResumeOutFile); err != nil {
				log.Printf("Could not create qBittorrent fastresume file %s error: %q", fastResumeOutFile, err)
				return err
			}

			if err = torrent.CopyFile(fastResume.TorrentFilePath, torrentOutFile); err != nil {
				log.Printf("Could not copy qBittorrent torrent file %s error %q", torrentOutFile, err)
				return err
			}
		}

		log.Printf("%d/%d %s Successfully imported: %s", positionNum, totalJobs, torrentID, metaInfo.Name)
	}

	return nil
}

func decodeFastresumeFile(path string) (map[string]interface{}, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var torrent map[string]interface{}
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}

	return torrent, nil
}
