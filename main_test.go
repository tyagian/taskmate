package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	// Create temporary config
	config := &Config{
		Port:         "8080",
		PasswordHash: hashString("testpassword"),
		TokenHashes:  []string{},
	}

	// Create temporary data file
	tmpFile := "test_tasks.json"
	server := NewServer(config, tmpFile)

	cleanup := func() {
		os.Remove(tmpFile)
	}

	return server, cleanup
}

func TestHashString(t *testing.T) {
	input := "randomforest"
	expected := "ea424017c57b0d0b2f262edd821dca2dc3cfcbb47e296a9007415af86bbc6ac1"
	result := hashString(input)

	if result != expected {
		t.Errorf("hashString(%s) = %s; want %s", input, result, expected)
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	if len(token) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("generateToken() length = %d; want 64", len(token))
	}
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health check status = %d; want %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Health check body = %s; want OK", w.Body.String())
	}
}

func TestGetTasksNoAuth(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	server.handleGetTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /tasks status = %d; want %d", w.Code, http.StatusOK)
	}

	var tasks []Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Initial tasks count = %d; want 0", len(tasks))
	}
}

func TestGenerateTokenWithValidPassword(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]string{"password": "testpassword"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleGenerateToken(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Generate token status = %d; want %d", w.Code, http.StatusCreated)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["token"] == "" {
		t.Error("Expected token in response, got empty string")
	}
}

func TestGenerateTokenWithInvalidPassword(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]string{"password": "wrongpassword"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleGenerateToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Generate token with wrong password status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestCreateTaskWithoutToken(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]string{
		"title":       "Test Task",
		"description": "Test Description",
		"priority":    "high",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.tokenAuthMiddleware(server.handleCreateTask)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Create task without token status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestTaskStoreOperations(t *testing.T) {
	tmpFile := "test_store.json"
	defer os.Remove(tmpFile)

	store := NewTaskStore(tmpFile)

	// Test Add
	task := store.Add("Test Task", "Description", "2024-12-31", "high")
	if task.ID != 1 {
		t.Errorf("First task ID = %d; want 1", task.ID)
	}
	if task.Status != "pending" {
		t.Errorf("New task status = %s; want pending", task.Status)
	}

	// Test Get
	retrieved, exists := store.Get(1)
	if !exists {
		t.Error("Task should exist")
	}
	if retrieved.Title != "Test Task" {
		t.Errorf("Retrieved task title = %s; want Test Task", retrieved.Title)
	}

	// Test GetAll
	all := store.GetAll()
	if len(all) != 1 {
		t.Errorf("GetAll count = %d; want 1", len(all))
	}

	// Test Update
	updated, exists := store.Update(1, "Updated Task", "New Description", "2024-12-31", "low", "completed")
	if !exists {
		t.Error("Task should exist for update")
	}
	if updated.Title != "Updated Task" {
		t.Errorf("Updated task title = %s; want Updated Task", updated.Title)
	}
	if updated.Status != "completed" {
		t.Errorf("Updated task status = %s; want completed", updated.Status)
	}

	// Test GetPending
	pending := store.GetPending()
	if len(pending) != 0 {
		t.Errorf("Pending tasks count = %d; want 0 (task is completed)", len(pending))
	}

	// Test Delete
	deleted := store.Delete(1)
	if !deleted {
		t.Error("Task should be deleted")
	}

	_, exists = store.Get(1)
	if exists {
		t.Error("Deleted task should not exist")
	}
}
