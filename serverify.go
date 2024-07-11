package serverify

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

// Serverify is top level struct to interact with Serverify API
type Serverify struct {
	client *resty.Client
}

// New create a Serverify instance with the given base URL
func New(baseURL string) *Serverify {
	client := resty.New().SetBaseURL(baseURL)
	return &Serverify{
		client: client,
	}
}

type Error struct {
	StatusCode     int
	ServerifyError struct {
		Message string
	}
}

func (e Error) Error() string {
	return e.ServerifyError.Message
}

// CreateSession creates new session with the given name
func (s *Serverify) CreateSession(name string) (*Session, error) {
	res := struct {
		Session string `json:"session"`
	}{}
	err := s.doRequest(resty.MethodPost, "/session", map[string]string{"session": name}, &res)
	if err != nil {
		return nil, err
	}

	return &Session{
		name:      res.Session,
		serverify: s,
	}, nil
}

type errorResponse struct {
	ServerifyError struct {
		Message string `json:"message"`
	} `json:"serverify_error"`
}

func (s *Serverify) doRequest(method, path string, body any, result any) error {
	errRes := errorResponse{}
	// req := s.client.R().SetResult(result).SetError(errRes)
	req := s.client.R().SetError(&errRes)
	if body != nil {
		req = req.SetBody(body)
	}
	if result != nil {
		req = req.SetResult(result)
	}

	res, err := req.Execute(method, path)
	if err != nil {
		return err
	}

	if res.IsError() {
		return Error{
			StatusCode: res.StatusCode(),
			ServerifyError: struct {
				Message string
			}{
				Message: errRes.ServerifyError.Message,
			},
		}
	}

	return nil
}

// Session is created by Serverify.CreateSession
type Session struct {
	name      string
	serverify *Serverify
}

// Name returns name of the session
func (s *Session) Name() string {
	return s.name
}

// Log represents the request log
type Log struct {
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Path        string            `json:"path"`
	Query       map[string]string `json:"query"`
	Body        string            `json:"body"`
	RequestedAt time.Time         `json:"requestedAt"`
}

// Logs represents request logs of the session
type Logs struct {
	Histories []Log `json:"histories"`
}

// Logs returns request logs of the session
func (s *Session) Logs() (*Logs, error) {
	logs := Logs{}
	err := s.serverify.doRequest(resty.MethodGet, fmt.Sprintf("/session/%s", s.Name()), nil, &logs)
	if err != nil {
		return nil, err
	}

	return &logs, nil
}

// Delete deletes the session
func (s *Session) Delete() error {
	return s.serverify.doRequest(resty.MethodDelete, fmt.Sprintf("/session/%s", s.Name()), nil, nil)
}

// BaseURL returns the base URL of mock endpoint for this session
func (s *Session) BaseURL() string {
	baseURL, _ := url.JoinPath(s.serverify.client.BaseURL, "mock", s.Name())
	return baseURL
}
