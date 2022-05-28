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

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := redismock.NewClientMock()
	mock.ExpectPing().SetVal("connected")
	canvas := make([]byte, TotalBytes)
	mock.ExpectGet(CanvasName).SetVal(string(canvas))
	mock.Regexp().ExpectBitField(CanvasName, "SET", "u32", "^[0-9]+$", "^[0-9]+$").SetVal([]int64{0})
	r, err := SetupRouter(db)
	if err != nil {
		t.Fatal("Problem initializing redis:", err)
	}
	formatStr := "{\"x\": %d, \"y\": %d, \"color\": %d}"
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
			http.MethodGet,
			nil,
			strings.NewReader(""),
			expectedCanvas,
		},
		{
			"test palette endpoint",
			"/palette",
			http.MethodGet,
			nil,
			strings.NewReader(""),
			expectedPalette,
		},
		{
			"test set pixel invalid object",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader("{\"invalid\": 1}"),
			expectedWrongPixel,
		},
		{
			"test set pixel invalid body",
			"/pixel",
			http.MethodPost,
			nil,
			strings.NewReader("bad body"),
			expectedWrongPixel,
		},
		{
			"test set pixel wrong color",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, 0, 0, palette[0]+1)),
			expectedWrongPixel,
		},
		{
			"test set pixel wrong x value",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, CanvasWidth, 0, palette[0])),
			expectedWrongPixel,
		},
		{
			"test set pixel wrong y value",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, 0, CanvasHeight, palette[0])),
			expectedWrongPixel,
		},
		{
			"test set pixel negative x value",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, -1, 0, palette[0])),
			expectedWrongPixel,
		},
		{
			"test set pixel negative y value",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, 0, -1, palette[0])),
			expectedWrongPixel,
		},
		{
			"test set pixel success",
			"/pixel",
			http.MethodPost,
			map[string]string{"Content-Type": "application/json"},
			strings.NewReader(fmt.Sprintf(formatStr, CanvasWidth-1, CanvasHeight-1, palette[0])),
			expectedSuccess,
		},
	}
	for _, test := range tc {
		t.Run(test.testName, func(t *testing.T) {
			record := httptest.NewRecorder()
			request, err := http.NewRequest(test.method, test.path, test.body)
			if err != nil {
				t.Error("test format incorrect", err)
			}
			if test.headers != nil {
				for k, v := range test.headers {
					request.Header.Set(k, v)
				}
			}
			r.ServeHTTP(record, request)
			if err = test.expected(record); err != nil {
				t.Logf("Record: %+v", record)
				t.Error(err)
			}
		})
	}
}

func expectedCanvas(recorder *httptest.ResponseRecorder) error {
	if recorder.Code != 200 {
		return errors.New("bad status code")
	}
	if len(recorder.Body.Bytes()) != TotalBytes {
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
	var paletteResponse []uint32
	if err := json.Unmarshal(recorder.Body.Bytes(), &paletteResponse); err != nil {
		return err
	}
	for i := range palette {
		if palette[i] != paletteResponse[i] {
			return fmt.Errorf("expected palette %v, but got instead %v", palette, paletteResponse)
		}
	}
	return nil
}

func expectedWrongPixel(recorder *httptest.ResponseRecorder) error {
	if recorder.Code != http.StatusBadRequest {
		return fmt.Errorf("it should return bad request, but returned %d", recorder.Code)
	}
	var m map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &m); err != nil {
		return nil
	}
	if _, ok := m["error"]; !ok {
		return fmt.Errorf("there was no error in the response: %+q", m)
	}
	return nil
}

func expectedSuccess(recorder *httptest.ResponseRecorder) error {
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || body != "OK" {
		return fmt.Errorf("it should return StatusOK, but returned %d, %s", recorder.Code, body)
	}
	return nil
}
