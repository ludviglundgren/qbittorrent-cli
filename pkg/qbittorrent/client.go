package qbittorrent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"golang.org/x/net/publicsuffix"
)

type Client struct {
	settings Settings
	http     *http.Client
}

type Settings struct {
	Hostname string
	Port     uint
	Username string
	Password string
}

func NewClient(s Settings) *Client {
	jarOptions := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	//store cookies in jar
	jar, err := cookiejar.New(jarOptions)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Timeout: time.Second * 10,
		Jar:     jar,
	}
	return &Client{
		settings: s,
		http:     httpClient,
	}
}

func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	reqUrl := fmt.Sprintf("http://%v:%v/api/v2/%v", c.settings.Hostname, c.settings.Port, endpoint)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	// add optional parameters that the user wants
	form := url.Values{}
	if opts != nil {
		for k, v := range opts {
			form.Add(k, v)
		}
	}

	reqUrl := fmt.Sprintf("http://%v:%v/api/v2/%v", c.settings.Hostname, c.settings.Port, endpoint)
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	//// add the content-type so qbittorrent knows what to expect
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) postFile(endpoint string, fileName string, opts map[string]string) (*http.Response, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	// Close the file later
	defer file.Close()

	// Buffer to store our request body as bytes
	var requestBody bytes.Buffer

	// Create a multipart writer
	multiPartWriter := multipart.NewWriter(&requestBody)

	// Initialize file field
	fileWriter, err := multiPartWriter.CreateFormFile("torrents", fileName)
	if err != nil {
		log.Fatalf("error initializing file field %v", err)
	}

	// Copy the actual file content to the fields writer
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("could not copy file to writer %v", err)
	}

	// Populate other fields
	if opts != nil {
		for key, val := range opts {
			fieldWriter, err := multiPartWriter.CreateFormField(key)
			if err != nil {
				log.Fatalf("could not add other fields %v", err)
			}

			_, err = fieldWriter.Write([]byte(val))
			if err != nil {
				log.Fatalf("could not write field %v", err)
			}
		}
	}

	// Close multipart writer
	multiPartWriter.Close()

	reqUrl := fmt.Sprintf("http://%v:%v/api/v2/%v", c.settings.Hostname, c.settings.Port, endpoint)
	req, err := http.NewRequest("POST", reqUrl, &requestBody)
	if err != nil {
		log.Fatalf("could not create request object %v", err)
	}

	// Set correct content type
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	res, err := c.http.Do(req)
	if err != nil {
		log.Fatalf("could not perform request %v", err)
	}

	return res, nil
}

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

func (c *Client) GetTorrentsRaw() (string, error) {
	resp, err := c.get("torrents/info", nil)
	if err != nil {
		log.Fatalf("error fetching torrents: %v", err)
	}

	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	return string(data), nil
}

// AddTorrentFromFile add new torrent from torrent file
func (c *Client) AddTorrentFromFile(file string, options map[string]string) (hash string, err error) {
	// Get meta info from file to find out the hash for later use
	t, err := metainfo.LoadFromFile(file)
	if err != nil {
		log.Fatalf("could not open file %v", err)
	}

	// Get hash from info
	torrentHash := metainfo.HashBytes(t.InfoBytes)
	hash = torrentHash.String()
	if hash == "" {
		return "", errors.New("could not stringify torrent hash")
	}

	res, err := c.postFile("torrents/add", file, options)
	if err != nil {
		return "", err
	} else if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()

	return hash, nil
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
