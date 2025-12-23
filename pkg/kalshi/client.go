package kalshi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/time/rate"
)

type Cents int

func (c Cents) String() string {
	dollars := float32(c) / 100
	return fmt.Sprintf("$%.2f", dollars)
}

const ( // denotes endpoints requiring authorization
	authenticated   = true
	unauthenticated = false
)

// Client must be instantiated via New.
type Client struct {
	// BaseURL is one of APIDemoURL or APIProdURL.
	BaseURL string

	// See https://trading-api.readme.io/reference/tiers-and-rate-limits.
	RateLimit *rate.Limiter

	httpClient    *http.Client
	requestSigner *KeySigner
}

type CursorResponse struct {
	Cursor string `json:"cursor"`
}

type CursorRequest struct {
	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
}

type request struct {
	CursorRequest
	Method       string
	Endpoint     string
	QueryParams  any
	JSONRequest  any
	JSONResponse any
}

func (c *Client) jsonRequestHeaders(
	ctx context.Context,
	headers http.Header,
	method string, reqURL string,
	jsonReq any, jsonResp any,
	auth bool,
) error {
	// get body if non-nil
	var body io.Reader = nil
	if jsonReq != nil {
		reqBodyByt, err := json.Marshal(jsonReq)
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}
		body = bytes.NewReader(reqBodyByt)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	if headers != nil {
		req.Header = headers
	}

	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	// sign request
	if auth {
		if err := c.requestSigner.SignRequestWithRSAKey(req); err != nil {
			return fmt.Errorf("c.requestSigner.SignRequestWithRSAKey: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("c.httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	respBodyByt, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if resp.StatusCode >= 400 {
		return NewHttpError(resp.StatusCode, string(respBodyByt))
	}

	if c.httpClient.Jar != nil {
		u, err := url.Parse(reqURL)
		if err != nil {
			return fmt.Errorf("url.Parse: %w", err)
		}
		c.httpClient.Jar.SetCookies(u, resp.Cookies())
	}

	if jsonResp != nil {
		err = json.Unmarshal(respBodyByt, jsonResp)
		if err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
	}
	return nil
}

func (c *Client) request(
	ctx context.Context, r request, auth bool,
) error {
	u, err := url.Parse(c.BaseURL + r.Endpoint)
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	if r.QueryParams != nil {
		v, err := query.Values(r.QueryParams)
		if err != nil {
			return fmt.Errorf("query.Values: %w", err)
		}
		u.RawQuery = v.Encode()
	}

	// Do not block via Wait! Trades have to be
	// fast to be meaningful!
	if !c.RateLimit.Allow() {
		return ErrRateLimitExceeded
	}

	if err := c.jsonRequestHeaders(
		ctx,
		nil,
		r.Method,
		u.String(), r.JSONRequest, r.JSONResponse,
		auth,
	); err != nil {
		return fmt.Errorf("jsonRequestHeaders: %w", err)
	}

	return nil
}

// Timestamp represents a POSIX Timestamp in seconds.
type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t)
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(int64(i), 0))
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(time.Time(t).UTC().Unix()))), nil
}

func newRateLimit(rps int) *rate.Limiter {
	return rate.NewLimiter(rate.Every(time.Second), rps)
}

// NewClient creates a new Kalshi c.httpClient. Login must be called to authenticate the
// the client before any other request.
func NewClient(baseURL, keyId, keyFilePath, key string, useKeyFile bool, rps int) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar.New: %w", err)
	}

	requestSigner := NewKeySigner(keyFilePath, key, keyId, useKeyFile)

	c := &Client{
		httpClient: &http.Client{
			Jar: jar,
		},
		BaseURL:       baseURL,
		RateLimit:     newRateLimit(rps),
		requestSigner: requestSigner,
	}

	return c, nil
}

// Time is a time.Time that tolerates additional '"' characters.
// Kalshi API endpoints use both RFC3339 and POSIX
// timestamps.
type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if len(bytes.Trim(b, "\"")) == 0 {
		return nil
	}
	err := t.Time.UnmarshalJSON(b)
	if err != nil {
		return fmt.Errorf("t.Time.UnmarshalJSON: %w", err)
	}
	return nil
}

// Side is either Yes or No.
type Side string

const (
	Yes Side = "yes"
	No  Side = "no"
)

// SideBool turns a Yes bool into a Side.
func SideBool(yes bool) Side {
	if yes {
		return Yes
	}
	return No
}
