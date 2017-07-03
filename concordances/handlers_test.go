package concordances

import (
	"bytes"
	"errors"
	"fmt"
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	TestConceptUuid = "4f50b156-6c50-4693-b835-02f70d3f3bc0"
	Path            = "/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0"
	GoodBody        = "{\"uuid\":\"4f50b156-6c50-4693-b835-02f70d3f3bc0\",\"concordedIds\":[\"1\",\"2\"]}\n"
)

var router *mux.Router
var h Handler

func init() {
	router = mux.NewRouter()
	srv := &MockService{}
	h = Handler{srv: srv}
	h.registerAPIHandlers(router)
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

func TestHandler_ResponseCodesAndMessages(t *testing.T) {
	testCases := []struct{
		description          string
		request              *http.Request
		expectedResponseCode int
		expectedContentType  string
		expectedResponseBody string
		service              Service
		errorString              string
	}{
		{
			description:          "GET 503 Service Not Available",
			request:              newRequest("GET", Path, ""),
			service:              &MockService{err: errors.New("")},
			expectedResponseCode: 503,
			expectedContentType:  ContentTypeJson,
			errorString:          "Error retrieving concordances",
		},
		{
			description:          "PUT 503 Service Not Available",
			request:              newRequest("PUT", Path, GoodBody),
			service:              &MockService{err: errors.New("")},
			expectedResponseCode: 503,
			expectedContentType:  ContentTypeJson,
			errorString:          "Error writing concordance",
		},
		{
			description:          "DELETE 503 Service Not Available",
			request:              newRequest("DELETE", Path, GoodBody),
			service:              &MockService{err: errors.New("")},
			expectedResponseCode: 503,
			expectedContentType:  ContentTypeJson,
			errorString:          "Error deleting concordance",
		},
		{
			description:          "GET 404 Not Found",
			request:              newRequest("GET", Path, ""),
			service:              &MockService{model: db.ConcordancesModel{}, status:db.CONCORDANCE_NOT_FOUND},
			expectedResponseCode: 404,
			expectedContentType:  ContentTypeJson,
		},
		{
			description:          "DELETE 404 Not Found",
			request:              newRequest("DELETE", Path, ""),
			service:              &MockService{status: db.CONCORDANCE_NOT_FOUND},
			expectedResponseCode: 404,
			expectedContentType:  ContentTypeJson,
		},
		{
			description:          "DELETE 204 Deleted",
			request:              newRequest("DELETE", Path, ""),
			service:              &MockService{status: db.CONCORDANCE_DELETED},
			expectedResponseCode: 204,
		},
		{
			description:          "PUT 201 Created",
			request:              newRequest("PUT", Path, GoodBody),
			service:              &MockService{status: db.CONCORDANCE_CREATED},
			expectedResponseCode: 201,
		},
		{
			description:          "PUT 200 Updated",
			request:              newRequest("PUT", Path, GoodBody),
			service:              &MockService{status: db.CONCORDANCE_UPDATED},
			expectedResponseCode: 200,
		},
		{
			description:          "GET 200 OK",
			request:              newRequest("GET", Path, ""),
			service:              &MockService{model: db.ConcordancesModel{UUID: TestConceptUuid, ConcordedIds: []string{"1", "2"}}},
			expectedResponseCode: 200,
			expectedResponseBody: GoodBody,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description,
			func(t *testing.T) {
				h.srv = testCase.service
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, testCase.request)

				assert.Equal(t, testCase.expectedResponseCode, rec.Result().StatusCode, "Response code incorrect.")
				if testCase.errorString != "" {
					assert.Contains(t, fmt.Sprintf("\"{\"message\":\"%s\"}", testCase.errorString), rec.Body.String(), "Response body incorrect.")
				} else if testCase.expectedResponseCode == 404 {
					assert.Contains(t, "{\"message\":\"Unable to find concordance\"}", rec.Body.String(), "Response body incorrect.")
				} else {
					assert.Equal(t, testCase.expectedResponseBody, rec.Body.String(), "Response body incorrect.")
				}
			})
	}
}

func TestHandler_BadPath(t *testing.T) {
	invalidPaths := []string{
		"/concordances/invalidUUID",
		"/not_concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0",
		"/4f50b156-6c50-4693-b835-02f70d3f3bc0",
		"/dfsdf",
		"/concordances",
		"/concordances/",
		"/",
	}
	methods := []string{"GET", "PUT", "DELETE"}

	for _, p := range invalidPaths {
		for _, m := range methods {
			t.Run(fmt.Sprintf("%s: %s", m, p),
				func(t *testing.T) {
					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, newRequest(m, p, ""))
					assert.Equal(t, 404, rec.Result().StatusCode, "Response code incorrect.")
				})
		}
	}
}

func TestHandler_BadBody(t *testing.T) {
	invalidPayloads := []struct {
		desc           string
		request        *http.Request
		path           string
		expectedErrMsg string
	}{
		{desc: "UUID in payload is different from UUID path parameter",
			request:        newRequest("PUT", "/concordances/7c4b3931-361f-4ea4-b694-75d1630d7746", "{\"uuid\": \"4f50b156-6c50-4693-b835-02f70d3f3bc0\", \"concordedIds\": [\"1\"]}"),
			expectedErrMsg: "{\"message\":\"Concept UUID (4f50b156-6c50-4693-b835-02f70d3f3bc0) in payload is different from UUID path parameter (7c4b3931-361f-4ea4-b694-75d1630d7746)\"}"},
		{desc: "ConceptId not found in payload", request: newRequest("PUT", Path, "{\"concordedIds\": [\"1\"]}"),
			expectedErrMsg: "{\"message\":\"Concept UUID is missing from the Payload\"}"},
		{desc: "concordedIds is an empty array", request: newRequest("PUT", Path, "{\"uuid\": \"4f50b156-6c50-4693-b835-02f70d3f3bc0\"}"),
			expectedErrMsg: "{\"message\":\"Payload has no concorded UUIDs to store\"}"},
		{desc: "concordedIds is null", request: newRequest("PUT", Path, "{\"uuid\": \"4f50b156-6c50-4693-b835-02f70d3f3bc0\", \"concordedIds\": null}"),
			expectedErrMsg: "{\"message\":\"Payload has no concorded UUIDs to store\"}"},
		{desc: "Invalid JSON", request: newRequest("PUT", Path, "{\"uuid\": \"4f50b156-6c50-4693-b835-02f70d3f3bc0\", \"}"),
			expectedErrMsg: "{\"message\":\"Error decoding the JSON of the request body\"}"},
	}

	for _, c := range invalidPayloads {
		t.Run(c.desc,
			func(t *testing.T) {
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, c.request)
				assert.Equal(t, 400, rec.Result().StatusCode, "Response code incorrect.")
				assert.Equal(t, ContentTypeJson, rec.HeaderMap["Content-Type"][0], "Incorrect Content-Type Header")
				assert.Equal(t, c.expectedErrMsg, rec.Body.String(), "Response body incorrect.")
			})
	}
}

func TestHandler_AdminEndpoints(t *testing.T) {
	adminHandlers := map[string]string{
		status.BuildInfoPath: "",
		status.GTGPath:       "",
		healthPath:           "",
	}
	router := mux.NewRouter()
	NewHandler(router, AppConfig{}, &MockService{})

	for url, expectedBody := range adminHandlers {
		t.Run(url,
			func(t *testing.T) {
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, newRequest("GET", url, ""))
				assert.Equal(t, 200, rec.Result().StatusCode)
				if expectedBody != "" {
					assert.Equal(t, expectedBody, rec.Body.String())
				}
			})
	}
}
