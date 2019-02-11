package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TempRecord struct {
	Hex           string
	Squawk        string
	Flight        string
	Latitude      float64 `json:"lat"`
	Longitude     float64 `json:"lon"`
	ValidPosition int     `json:"validposition"`
	Altitude      int
	VerticalRate  int `json:"vert_rate"`
	Track         int
	ValidTrack    int `json:"validtrack"`
	Speed         int
	Messages      int
	Seen          int
}

type Record struct {
	Hex           string
	Squawk        string
	Flight        string
	Latitude      float64
	Longitude     float64
	ValidPosition bool
	Altitude      int
	VerticalRate  int
	Track         int
	ValidTrack    bool
	Speed         int
	Messages      int
	Seen          int
}

type Client struct {
	URL        *url.URL
	HTTPClient *http.Client
	Logger     *log.Logger
}

func NewClient(urlStr string, logger *log.Logger) (*Client, error) {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", urlStr)
	}

	if logger == nil {
		var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
		logger = discardLogger
	}

	return &Client{
		URL: parsedURL,
		HTTPClient: &http.Client{
			Timeout: time.Duration(10 * time.Second),
		},
		Logger: logger,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Accept", "application/json")
	// TODO: req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func (c *Client) GetRecords(ctx context.Context) (*[]Record, error) {
	req, err := c.newRequest(ctx, "GET", "/dump1090/data.json", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check status code here…

	var tempRecords []TempRecord
	if err := decodeBody(res, &tempRecords); err != nil {
		return nil, err
	}

	var records []Record
	for _, record := range tempRecords {
		records = append(records, Record{
			Hex:           record.Hex,
			Squawk:        record.Squawk,
			Flight:        strings.Trim(record.Flight, " "),
			Latitude:      record.Latitude,
			Longitude:     record.Longitude,
			ValidPosition: record.ValidPosition != 0,
			Altitude:      record.Altitude,
			VerticalRate:  record.VerticalRate,
			Track:         record.Track,
			ValidTrack:    record.ValidTrack != 0,
			Speed:         record.Speed,
			Messages:      record.Messages,
			Seen:          record.Seen,
		})
	}

	return &records, nil
}
