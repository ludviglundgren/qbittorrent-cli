package qbittorrent

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

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

	// add the content-type so qbittorrent knows what to expect
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
