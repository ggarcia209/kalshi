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

const (
	APIDemoURL = "https://demo-api.kalshi.co/trade-api/v2/"
	APIProdURL = "https://trading-api.kalshi.com/trade-api/v2/"
)

type Cents int

func (c Cents) String() string {
	dollars := float32(c) / 100
	return fmt.Sprintf("$%.2f", dollars)
}

// Client must be instantiated via New.
type Client struct {
	// BaseURL is one of APIDemoURL or APIProdURL.
	BaseURL string

	// See https://trading-api.readme.io/reference/tiers-and-rate-limits.
	WriteRatelimit *rate.Limiter
	ReadRateLimit  *rate.Limiter

	httpClient *http.Client
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

func jsonRequestHeaders(
	ctx context.Context,
	client *http.Client,
	headers http.Header,
	method string, reqURL string,
	jsonReq any, jsonResp any,
) error {
	reqBodyByt, err := json.Marshal(jsonReq)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(reqBodyByt))
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	if headers != nil {
		req.Header = headers
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()

	respBodyByt, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	if resp.StatusCode >= 400 {
		return NewHttpError(resp.StatusCode, string(respBodyByt))
	}

	if client.Jar != nil {
		u, err := url.Parse(reqURL)
		if err != nil {
			return fmt.Errorf("url.Parse: %w", err)
		}
		client.Jar.SetCookies(u, resp.Cookies())
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
	ctx context.Context, r request,
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
	if r.Method == "GET" {
		if !c.ReadRateLimit.Allow() {
			return ErrReadLimitExceeded
		}
	} else {
		if !c.WriteRatelimit.Allow() {
			return ErrWriteLimitExceeded
		}
	}

	if err := jsonRequestHeaders(
		ctx,
		c.httpClient,
		nil,
		r.Method,
		u.String(), r.JSONRequest, r.JSONResponse,
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

func basicRateLimit() *rate.Limiter {
	return rate.NewLimiter(rate.Every(time.Second), 10)
}

// NewClient creates a new Kalshi client. Login must be called to authenticate the
// the client before any other request.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar.New: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{
			Jar: jar,
		},
		BaseURL: baseURL,
		// See https://trading-api.readme.io/reference/tiers-and-rate-limits.
		// Default to Basic access.
		WriteRatelimit: basicRateLimit(),
		ReadRateLimit:  basicRateLimit(),
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
