package qbittorrent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (c *Client) Login(ctx context.Context) error {
	v := url.Values{}
	v.Add("username", c.username)
	v.Add("password", c.password)

	resp, err := c.postCtx(ctx, "auth/login", v)
	if err != nil {
		return errors.Wrap(err, "login error")
	}

	if resp.StatusCode != http.StatusOK { // check for correct status code
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("login: unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	// place cookies in jar for future requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		reqUrl, err := c.buildUrl("")
		if err != nil {
			return errors.Wrap(err, "login error, could not parse url for cookie")
		}
		cookieURL, _ := url.Parse(reqUrl)
		c.http.Jar.SetCookies(cookieURL, cookies)
	}

	return nil
}

func (c *Client) GetTorrentsWithFilters(ctx context.Context, req *GetTorrentsRequest) ([]Torrent, error) {
	v := url.Values{}

	if req.Filter != "" {
		v.Add("filter", string(req.Filter))
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
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal torrents list")
	}

	return torrents, nil
}

func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	resp, err := c.getCtx(ctx, "torrents/info", nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch torrents")
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, errors.Wrap(err, "could not read body")
	}

	var torrents []Torrent
	if err := json.Unmarshal(body, &torrents); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal torrents list")
	}

	return torrents, nil
}

func (c *Client) GetTorrentsRaw(ctx context.Context) (string, error) {
	resp, err := c.getCtx(ctx, "torrents/info", nil)
	if err != nil {
		return "", errors.Wrap(err, "could not fetch torrents")
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read body")
	}

	return string(data), nil
}

// GetTorrentsByPrefixes Search for torrents using provided prefixes; checks against either hashes, names, or both
func (c *Client) GetTorrentsByPrefixes(ctx context.Context, terms []string, hashes bool, names bool) ([]Torrent, error) {
	torrents, err := c.GetTorrents(ctx)
	if err != nil {
		return nil, err
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
	params := url.Values{}
	params.Add("hash", hash)

	resp, err := c.getCtx(ctx, "torrents/trackers?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching torrent trackers for: %s", hash)
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "could not read body")
	}

	var trackers []TorrentTracker
	if err := json.Unmarshal(body, &trackers); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal raw json: %s", body)
	}

	return trackers, nil
}

// AddTorrentFromFile add new torrent from torrent file
func (c *Client) AddTorrentFromFile(ctx context.Context, file string, options map[string]string) (hash string, err error) {
	// Get meta info from file to find out the hash for later use
	t, err := metainfo.LoadFromFile(file)
	if err != nil {
		return "", errors.Wrapf(err, "could not open file: %s", file)
	}

	resp, err := c.postFileCtx(ctx, "torrents/add", file, options)
	if err != nil {
		return "", errors.Wrapf(err, "could not add torrent from file: %s", file)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "could not read body")
		}

		return "", errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return t.HashInfoBytes().HexString(), nil
}

func (c *Client) AddTorrentFromMagnet(ctx context.Context, magnetUri string, options map[string]string) (hash string, err error) {
	magnet, err := metainfo.ParseMagnetUri(magnetUri)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse magnet URI: %s", magnetUri)
	}

	form := url.Values{}
	form.Add("urls", magnetUri)

	for key, val := range options {
		form.Add(key, val)
	}

	resp, err := c.postFormCtx(ctx, "torrents/add", form)
	if err != nil {
		return "", errors.Wrapf(err, "could not add torrent from magnet: %s", magnetUri)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "could not read body")
		}

		return "", errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return magnet.InfoHash.HexString(), nil
}

func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	form := url.Values{}
	// Add hashes together with | separator
	form.Add("hashes", strings.Join(hashes, "|"))

	// Only include the deleteFiles parameter if it's set to true
	if deleteFiles {
		form.Add("deleteFiles", "true")
	}

	resp, err := c.postFormCtx(ctx, "torrents/delete", form)
	if err != nil {
		return errors.Wrapf(err, "could not delete torrents by hashes: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}

func (c *Client) ReAnnounceTorrents(ctx context.Context, hashes []string) error {
	v := url.Values{}

	// Add hashes together with | separator
	v.Add("hashes", strings.Join(hashes, "|"))

	resp, err := c.postCtx(ctx, "torrents/reannounce?"+v.Encode(), nil)
	if err != nil {
		return errors.Wrapf(err, "could not reannounce torrents by hashes: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}

func (c *Client) Pause(ctx context.Context, hashes []string) error {
	form := url.Values{}
	// Add hashes together with | separator
	form.Add("hashes", strings.Join(hashes, "|"))

	resp, err := c.postFormCtx(ctx, "torrents/pause", form)
	if err != nil {
		return errors.Wrapf(err, "could not pause torrents: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}

func (c *Client) Resume(ctx context.Context, hashes []string) error {
	form := url.Values{}
	// Add hashes together with | separator
	form.Add("hashes", strings.Join(hashes, "|"))

	resp, err := c.postFormCtx(ctx, "torrents/resume", form)
	if err != nil {
		return errors.Wrapf(err, "could not resume torrents: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}

func (c *Client) SetCategory(ctx context.Context, hashes []string, category string) error {
	form := url.Values{}
	// Add hashes together with | separator
	form.Add("hashes", strings.Join(hashes, "|"))
	form.Add("category", category)

	resp, err := c.postFormCtx(ctx, "torrents/setCategory", form)
	if err != nil {
		return errors.Wrapf(err, "could not set category for torrents: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}

func (c *Client) SetTag(ctx context.Context, hashes []string, tag string) error {
	form := url.Values{}
	// Add hashes together with | separator
	form.Add("hashes", strings.Join(hashes, "|"))
	form.Add("tags", tag)

	resp, err := c.postFormCtx(ctx, "torrents/addTags", form)
	if err != nil {
		return errors.Wrapf(err, "could not set tag for torrents: %v", hashes)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read body")
		}

		return errors.New(fmt.Sprintf("unexpected response status: %d body: %s", resp.StatusCode, bodyBytes))
	}

	return nil
}
