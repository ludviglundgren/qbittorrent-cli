package qbittorrent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

type QbittorrentClient interface {
	Login() error
	GetTorrents() ([]Torrent, error)
}

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

func (c *Client) Login() error {
	credentials := make(map[string]string)
	credentials["username"] = c.settings.Username
	credentials["password"] = c.settings.Password

	resp, err := c.post("auth/login", credentials)
	if err != nil {
		log.Fatalf("login error", err)
	} else if resp.StatusCode != http.StatusOK { // check for correct status code
		log.Fatalf("login error bad status %v\n", err)
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
		log.Fatalf("error fetching torrents", err)
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
	resp, err := c.http.Get("http://192.168.60.5:10039/api/v2/torrents/info")
	if err != nil {
		log.Fatalf("error fetching torrents", err)
	}

	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	return string(data), nil
}

func (c *Client) AddTorrentsFromFile() error {
	return nil
}
