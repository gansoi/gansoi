package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	Client struct {
		baseUrl string
		key     string
	}
)

var (
	ErrNotFound            = errors.New("Ressource not found")
	ErrForbidden           = errors.New("Forbidden")
	ErrInternalServerError = errors.New("Internal server error")
)

func httpStatusToError(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusInternalServerError:
		fallthrough
	default:
		return ErrInternalServerError
	}
}

func NewClient(baseUrl string, key string) *Client {
	return &Client{
		baseUrl: baseUrl,
		key:     key,
	}
}

func (r *Client) request(method, urlStr string, in interface{}, out interface{}) error {
	client := &http.Client{}
	var reader io.Reader

	if in != nil {
		b, _ := json.Marshal(in)
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Secret", r.key)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = httpStatusToError(resp.StatusCode)
	if err != nil {
		return err
	}

	if out != nil {
		d := json.NewDecoder(resp.Body)
		err = d.Decode(&out)
		if err != nil {
			return err
		}
	}

	return resp.Body.Close()
}

func (r *Client) Create(v interface{}, created interface{}) error {
	return r.request("POST", r.baseUrl+"new", v, created)
}

func (r *Client) Read(id string, v interface{}) error {
	err := r.request("GET", r.baseUrl+id, nil, v)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	return err
}

func (r *Client) Update(id string, v interface{}) error {
	return r.request("PUT", r.baseUrl+id, v, nil)
}

func (r *Client) Delete(id string) error {
	return r.request("DELETE", r.baseUrl+id, nil, nil)
}

func (r *Client) List(v interface{}) error {
	return r.Read("", v)
}
