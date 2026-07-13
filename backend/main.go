package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pingpong/db"
	"pingpong/handlers"
)

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func logAccess(r *http.Request) {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" { ip = r.Header.Get("X-Forwarded-For") }
	if ip == "" { ip = r.RemoteAddr }
	// strip port from RemoteAddr if present
	if idx := strings.LastIndex(ip, ":"); idx > 0 && !strings.Contains(ip, "[") {
		ip = ip[:idx]
	}

	pid := 0
	if c, err := r.Cookie("ping_id"); err == nil && c.Value != "" {
		fmt.Sscanf(c.Value, "%d:", &pid)
	}
	db.DB.Exec(`INSERT INTO access_logs (ip, path, method, user_agent, referer, player_id) VALUES ($1,$2,$3,$4,$5,$6)`,
		ip, r.URL.Path, r.Method, r.UserAgent(), r.Referer(), pid)
}

func loadDotEnv(paths ...string) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.Trim(strings.TrimSpace(value), `"'`)
			if key != "" && os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
		return
	}
}

func main() {
	loadDotEnv(".env", filepath.Join("backend", ".env"))
	db.Init()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/players/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			idStr := strings.TrimPrefix(r.URL.Path, "/api/players/")
			if idStr != "" {
				handlers.GetPlayer(w, r)
				return
			}
			handlers.GetPlayers(w, r)
		case http.MethodPost:
			handlers.CreatePlayer(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/players", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.GetPlayers(w, r)
		case http.MethodPost:
			handlers.CreatePlayer(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.GetMatches(w, r)
		case http.MethodPost:
			handlers.CreateMatch(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Head-to-head
	mux.HandleFunc("/api/headtohead", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handlers.GetHeadToHead(w, r)
	})

	mux.HandleFunc("/api/rankings", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handlers.GetRankings(w, r)
	})

	// Session routes (prefix: /api/sessions/{id}...)
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")

		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/matches/") && strings.HasSuffix(path, "/forfeit") {
				handlers.ForfeitMatch(w, r)
				return
			}
			if strings.Contains(path, "/matches/") {
				handlers.ScoreSessionMatch(w, r)
				return
			}
			if strings.HasSuffix(path, "/complete") || strings.HasSuffix(path, "complete") {
				handlers.CompleteSession(w, r)
				return
			}
			if strings.HasSuffix(path, "/players") || strings.Contains(path, "/players") {
				handlers.AddPlayerToSession(w, r)
				return
			}
			handlers.CreateSession(w, r)
		case http.MethodDelete:
			if strings.Contains(path, "/matches/") {
				handlers.DeleteMatch(w, r)
				return
			}
			if path != "" && !strings.Contains(path, "/") {
				handlers.DeleteSession(w, r)
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodPut:
			if path != "" && !strings.Contains(path, "/") {
				handlers.UpdateSession(w, r)
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodGet:
			if path == "" {
				handlers.GetSessions(w, r)
				return
			}
			handlers.GetSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.GetSessions(w, r)
		case http.MethodPost:
			handlers.CreateSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Fun session routes (prefix: /api/fun-sessions/{id}...)
	mux.HandleFunc("/api/fun-sessions/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")

		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/draw-card/") {
				handlers.DrawFunCard(w, r)
				return
			}
			if strings.Contains(path, "/matches/") && strings.HasSuffix(path, "/forfeit") {
				handlers.ForfeitFunMatch(w, r)
				return
			}
			if strings.Contains(path, "/matches/") {
				handlers.ScoreFunMatch(w, r)
				return
			}
			if strings.HasSuffix(path, "/complete") {
				handlers.CompleteFunSession(w, r)
				return
			}
			handlers.CreateFunSession(w, r)
		case http.MethodDelete:
			handlers.DeleteFunSession(w, r)
		case http.MethodPut:
			handlers.UpdateFunSession(w, r)
		case http.MethodGet:
			if path == "" {
				handlers.GetFunSessions(w, r)
				return
			}
			handlers.GetFunSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/fun-sessions", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.GetFunSessions(w, r)
		case http.MethodPost:
			handlers.CreateFunSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin routes
	mux.HandleFunc("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AdminLogin(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AdminLogout(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/me", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			handlers.AdminMe(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/logs", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			handlers.AdminGetLogs(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/users/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodPut:
			handlers.AdminUpdateUserV2(w, r)
		case http.MethodDelete:
			handlers.AdminDeleteUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handlers.AdminListUsers(w, r)
		case http.MethodPost:
			handlers.AdminCreateUserV2(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/admin/groups", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			handlers.AdminListGroups(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/access-logs", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.AdminGetAccessLogs(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/admin/rating", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AdminAdjustRating(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Team battle routes
	mux.HandleFunc("/api/team-battles/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		path := strings.TrimPrefix(r.URL.Path, "/api/team-battles/")
		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/matches/") && strings.HasSuffix(path, "/forfeit") { handlers.ForfeitTeamBattleMatch(w, r); return }
			if strings.Contains(path, "/matches/") { handlers.ScoreTeamBattleMatch(w, r); return }
			if path == "complete" || strings.HasSuffix(path, "/complete") { handlers.CompleteTeamBattle(w, r); return }
			handlers.CreateTeamBattle(w, r)
		case http.MethodDelete: handlers.DeleteTeamBattle(w, r)
		case http.MethodGet:
			if path == "" { handlers.GetTeamBattles(w, r); return }
			handlers.GetTeamBattle(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/team-battles", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetTeamBattles(w, r)
		case http.MethodPost: handlers.CreateTeamBattle(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Auth routes
	mux.HandleFunc("/api/auth/send-code", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AuthSendCode(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AuthVerify(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			handlers.AuthMe(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/auth/check-name", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AuthCheckName(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/auth/complete", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			handlers.AuthComplete(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Skills library
	mux.HandleFunc("/api/skills", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.GetSkills(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Training logs
	mux.HandleFunc("/api/training-logs/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		path := strings.TrimPrefix(r.URL.Path, "/api/training-logs/")
		switch r.Method {
		case http.MethodGet:
			if path == "" { handlers.GetTrainingLogs(w, r); return }
			handlers.GetTrainingLog(w, r)
		case http.MethodPost:
			if path == "" { handlers.CreateTrainingLog(w, r); return }
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodPut:
			if path != "" { handlers.UpdateTrainingLog(w, r); return }
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		case http.MethodDelete:
			if path != "" { handlers.DeleteTrainingLog(w, r); return }
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/training-logs", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetTrainingLogs(w, r)
		case http.MethodPost: handlers.CreateTrainingLog(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Training stats
	mux.HandleFunc("/api/training-stats", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.GetTrainingStats(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Skill mastery
	mux.HandleFunc("/api/skill-mastery/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodPut { handlers.UpdateSkillMastery(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/skill-mastery", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.GetSkillMastery(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Locations
	mux.HandleFunc("/api/locations/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetLocation(w, r)
		case http.MethodPut: handlers.UpdateLocation(w, r)
		case http.MethodDelete: handlers.DeleteLocation(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/locations", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		switch r.Method {
		case http.MethodGet: handlers.GetLocations(w, r)
		case http.MethodPost: handlers.CreateLocation(w, r)
		default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Skill training
	mux.HandleFunc("/api/skill-train/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.GetSkillTrainHistory(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/skill-train", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodPost { handlers.CreateSkillTraining(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Skill goals
	mux.HandleFunc("/api/skill-goals/", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusOK); return }
		if r.Method == http.MethodGet { handlers.GetSkillGoals(w, r); return }
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap mux with access logging
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip OPTIONS and static file requests
		if r.Method != http.MethodOptions && !strings.HasPrefix(r.URL.Path, "/assets/") && r.URL.Path != "/favicon.svg" && r.URL.Path != "/icons.svg" {
			go logAccess(r)
		}
		mux.ServeHTTP(w, r)
	})

	port := ":8090"
	log.Printf("PingPong server starting on %s", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
