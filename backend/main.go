package main

import (
	"log"
	"net/http"
	"strings"

	"pingpong/db"
	"pingpong/handlers"
)

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	db.Init()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/players/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet:
			idStr := strings.TrimPrefix(r.URL.Path, "/api/players/")
			if idStr != "" { handlers.GetPlayer(w, r); return }
			handlers.GetPlayers(w, r)
		case http.MethodPost: handlers.CreatePlayer(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/players", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetPlayers(w, r)
		case http.MethodPost: handlers.CreatePlayer(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetMatches(w, r)
		case http.MethodPost: handlers.CreateMatch(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Head-to-head
	mux.HandleFunc("/api/headtohead", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		handlers.GetHeadToHead(w, r)
	})

	mux.HandleFunc("/api/rankings", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		handlers.GetRankings(w, r)
	})

	// Session routes (prefix: /api/sessions/{id}...)
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")

		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/matches/") && strings.HasSuffix(path, "/forfeit") { handlers.ForfeitMatch(w, r); return }
			if strings.Contains(path, "/matches/") { handlers.ScoreSessionMatch(w, r); return }
			if strings.HasSuffix(path, "/complete") || strings.HasSuffix(path, "complete") { handlers.CompleteSession(w, r); return }
			if strings.HasSuffix(path, "/players") || strings.Contains(path, "/players") { handlers.AddPlayerToSession(w, r); return }
			handlers.CreateSession(w, r)
		case http.MethodDelete:
			if strings.Contains(path, "/matches/") { handlers.DeleteMatch(w, r); return }
			if path != "" && !strings.Contains(path, "/") { handlers.DeleteSession(w, r); return }
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodPut:
			if path != "" && !strings.Contains(path, "/") { handlers.UpdateSession(w, r); return }
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodGet:
			if path == "" { handlers.GetSessions(w, r); return }
			handlers.GetSession(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetSessions(w, r)
		case http.MethodPost: handlers.CreateSession(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Fun session routes (prefix: /api/fun-sessions/{id}...)
	mux.HandleFunc("/api/fun-sessions/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")

		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/draw-card/") { handlers.DrawFunCard(w, r); return }
			if strings.Contains(path, "/matches/") { handlers.ScoreFunMatch(w, r); return }
			if strings.HasSuffix(path, "/complete") { handlers.CompleteFunSession(w, r); return }
			handlers.CreateFunSession(w, r)
		case http.MethodDelete:
			handlers.DeleteFunSession(w, r)
		case http.MethodPut:
			handlers.UpdateFunSession(w, r)
		case http.MethodGet:
			if path == "" { handlers.GetFunSessions(w, r); return }
			handlers.GetFunSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/fun-sessions", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetFunSessions(w, r)
		case http.MethodPost: handlers.CreateFunSession(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := ":8090"
	log.Printf("PingPong server starting on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
