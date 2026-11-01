package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task), // to avoid 'writing to nil-map' panic later!
	}
}

func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	task := Task{
		ID:        strconv.FormatUint(r.seq, 10),
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[task.ID] = task

	return task, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]

	return task, ok
}

func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resultingTasks := make([]Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		if task.Done == done {
			resultingTasks = append(resultingTasks, task)
		}
	}

	return resultingTasks
}

func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()

		r.tasks[id] = task // update actual elem ('task' here is just a copy)
	}

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path // 'GET /tasks/42?done=true' -> '/tasks/42'
	specificTaskPrefix := "/tasks/"

	switch {
	case path == "/tasks":
		switch r.Method {
		case http.MethodPost: // == "POST" (const; good practice to avoid misspellings)
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			allowedMethods := "GET, POST"
			methodNotAllowed(w, allowedMethods)
		}
	case strings.HasPrefix(path, specificTaskPrefix):
		taskID := path[len(specificTaskPrefix):]

		wrongID := taskID == "" || strings.Contains(taskID, "/")
		if wrongID {
			writeJsonError(w, http.StatusNotFound, "not found")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, taskID)
		case http.MethodPatch:
			h.handlePatch(w, r, taskID)
		default:
			allowedMethods := "GET, PATCH"
			methodNotAllowed(w, allowedMethods)
		}
	default:
		writeJsonError(w, http.StatusNotFound, "not found")
	}
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var decodedBody struct {
		Title *string `json:"title"` // ptr so that {"title":""} is different from {"title":null} or {}
	}
	if err := decodeStrictJSON(r.Body, &decodedBody); err != nil || decodedBody.Title == nil {
		writeJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	taskTitle := *decodedBody.Title
	task, err := h.repo.Create(taskTitle)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			writeJsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJsonError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		writeJsonError(w, http.StatusNotFound, ErrNotFound.Error())
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done := false
	// NOTE: "?done=true&done=false".Query() -> {"done" : ["true", "true", "false"]}
	if vals, present := r.URL.Query()["done"]; present {
		if len(vals) != 1 {
			writeJsonError(w, http.StatusBadRequest, "invalid done parameter")
			return
		}
		switch vals[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			writeJsonError(w, http.StatusBadRequest, "invalid done parameter")
			return
		}
	}

	tasks := slices.Clone(h.repo.List(done)) // make sure we don't modify original slice (just in case)
	if tasks == nil {
		tasks = []Task{}
	}

	slices.SortFunc(tasks, compareTasksNewestFirst)

	writeJSON(w, http.StatusOK, tasks)
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var decodedBody struct {
		Done *bool `json:"done"`
	}
	if err := decodeStrictJSON(r.Body, &decodedBody); err != nil || decodedBody.Done == nil {
		writeJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.repo.SetDone(id, *decodedBody.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJsonError(w, http.StatusNotFound, err.Error())
			return
		}

		writeJsonError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// converts bytes into struct ONLY IF format is fine (instead of json.Unmarshal)
// i.e. single JSON-object with no extra fields
func decodeStrictJSON(r io.Reader, resultPtr any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields() // return err if unknown field is provided, not just ignore it

	if err := dec.Decode(resultPtr); err != nil {
		return err
	}

	// check if second JSON object or anything else is present in the bytes ({"title":"a"}{"title":"b"})
	if _, err := dec.Token(); err != io.EOF {
		// something is indeed present

		if err == nil { // smt else was successfully read -> custom error
			return errors.New("unexpected data after JSON object")
		}

		return err // smt else couldn't be read -> library err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, jsonLikeGoObj any) {
	bodyBytes, err := json.Marshal(jsonLikeGoObj)

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(bodyBytes)
}

// helpers
func writeJsonError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter, allowedMethods string) {
	w.Header().Set("Allow", allowedMethods) // 405 requires http header 'Allow: GET, POST'
	writeJsonError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func compareTasksNewestFirst(a, b Task) int {
	if c := b.UpdatedAt.Compare(a.UpdatedAt); c != 0 {
		return c
	}

	return cmp.Compare(a.ID, b.ID)
}
