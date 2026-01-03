package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strconv"

	"github.com/amterp/council/internal/errors"
	"github.com/amterp/council/internal/session"
)

// Server handles HTTP requests for the Council web interface
type Server struct {
	sessionID string
	port      int
	mux       *http.ServeMux
}

// NewServer creates a new web server for the given session
func NewServer(sessionID string, port int) *Server {
	s := &Server{
		sessionID: sessionID,
		port:      port,
		mux:       http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/post", s.handlePost)
	s.mux.HandleFunc("/api/participants", s.handleParticipants)

	// Serve embedded frontend with SPA fallback
	distFS, err := fs.Sub(WebAssets, "dist")
	if err != nil {
		// If dist doesn't exist yet (during development), serve a placeholder
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Council Watch</title></head>
<body>
<h1>Council Watch</h1>
<p>Frontend not built yet. Run <code>cd web && npm run build</code> to build the frontend.</p>
</body>
</html>`))
		})
		return
	}

	fileServer := http.FileServer(http.FS(distFS))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists in embedded FS
		f, err := distFS.Open(path[1:]) // Remove leading /
		if err != nil {
			// File not found, serve index.html for SPA routing
			r.URL.Path = "/index.html"
		} else {
			f.Close()
		}

		fileServer.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server (blocking)
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s.mux)
}

// handleStatus implements GET /api/status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		writeJSONError(w, "session parameter required", http.StatusBadRequest)
		return
	}

	afterStr := r.URL.Query().Get("after")
	afterN := 0
	if afterStr != "" {
		var err error
		afterN, err = strconv.Atoi(afterStr)
		if err != nil {
			writeJSONError(w, "invalid after parameter", http.StatusBadRequest)
			return
		}
	}

	sess, err := session.LoadSession(sessionID)
	if err != nil {
		if _, ok := err.(*errors.SessionNotFoundError); ok {
			writeJSONError(w, "session not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert events to API format, filtering by after parameter
	apiEvents := make([]APIEvent, 0)
	for i, event := range sess.Events {
		eventNum := i + 1 // 1-indexed
		if eventNum <= afterN {
			continue
		}
		apiEvents = append(apiEvents, convertToAPIEvent(event, eventNum))
	}

	participants := sess.ActiveParticipants()
	sort.Strings(participants)

	resp := StatusResponse{
		SessionID:    sessionID,
		Participants: participants,
		EventCount:   sess.EventCount(),
		Events:       apiEvents,
	}

	writeJSON(w, resp)
}

// handlePost implements POST /api/post
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Session == "" {
		writeJSONError(w, "session field required", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		writeJSONError(w, "content field required", http.StatusBadRequest)
		return
	}

	// Use "Moderator" as the participant name (per spec)
	next := ""
	if req.Next != nil {
		next = *req.Next
	}

	eventNum, err := session.PostMessage(req.Session, "Moderator", req.Content, next, req.After)
	if err != nil {
		switch err.(type) {
		case *errors.SessionNotFoundError:
			writeJSONError(w, "session not found", http.StatusNotFound)
		case *errors.StaleStateError:
			writeJSONError(w, err.Error(), http.StatusConflict)
		case *errors.InvalidNextParticipantError:
			writeJSONError(w, err.Error(), http.StatusBadRequest)
		default:
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, PostResponse{EventNumber: eventNum})
}

// handleParticipants implements GET /api/participants
func (s *Server) handleParticipants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		writeJSONError(w, "session parameter required", http.StatusBadRequest)
		return
	}

	sess, err := session.LoadSession(sessionID)
	if err != nil {
		if _, ok := err.(*errors.SessionNotFoundError); ok {
			writeJSONError(w, "session not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	participants := sess.ActiveParticipants()
	sort.Strings(participants)

	writeJSON(w, ParticipantsResponse{Participants: participants})
}

// convertToAPIEvent converts an internal Event to an APIEvent
func convertToAPIEvent(event session.Event, number int) APIEvent {
	api := APIEvent{
		Number:          number,
		Type:            string(event.GetType()),
		TimestampMillis: event.GetTimestamp(),
	}

	switch e := event.(type) {
	case *session.SessionCreatedEvent:
		api.ID = e.ID
	case *session.JoinedEvent:
		api.Participant = e.Participant
	case *session.LeftEvent:
		api.Participant = e.Participant
	case *session.MessageEvent:
		api.Participant = e.Participant
		api.Content = e.Content
		api.Next = e.Next
	}

	return api
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// FindAvailablePort scans for an available port starting from startPort
func FindAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	return startPort // Fallback
}

// OpenBrowser opens the given URL in the default browser
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
