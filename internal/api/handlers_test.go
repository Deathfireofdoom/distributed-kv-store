package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/kvstore"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/models"
)

func TestPutHandler(t *testing.T) {
	_ = kvstore.NewStore()
	handler := http.HandlerFunc(PutHandler)

	tests := []struct {
		name           string
		input          models.PutRequest
		expectedStatus int
		expectedBody   models.PutResponse
	}{
		{
			name:           "Valid PUT",
			input:          models.PutRequest{Key: "key1", Value: "value1"},
			expectedStatus: http.StatusOK,
			expectedBody:   models.PutResponse{Success: true},
		},
		{
			name:           "Invalid PUT - Empty Key",
			input:          models.PutRequest{Key: "", Value: "value1"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.PutResponse{Success: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest("POST", "/store", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if rr.Code == http.StatusOK {
				var resp models.PutResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("could not decode response: %v", err)
				}
				if resp != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						resp, tt.expectedBody)
				}
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	store := kvstore.NewStore()
	handler := http.HandlerFunc(GetHandler)

	store.Put("key1", "value1")

	tests := []struct {
		name           string
		queryParam     string
		expectedStatus int
		expectedBody   models.GetResponse
	}{
		{
			name:           "Valid GET",
			queryParam:     "key1",
			expectedStatus: http.StatusOK,
			expectedBody:   models.GetResponse{Value: "value1", Found: true},
		},
		{
			name:           "Invalid GET - Missing Key",
			queryParam:     "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.GetResponse{},
		},
		{
			name:           "GET - Key Not Found",
			queryParam:     "key2",
			expectedStatus: http.StatusNotFound,
			expectedBody:   models.GetResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/store?key="+tt.queryParam, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if rr.Code == http.StatusOK {
				var resp models.GetResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("could not decode response: %v", err)
				}
				if resp != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						resp, tt.expectedBody)
				}
			}
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	store := kvstore.NewStore()
	handler := http.HandlerFunc(DeleteHandler)

	store.Put("key1", "value1")

	tests := []struct {
		name           string
		input          models.DeleteRequest
		expectedStatus int
		expectedBody   models.DeleteResponse
	}{
		{
			name:           "Valid DELETE",
			input:          models.DeleteRequest{Key: "key1"},
			expectedStatus: http.StatusOK,
			expectedBody:   models.DeleteResponse{Success: true},
		},
		{
			name:           "Invalid DELETE - Empty Key",
			input:          models.DeleteRequest{Key: ""},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.DeleteResponse{Success: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest("DELETE", "/store", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if rr.Code == http.StatusOK {
				var resp models.DeleteResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("could not decode response: %v", err)
				}
				if resp != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						resp, tt.expectedBody)
				}
			}
		})
	}
}
