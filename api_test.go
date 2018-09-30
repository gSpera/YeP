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

//Main setups the server for future tests
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
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

	for _, tt := range tm {
		t.Run(tt.name, func(t *testing.T) {
			if tt.disableLog {
				log.SetOutput(ioutil.Discard)
				defer log.SetOutput(os.Stderr)
			}
			server := NewServer(MemoryDB{}, defaultCfg)
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
				t.Fatalf("Outputs are different: expected: %v; got: %v", tt.output, output)
			}
		})
	}
}

func getExpireTime(t *testing.T) string {
	if len(defaultCfg.ExpireAfter) == 0 {
		return "0"
	}
	res, err := defaultCfg.ExpireAfter[0].MarshalText()
	if err != nil {
		t.Fatalf("Could not get ExpireTime: %v", err)
	}
	return string(res)
}
