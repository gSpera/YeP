package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewPaste(t *testing.T) {
	tm := []struct {
		name       string
		method     string
		code       int
		input      newPasteRequest
		output     newPasteResponse
		deepCheck  bool
		disableLog bool
	}{
		{
			"OK",
			"POST",
			http.StatusOK,
			newPasteRequest{
				ExpireTime: getExpireTime(t),
				Code:       "example paste",
			},
			newPasteResponse{true, "", ""},
			false,
			false,
		},
		{
			"GET not allowed",
			"GET",
			http.StatusMethodNotAllowed,
			newPasteRequest{},
			newPasteResponse{
				OK:    false,
				Error: "Method not allowed",
				Path:  "",
			},
			true,
			false,
		},
		{
			"Empty paste",
			"POST",
			http.StatusInternalServerError,
			newPasteRequest{
				ExpireTime: getExpireTime(t),
				Code:       "",
			},
			newPasteResponse{false, ErrInternalServerError, ""},
			true,
			true,
		},
		{
			"ExpireTime not valid",
			"POST",
			http.StatusBadRequest,
			newPasteRequest{
				ExpireTime: "1",
				Code:       "example paste",
			},
			newPasteResponse{false, ErrExpireTimeNotValid.Error(), ""},
			true,
			false,
		},
	}

	server := NewServer(MemoryDB{}, defaultCfg)

	for _, tt := range tm {
		t.Run(tt.name, func(t *testing.T) {
			if tt.disableLog {
				log.SetOutput(ioutil.Discard)
				defer log.SetOutput(os.Stderr)
			}
			res := httptest.NewRecorder()

			inputBytes, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Could not encode input: %v", err)
			}

			input := bytes.NewReader(inputBytes)
			req := httptest.NewRequest(tt.method, "/api/new", input)

			handleAPINewPaste(server, res, req)

			if res.Code != tt.code {
				t.Errorf("Wrong code: expected: %d; got: %d", tt.code, res.Code)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Could not read body: %v", err)
			}

			output := newPasteResponse{}
			if err := json.Unmarshal(body, &output); err != nil {
				t.Fatalf("Could not decode output: %v", err)
			}

			if tt.output.OK != output.OK {
				t.Errorf("Expected: %v; got: %v, error: %v", tt.output.OK, output.OK, output.Error)
			}

			if tt.deepCheck && (output != tt.output) {
				t.Fatalf("Outputs are different: expected: %+v; got: %+v", tt.output, output)
			}
		})
	}

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	t.Run("Bad Writer", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/new", new(bytes.Buffer))
		handleAPINewPaste(server, badWriter{}, req)
	})

	t.Run("Bad Reader", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/new", badReader{})
		writer := httptest.NewRecorder()
		handleAPINewPaste(server, writer, req)

		if writer.Code == http.StatusOK {
			t.Error("Request accepted")
		}
	})
}
func TestGetPaste(t *testing.T) {
	tm := []struct {
		name      string
		code      int
		deepCheck bool
		input     getPasteRequest
		output    getPasteResponse
	}{
		{
			"Simple",
			http.StatusOK,
			true,
			getPasteRequest{
				Name:   "test",
				Render: false,
			},
			getPasteResponse{
				OK:      true,
				Error:   "",
				Name:    "test",
				Code:    "test",
				User:    "test",
				Created: 0,
				Expire:  0,
			},
		},
		{
			"Render",
			http.StatusOK,
			true,
			getPasteRequest{
				Name:   "test",
				Render: true,
			},
			getPasteResponse{
				OK:      true,
				Error:   "",
				Name:    "test",
				Code:    "test",
				User:    "test",
				Created: 0,
				Expire:  0,
				Render:  "<h1>test</h1>",
				Style:   "",
			},
		},
		{
			"Not Found",
			http.StatusNotFound,
			true,
			getPasteRequest{
				Name:   "notfound",
				Render: false,
			},
			getPasteResponse{
				OK:    false,
				Error: "Paste not found",
			},
		},
	}

	server := NewServer(NewTestDB(), defaultCfg)

	for _, tt := range tm {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			inputBytes, _ := json.Marshal(tt.input)
			input := bytes.NewReader(inputBytes)

			req := httptest.NewRequest("GET", "/api/get", input)

			handleAPIGetPaste(server, res, req)

			if res.Code != tt.code {
				t.Errorf("Wrong code: expected: %d; got: %d", tt.code, res.Code)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Could not read body: %v", err)
			}

			output := getPasteResponse{}
			if err := json.Unmarshal(body, &output); err != nil {
				t.Fatalf("Could not decode output: %v", err)
			}

			if tt.output.OK != output.OK {
				t.Errorf("Expected: %v; got: %v, error: %v", tt.output.OK, output.OK, output.Error)
			}

			if tt.deepCheck && (output != tt.output) {
				t.Fatalf("Outputs are different: expected: %+v; got: %+v", tt.output, output)
			}
		})
	}

	t.Run("Method not allowed", func(t *testing.T) {
		res := httptest.NewRecorder()

		inputBytes, _ := json.Marshal([]byte("{}"))
		input := bytes.NewReader(inputBytes)

		req := httptest.NewRequest("POST", "/api/get", input)

		handleAPIGetPaste(server, res, req)
		if res.Code == http.StatusOK {
			t.Error("Request accepted")
		}
	})

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	t.Run("Bad Writer", func(t *testing.T) {

		inputBytes, _ := json.Marshal([]byte("{}"))
		input := bytes.NewReader(inputBytes)

		req := httptest.NewRequest("GET", "/api/get", input)

		handleAPIGetPaste(server, badWriter{}, req)
	})

	t.Run("Bad Reader", func(t *testing.T) {
		res := httptest.NewRecorder()

		req := httptest.NewRequest("GET", "/api/get", badReader{})

		handleAPIGetPaste(server, res, req)

		if res.Code == http.StatusOK {
			t.Error("Request accepted")
		}
	})
}
