package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
)

func newTestRoutes() (*gin.Engine, error) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()
	mock.ExpectPing().SetVal("connected")
	canvas := make([]byte, TotalBytes)
	mock.ExpectGet(CanvasName).SetVal(string(canvas))
	return SetupRouter(db)
}

func TestRoutes(t *testing.T) {
	r, err := newTestRoutes()
	if err != nil {
		t.Fatal("Problem initializing redis:", err)
	}
	tc := []struct {
		testName string
		path     string
		method   string
		headers  map[string]string
		body     io.Reader
		expected func(*httptest.ResponseRecorder) error
	}{
		{
			"test canvas endpoint",
			"/canvas",
			"GET",
			nil,
			strings.NewReader(""),
			expectedCanvas,
		},
		{
			"test palette endpoint",
			"/palette",
			"GET",
			nil,
			strings.NewReader(""),
			expectedPalette,
		},
	}
	for _, test := range tc {
		t.Run(test.testName, func(t *testing.T) {
			record := httptest.NewRecorder()
			request, err := http.NewRequest(test.method, test.path, test.body)
			if err != nil {
				t.Error("test format incorrect", err)
			}
			r.ServeHTTP(record, request)
			if err = test.expected(record); err != nil {
				t.Error(err)
			}
		})
	}
}

func expectedCanvas(recorder *httptest.ResponseRecorder) error {
	if recorder.Code != 200 {
		return errors.New("bad status code")
	}
	length, err := io.ReadAll(recorder.Body)
	if err != nil {
		return err
	}
	if len(length) != TotalBytes {
		return errors.New("bad canvas length")
	}
	return nil
}

func expectedPalette(recorder *httptest.ResponseRecorder) error {
	if recorder.Code != 200 {
		return errors.New("bad status code")
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		return fmt.Errorf("bad content type: %s", ct)
	}
	all, err := io.ReadAll(recorder.Body)
	if err != nil {
		return err
	}
	var paletteResponse []uint32
	if json.Unmarshal(all, &paletteResponse) != nil {
		return err
	}
	for i := range palette {
		if palette[i] != paletteResponse[i] {
			return fmt.Errorf("expected palette %v, but got instead %v", palette, paletteResponse)
		}
	}
	return nil
}
