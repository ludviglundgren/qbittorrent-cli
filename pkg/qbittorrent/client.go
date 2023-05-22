package qbittorrent

import (
	"bytes"
	"crypto/tls"
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

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	hostname      string
	port          uint
	addr          string
	username      string
	password      string
	basicUser     string
	basicPass     string
	tlsSkipVerify bool

	http *http.Client
}

type Settings struct {
	Hostname      string
	Port          uint
	Addr          string
	Username      string
	Password      string
	BasicUser     string
	BasicPass     string
	TLSSkipVerify bool
}

func NewClient(s Settings) *Client {
	jarOptions := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
	//store cookies in jar
	jar, err := cookiejar.New(jarOptions)
	if err != nil {
		log.Fatal(err)
	}

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: s.TLSSkipVerify,
		},
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Jar:       jar,
		Transport: t,
	}

	return &Client{
		hostname:      s.Hostname,
		port:          s.Port,
		addr:          s.Addr,
		username:      s.Username,
		password:      s.Password,
		basicUser:     s.BasicUser,
		basicPass:     s.BasicPass,
		tlsSkipVerify: s.TLSSkipVerify,
		http:          httpClient,
	}
}

func (c *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	reqUrl, err := c.buildUrl(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	// set basic auth
	c.setBasicAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) getCtx(ctx context.Context, endpoint string, values url.Values) (*http.Response, error) {
	reqUrl, err := c.buildUrl(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(reqUrl)
	if err != nil {
		log.Fatal(err)
	}

	u.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// set basic auth
	c.setBasicAuth(req)

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

	reqUrl, err := c.buildUrl(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// set basic auth
	c.setBasicAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) postCtx(ctx context.Context, endpoint string, opts map[string]string) (*http.Response, error) {
	// add optional parameters that the user wants
	form := url.Values{}
	if opts != nil {
		for k, v := range opts {
			form.Add(k, v)
		}
	}

	reqUrl, err := c.buildUrl(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// set basic auth
	c.setBasicAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func (c *Client) postFileCtx(ctx context.Context, endpoint string, fileName string, opts map[string]string) (*http.Response, error) {
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

	reqUrl, err := c.buildUrl(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, &requestBody)
	if err != nil {
		log.Fatalf("could not create request object %v", err)
	}

	// Set correct content type
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// set basic auth
	c.setBasicAuth(req)

	res, err := c.http.Do(req)
	if err != nil {
		log.Fatalf("could not perform request %v", err)
	}

	return res, nil
}

func (c *Client) buildUrl(endpoint string) (string, error) {
	reqUrl, err := url.JoinPath(c.addr, "/api/v2", endpoint)
	if err != nil {
		return "", errors.Wrap(err, "could not build addr")
	}

	if c.hostname != "" && c.port != 0 {
		reqUrl = fmt.Sprintf("http://%s:%d/api/v2/%s", c.hostname, c.port, endpoint)
	}

	return reqUrl, nil
}

func (c *Client) setBasicAuth(req *http.Request) {
	if c.basicUser != "" && c.basicPass != "" {
		req.SetBasicAuth(c.basicUser, c.basicPass)
	}
}
