package importer

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	fsutil "github.com/ludviglundgren/qbittorrent-cli/internal/fs"

	"github.com/pkg/errors"
)

type QbittorrentImport struct{}

func NewQbittorrentImporter() Importer {
	return &QbittorrentImport{}
}

func (im *QbittorrentImport) Import(opts Options) error {
	sourceDir := opts.SourceDir

	sourceDirInfo, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("source directory does not exist: %s", sourceDir)
		}

		return errors.Wrapf(err, "source directory error: %s\n", sourceDir)
	}

	if !sourceDirInfo.IsDir() {
		return errors.Errorf("source is a file, not a directory: %s\n", sourceDir)
	}

	if err := fsutil.MkDirIfNotExists(opts.QbitDir); err != nil {
		return errors.Wrapf(err, "qbit directory error: %s\n", opts.QbitDir)
	}

	// TODO read from manifest.json to add categories and tags to client

	// TODO pause for execution and wait until qbit is killed

	err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".torrent") || !strings.HasSuffix(d.Name(), ".fastresume") {
			return nil
		}

		fileName := d.Name()

		outfile := filepath.Join(sourceDir, fileName)

		if err = fsutil.CopyFile(path, outfile); err != nil {
			log.Printf("Could not copy file %s error %q\n", outfile, err)
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "walk directory failure: %s\n", sourceDir)
	}

	return nil
}
