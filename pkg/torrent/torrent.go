package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"

	"github.com/zeebo/bencode"
)

// TorrentInfo torrent meta info
type TorrentInfo struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	CreatedBy    string     `bencode:"created by"`
	CreationDate int64      `bencode:"creation date"`
	Info         struct {
		Length      int64             `bencode:"length"`
		Name        string            `bencode:"name"`
		PieceLength int64             `bencode:"piece length"`
		Pieces      string            `bencode:"pieces"`
		Private     bool              `bencode:"private"`
		Files       []TorrentInfoFile `bencode:"files"`
	} `bencode:"info"`
	UrlList []string `bencode:"url-list"`
}

type TorrentInfoFile struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
}

func Decode(path string) (*TorrentInfo, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var torrent TorrentInfo
	if err := bencode.DecodeBytes(dat, &torrent); err != nil {
		return nil, err
	}

	return &torrent, nil
}

func OpenDecodeRaw(path string) (map[string]interface{}, error) {
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

func CopyFile(src string, dst string) error {
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

func CalculateInfoHash(torrent map[string]interface{}) (hash string) {
	t, _ := bencode.EncodeString(torrent["info"])
	h := sha1.New()
	io.WriteString(h, t)
	hash = hex.EncodeToString(h.Sum(nil))
	return hash
}

func GetName(torrent map[string]interface{}) (name string) {
	info, _ := torrent["info"].(map[string]interface{})
	//info, _ := bencode.EncodeString(torrent["info"])
	//h := sha1.New()
	//io.WriteString(h, t)
	//hash = hex.EncodeToString(h.Sum(nil))
	return info["name"].(string)
}
