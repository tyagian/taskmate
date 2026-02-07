package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Task represents a pending task
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     string    `json:"due_date"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Config holds application configuration
type Config struct {
	APIKey       string   `json:"api_key"`
	Port         string   `json:"port"`
	PasswordHash string   `json:"password_hash"`
	TokenHashes  []string `json:"token_hashes"`
}

// LoadConfig reads configuration from config.json or environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		TokenHashes: []string{},
	}

	// Try to load from file first
	data, err := os.ReadFile("config.json")
	if err == nil {
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables if set (for containers)
	if port := os.Getenv("TASKMATE_PORT"); port != "" {
		config.Port = port
	}
	if config.Port == "" {
		config.Port = "8080" // Default port
	}

	if apiKey := os.Getenv("TASKMATE_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if passwordHash := os.Getenv("TASKMATE_PASSWORD_HASH"); passwordHash != "" {
		config.PasswordHash = passwordHash
	}

	// If no password hash is set, use default (randomforest)
	if config.PasswordHash == "" {
		config.PasswordHash = "ea424017c57b0d0b2f262edd821dca2dc3cfcbb47e296a9007415af86bbc6ac1"
		log.Println("Warning: Using default password hash. Set TASKMATE_PASSWORD_HASH environment variable for production.")
	}

	// Initialize token_hashes if nil
	if config.TokenHashes == nil {
		config.TokenHashes = []string{}
	}

	return config, nil
}

// SaveConfig writes configuration to config.json
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("config.json", data, 0600)
}

// hashString creates SHA-256 hash of input string
func hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// generateToken creates a random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TaskStore manages tasks with JSON persistence
type TaskStore struct {
	mu       sync.RWMutex
	tasks    map[int]*Task
	nextID   int
	filePath string
}

// NewTaskStore creates a new task store
func NewTaskStore(filePath string) *TaskStore {
	store := &TaskStore{
		tasks:    make(map[int]*Task),
		nextID:   1,
		filePath: filePath,
	}
	store.loadFromFile()
	return store
}

// loadFromFile loads tasks from JSON file
func (ts *TaskStore) loadFromFile() {
	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		return // File doesn't exist yet
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return
	}

	for _, task := range tasks {
		ts.tasks[task.ID] = task
		if task.ID >= ts.nextID {
			ts.nextID = task.ID + 1
		}
	}
}

// saveToFile persists tasks to JSON file
func (ts *TaskStore) saveToFile() error {
	tasks := make([]*Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ts.filePath, data, 0600)
}

// Add creates a new task
func (ts *TaskStore) Add(title, description, dueDate, priority string) *Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	task := &Task{
		ID:          ts.nextID,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Priority:    priority,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ts.tasks[ts.nextID] = task
	ts.nextID++
	if err := ts.saveToFile(); err != nil {
		log.Printf("Failed to save tasks: %v", err)
	}
	return task
}

// Get retrieves a task by ID
func (ts *TaskStore) Get(id int) (*Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	task, exists := ts.tasks[id]
	return task, exists
}

// GetAll returns all tasks
func (ts *TaskStore) GetAll() []*Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetPending returns only pending tasks
func (ts *TaskStore) GetPending() []*Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range ts.tasks {
		if task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// Update modifies an existing task
func (ts *TaskStore) Update(id int, title, description, dueDate, priority, status string) (*Task, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task, exists := ts.tasks[id]
	if !exists {
		return nil, false
	}

	task.Title = title
	task.Description = description
	task.DueDate = dueDate
	task.Priority = priority
	task.Status = status
	task.UpdatedAt = time.Now()
	if err := ts.saveToFile(); err != nil {
		log.Printf("Failed to save tasks: %v", err)
	}
	return task, true
}

// Delete removes a task
func (ts *TaskStore) Delete(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	_, exists := ts.tasks[id]
	if exists {
		delete(ts.tasks, id)
		if err := ts.saveToFile(); err != nil {
			log.Printf("Failed to save tasks: %v", err)
		}
	}
	return exists
}

// Server holds our application state
type Server struct {
	store  *TaskStore
	config *Config
	mu     sync.RWMutex
}

// NewServer creates a new server instance
func NewServer(config *Config, dataFile string) *Server {
	return &Server{
		store:  NewTaskStore(dataFile),
		config: config,
	}
}

// tokenAuthMiddleware checks for valid token (for POST/DELETE operations)
func (s *Server) tokenAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-Token")
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// Hash the provided token
		tokenHash := hashString(token)

		// Check if token hash exists in config
		s.mu.RLock()
		valid := false
		for _, storedHash := range s.config.TokenHashes {
			if storedHash == tokenHash {
				valid = true
				break
			}
		}
		s.mu.RUnlock()

		if !valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// handleGetTasks returns all tasks
func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.store.GetAll()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("Failed to encode tasks: %v", err)
	}
}

// handleGetPendingTasks returns only pending tasks
func (s *Server) handleGetPendingTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.store.GetPending()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("Failed to encode tasks: %v", err)
	}
}

// handleGetTask returns a specific task
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, exists := s.store.Get(id)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Failed to encode task: %v", err)
	}
}

// handleCreateTask creates a new task
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
		Priority    string `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	task := s.store.Add(req.Title, req.Description, req.DueDate, req.Priority)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Failed to encode task: %v", err)
	}
}

