package qbittorrent

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (c *Client) Login(ctx context.Context) error {
	credentials := make(map[string]string)
	credentials["username"] = c.username
	credentials["password"] = c.password

	resp, err := c.postCtx(ctx, "auth/login", credentials)
	if err != nil {
		log.Fatalf("login error: %v", err)
	} else if resp.StatusCode != http.StatusOK { // check for correct status code
		log.Fatalf("login error bad status %v", err)
	}

	// place cookies in jar for future requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		reqUrl, err := c.buildUrl("")
		if err != nil {
			log.Fatalf("login error, could not parse url for cookie: %q", err)
		}
		cookieURL, _ := url.Parse(reqUrl)
		c.http.Jar.SetCookies(cookieURL, cookies)
	}

	return nil
}

func (c *Client) GetTorrentsWithFilters(ctx context.Context, req *GetTorrentsRequest) ([]Torrent, error) {
	v := url.Values{}

	if req.Filter != "" {
		v.Add("filter", req.Filter)
	} else if req.Category != "" {
		v.Add("category", req.Category)
	} else if req.Tag != "" {
		v.Add("tag", req.Tag)
	} else if req.Hashes != "" {
		v.Add("hashes", req.Hashes)
	}

	resp, err := c.getCtx(ctx, "torrents/info", v)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch torrents")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	var torrents []Torrent
	if err = json.Unmarshal(body, &torrents); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal torrents list")
	}

	return torrents, nil
}

func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	var torrents []Torrent

	resp, err := c.getCtx(ctx, "torrents/info", nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}

	return torrents, nil
}

func (c *Client) GetTorrentsFilter(ctx context.Context, filter TorrentFilter) ([]Torrent, error) {
	var torrents []Torrent

	v := url.Values{}
	v.Add("filter", string(filter))
	params := v.Encode()

	resp, err := c.getCtx(ctx, "torrents/info?"+params, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}

	return torrents, nil
}

func (c *Client) GetTorrentsByCategory(ctx context.Context, category string) ([]Torrent, error) {
	var torrents []Torrent

	v := url.Values{}
	//v.Add("filter", string(TorrentFilterSeeding))
	v.Add("category", category)
	params := v.Encode()

	resp, err := c.getCtx(ctx, "torrents/info?"+params, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	err = json.Unmarshal(body, &torrents)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}

	return torrents, nil
}

func (c *Client) GetTorrentsRaw(ctx context.Context) (string, error) {
	resp, err := c.getCtx(ctx, "torrents/info", nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	return string(data), nil
}

func (c *Client) GetTorrentByHash(ctx context.Context, hash string) (string, error) {
	v := url.Values{}
	v.Add("hashes", hash)
	params := v.Encode()

	resp, err := c.getCtx(ctx, "torrents/info?"+params, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	return string(data), nil
}

// GetTorrentsByPrefixes Search for torrents using provided prefixes; checks against either hashes, names, or both
func (c *Client) GetTorrentsByPrefixes(ctx context.Context, terms []string, hashes bool, names bool) ([]Torrent, error) {
	torrents, err := c.GetTorrents(ctx)
	if err != nil {
		log.Fatalf("ERROR: could not retrieve torrents: %v\n", err)
	}

	matchedTorrents := map[Torrent]bool{}
	for _, torrent := range torrents {
		if hashes {
			for _, targetHash := range terms {
				if strings.HasPrefix(torrent.Hash, targetHash) {
					matchedTorrents[torrent] = true
					break
				}
			}

			if matchedTorrents[torrent] {
				continue
			}
		}

		if names {
			for _, targetName := range terms {
				if strings.HasPrefix(torrent.Name, targetName) {
					matchedTorrents[torrent] = true
					break
				}
			}
		}
	}

	var foundTorrents []Torrent
	for torrent := range matchedTorrents {
		foundTorrents = append(foundTorrents, torrent)
	}

	return foundTorrents, nil
}

func (c *Client) GetTorrentTrackers(ctx context.Context, hash string) ([]TorrentTracker, error) {
	var trackers []TorrentTracker

	params := url.Values{}
	params.Add("hash", hash)

	p := params.Encode()

	resp, err := c.getCtx(ctx, "torrents/trackers?"+p, nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
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
func (c *Client) AddTorrentFromFile(ctx context.Context, file string, options map[string]string) (hash string, err error) {
	// Get meta info from file to find out the hash for later use
	t, err := metainfo.LoadFromFile(file)
	if err != nil {
		log.Fatalf("could not open file %v", err)
	}

	res, err := c.postFileCtx(ctx, "torrents/add", file, options)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()

	return t.HashInfoBytes().HexString(), nil
}

func (c *Client) AddTorrentFromMagnet(ctx context.Context, u string, options map[string]string) (hash string, err error) {
	m, err := metainfo.ParseMagnetUri(u)
	if err != nil {
		log.Fatalf("could not parse magnet URI %v", err)
	}

	options["urls"] = u
	res, err := c.postCtx(ctx, "torrents/add", options)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()

	return m.InfoHash.HexString(), nil
}

func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)
	v.Add("deleteFiles", strconv.FormatBool(deleteFiles))

	encodedHashes := v.Encode()

	resp, err := c.postCtx(ctx, "torrents/delete?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error deleting torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) ReAnnounceTorrents(ctx context.Context, hashes []string) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)

	encodedHashes := v.Encode()

	resp, err := c.postCtx(ctx, "torrents/reannounce?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error reannouncing torrent: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) Pause(ctx context.Context, hashes []string) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)

	encodedHashes := v.Encode()

	resp, err := c.postCtx(ctx, "torrents/pause?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error pausing torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) Resume(ctx context.Context, hashes []string) error {
	v := url.Values{}

	// Add hashes together with | separator
	hv := strings.Join(hashes, "|")
	v.Add("hashes", hv)

	encodedHashes := v.Encode()

	resp, err := c.postCtx(ctx, "torrents/resume?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error resuming torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) SetCategory(ctx context.Context, hashes []string, category string) error {
	v := url.Values{}
	encodedHashes := ""

	if len(hashes) > 0 {
		// Add hashes together with | separator
		encodedHashes = strings.Join(hashes, "|")
	}

	// TODO batch action if more than 25

	v.Add("hashes", encodedHashes)
	v.Add("category", category)
	encodedHashes = v.Encode()

	resp, err := c.postCtx(ctx, "torrents/setCategory?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error resuming torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) SetTag(ctx context.Context, hashes []string, tag string) error {
	v := url.Values{}
	encodedHashes := ""

	if len(hashes) > 0 {
		// Add hashes together with | separator
		encodedHashes = strings.Join(hashes, "|")
	}

	// TODO batch action if more than 25

	v.Add("hashes", encodedHashes)
	v.Add("tags", tag)
	encodedHashes = v.Encode()

	resp, err := c.postCtx(ctx, "torrents/addTags?"+encodedHashes, nil)
	if err != nil {
		log.Fatalf("error resuming torrents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	return nil
}
