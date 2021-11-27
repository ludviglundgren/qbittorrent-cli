package qbittorrent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
)

func (c *Client) Login() error {
	credentials := make(map[string]string)
	credentials["username"] = c.settings.Username
	credentials["password"] = c.settings.Password

	resp, err := c.post("auth/login", credentials)
	if err != nil {
		log.Fatalf("login error: %v", err)
	} else if resp.StatusCode != http.StatusOK { // check for correct status code
		log.Fatalf("login error bad status %v", err)
	}

	// place cookies in jar for future requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		cookieURL, _ := url.Parse("http://localhost:8080")
		c.http.Jar.SetCookies(cookieURL, cookies)
	}

	return nil
}

func (c *Client) GetTorrents() ([]Torrent, error) {
	var torrents []Torrent

	resp, err := c.get("torrents/info", nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}

	return torrents, nil
}

func (c *Client) GetTorrentsFilter(filter TorrentFilter) ([]Torrent, error) {
	var torrents []Torrent

	v := url.Values{}
	v.Add("filter", string(filter))
	params := v.Encode()

	resp, err := c.get("torrents/info?"+params, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}

	return torrents, nil
}

func (c *Client) GetTorrentsRaw() (string, error) {
	resp, err := c.get("torrents/info", nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	return string(data), nil
}

func (c *Client) GetTorrentTrackers(hash string) ([]TorrentTracker, error) {
	var trackers []TorrentTracker

	params := url.Values{}
	params.Add("hash", hash)

	p := params.Encode()

	resp, err := c.get("torrents/trackers?"+p, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &trackers)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v raw: %v", err, body)
	}

	return trackers, nil
}

// AddTorrentFromFile add new torrent from torrent file
func (c *Client) AddTorrentFromFile(file string, options map[string]string) (hash string, err error) {
	// Get meta info from file to find out the hash for later use
	t, err := metainfo.LoadFromFile(file)
	if err != nil {
		log.Fatalf("could not open file %v", err)
	}

	res, err := c.postFile("torrents/add", file, options)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()

	return t.HashInfoBytes().HexString(), nil
}

func (c *Client) AddTorrentFromMagnet(u string, options map[string]string) (hash string, err error) {
	m, err := metainfo.ParseMagnetURI(u)
	if err != nil {
		log.Fatalf("could not parse magnet URI %v", err)
	}

	options["urls"] = u
	res, err := c.post("torrents/add", options)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()

	return m.InfoHash.HexString(), nil
}

func (c *Client) DeleteTorrents(hashes []string, deleteFiles bool) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)
	v.Add("deleteFiles", strconv.FormatBool(deleteFiles))

	encodedHashes := v.Encode()

	resp, err := c.get("torrents/delete?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error deleting torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) ReAnnounceTorrents(hashes []string) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)

	encodedHashes := v.Encode()

	resp, err := c.get("torrents/reannounce?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error reannouncing torrent: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) Pause(hashes []string) error {
	v := url.Values{}
	encodedHashes := "all"

	if len(hashes) > 0 {
		// Add hashes together with | separator
		encodedHashes = strings.Join(hashes, "|")
	}

	v.Add("hashes", encodedHashes)
	encodedHashes = v.Encode()

	resp, err := c.get("torrents/pause?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error pausing torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) Resume(hashes []string) error {
	v := url.Values{}
	encodedHashes := "all"

	if len(hashes) > 0 {
		// Add hashes together with | separator
		encodedHashes = strings.Join(hashes, "|")
	}

	v.Add("hashes", encodedHashes)
	encodedHashes = v.Encode()

	resp, err := c.get("torrents/resume?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error resuming torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}