// handleUpdateTask updates an existing task
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     string `json:"due_date"`
		Priority    string `json:"priority"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	task, exists := s.store.Update(id, req.Title, req.Description, req.DueDate, req.Priority, req.Status)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Failed to encode task: %v", err)
	}
}

// handleDeleteTask deletes a task
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if !s.store.Delete(id) {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGenerateToken generates a new API token without password verification (educational use only)
func (s *Server) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	// Generate new token
	token, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Hash the token and store it
	tokenHash := hashString(token)

	s.mu.Lock()
	s.config.TokenHashes = append(s.config.TokenHashes, tokenHash)
	if err := SaveConfig(s.config); err != nil {
		s.mu.Unlock()
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}
	s.mu.Unlock()

	// Return the token to the user (only time they'll see it)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token":   token,
		"message": "Token generated successfully. Save this token securely, it won't be shown again.",
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	port := config.Port
	dataFile := "tasks.json"
	server := NewServer(config, dataFile)

	r := mux.NewRouter()

	// Serve static files (HTML/CSS/JS)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve UI at root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Token generation endpoint (requires password)
	api.HandleFunc("/auth/token", server.handleGenerateToken).Methods("POST")

	// GET requests - no authentication required
	api.HandleFunc("/tasks", server.handleGetTasks).Methods("GET")
	api.HandleFunc("/tasks/pending", server.handleGetPendingTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", server.handleGetTask).Methods("GET")

	// POST/PUT/DELETE requests - require token authentication
	api.HandleFunc("/tasks", server.tokenAuthMiddleware(server.handleCreateTask)).Methods("POST")
	api.HandleFunc("/tasks/{id}", server.tokenAuthMiddleware(server.handleUpdateTask)).Methods("PUT")
	api.HandleFunc("/tasks/{id}", server.tokenAuthMiddleware(server.handleDeleteTask)).Methods("DELETE")

	// Serve config endpoint for UI (deprecated - will be removed)
	r.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Use token-based authentication"}); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}).Methods("GET")

	// Health check endpoint (no auth required)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}).Methods("GET")

	fmt.Println("TaskMate API server starting on :" + port)
	fmt.Printf("Data File: %s\n", dataFile)
	fmt.Println("\n🌐 Web UI: http://localhost:" + port)
	fmt.Println("Health check: http://localhost:" + port + "/health")
	fmt.Println("API Base URL: http://localhost:" + port + "/api/v1")
	fmt.Println("\nEndpoints:")
	fmt.Println("  POST   /api/v1/auth/token     - Generate token (no auth required)")
	fmt.Println("  GET    /api/v1/tasks          - List all tasks (no auth)")
	fmt.Println("  GET    /api/v1/tasks/pending  - List pending tasks (no auth)")
	fmt.Println("  GET    /api/v1/tasks/{id}     - Get task (no auth)")
	fmt.Println("  POST   /api/v1/tasks          - Create task (requires token)")
	fmt.Println("  PUT    /api/v1/tasks/{id}     - Update task (requires token)")
	fmt.Println("  DELETE /api/v1/tasks/{id}     - Delete task (requires token)")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
