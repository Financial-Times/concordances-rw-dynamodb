package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"io"
)

type MockService struct {
	model   Model
	created bool
	deleted bool
	count   int64
	err     error
}

func (mock *MockService) Read(uuid string) (Model, error) {
	return mock.model, mock.err
}

func (mock *MockService) Write(m Model) (bool, error) {
	return mock.created, mock.err
}
func (mock *MockService) Delete(uuid string) (bool, error) {
	return mock.deleted, mock.err
}
func (mock *MockService) Count() (int64, error) {
	return mock.count, mock.err
}

type mockHttpClient struct {
	resp       string
	statusCode int
	err        error
}

func (c mockHttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	cb := ioutil.NopCloser(bytes.NewReader([]byte(c.resp)))
	return &http.Response{Body: cb, StatusCode: c.statusCode}, c.err
}

func newRequest(method, url string, body string) *http.Request {
	var payload io.Reader
	if body != "" {
		payload = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		panic(err)
	}
	return req
}
