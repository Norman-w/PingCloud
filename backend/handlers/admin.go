package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"pingpong/db"
)

// Simple in-memory token store
var (
	tokens   = map[string]*AdminUser{}
	tokenMu  sync.RWMutex
)

type Permissions struct {
	ManageAdmins    bool `json:"manage_admins"`
	ManagePlayers   bool `json:"manage_players"`
	ManageSessions  bool `json:"manage_sessions"`
	ManageFunmatch  bool `json:"manage_funmatch"`
	ScoreMatches    bool `json:"score_matches"`
	EditRatings     bool `json:"edit_ratings"`
	ViewLogs        bool `json:"view_logs"`
	ViewData        bool `json:"view_data"`
	Participate     bool `json:"participate"`
}

type AdminGroup struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Permissions Permissions `json:"permissions"`
}

type AdminUser struct {
	ID          int         `json:"id"`
	Username    string      `json:"username"`
	DisplayName string      `json:"display_name"`
	Role        string      `json:"role"`
	GroupID     int         `json:"group_id"`
	GroupName   string      `json:"group_name"`
	PlayerID    int         `json:"player_id"`
	PlayerName  string      `json:"player_name"`
	CreatedBy   string      `json:"created_by"`
	UpdatedBy   string      `json:"updated_by"`
	Permissions Permissions `json:"permissions"`
	CreatedAt   string      `json:"created_at"`
}

type AdminLog struct {
	ID         int    `json:"id"`
	AdminID    int    `json:"admin_id"`
	AdminName  string `json:"admin_name"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   int    `json:"target_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"created_at"`
}

func hashPassword(pw string) string {
	h := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(h[:])
}

func genToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getAdmin(r *http.Request) *AdminUser {
	cookie, err := r.Cookie("admin_token")
	if err != nil { return nil }
	token := cookie.Value
	// Check memory first
	tokenMu.RLock()
	if u, ok := tokens[token]; ok { tokenMu.RUnlock(); return u }
	tokenMu.RUnlock()
	// Fallback to DB (survives restart)
	var uid int
	var data string
	err = db.DB.QueryRow(`SELECT admin_id, data FROM admin_sessions WHERE token=$1 AND expires_at > NOW()`, token).Scan(&uid, &data)
	if err != nil { return nil }
	var u AdminUser
	json.Unmarshal([]byte(data), &u)
	// Load into memory cache
	tokenMu.Lock()
	tokens[token] = &u
	tokenMu.Unlock()
	return &u
}

func loadPermissions(groupID int) Permissions {
	var p Permissions
	db.DB.QueryRow(`SELECT manage_admins,manage_players,manage_sessions,manage_funmatch,score_matches,edit_ratings,view_logs,view_data,participate FROM admin_groups WHERE id=$1`, groupID).
		Scan(&p.ManageAdmins, &p.ManagePlayers, &p.ManageSessions, &p.ManageFunmatch, &p.ScoreMatches, &p.EditRatings, &p.ViewLogs, &p.ViewData, &p.Participate)
	return p
}

func loadAdminUser(row interface{ Scan(...interface{}) error }) AdminUser {
	var u AdminUser
	var gid int
	var gname string
	row.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Role, &gid, &gname, &u.PlayerID, &u.PlayerName, &u.CreatedBy, &u.UpdatedBy, &u.CreatedAt)
	u.GroupID = gid
	u.GroupName = gname
	u.Permissions = loadPermissions(gid)
	return u
}

func checkPerm(r *http.Request, perm string) *AdminUser {
	admin := getAdmin(r)
	if admin == nil { return nil }
	switch perm {
	case "manage_admins": if admin.Permissions.ManageAdmins { return admin }
	case "manage_players": if admin.Permissions.ManagePlayers { return admin }
	case "manage_sessions": if admin.Permissions.ManageSessions { return admin }
	case "manage_funmatch": if admin.Permissions.ManageFunmatch { return admin }
	case "score_matches": if admin.Permissions.ScoreMatches { return admin }
	case "edit_ratings": if admin.Permissions.EditRatings { return admin }
	case "view_logs": if admin.Permissions.ViewLogs { return admin }
	case "view_data": if admin.Permissions.ViewData { return admin }
	}
	return nil
}

func writeLog(adminID int, action, targetType string, targetID int, detail, ip string) {
	db.DB.Exec(`INSERT INTO admin_logs (admin_id, action, target_type, target_id, detail, ip) VALUES ($1,$2,$3,$4,$5,$6)`,
		adminID, action, targetType, targetID, detail, ip)
}

// POST /api/admin/login
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}

	// Query user with group info
	var u AdminUser
	var gid int
	var gname string
	var hash string
	err := db.DB.QueryRow(`
		SELECT au.id, au.username, COALESCE(au.display_name,''), au.role,
			COALESCE(au.group_id,0), COALESCE(ag.name,''), au.password_hash, au.created_at
		FROM admin_users au
		LEFT JOIN admin_groups ag ON ag.id = au.group_id
		WHERE au.username=$1`, req.Username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Role, &gid, &gname, &hash, &u.CreatedAt)
	if err != nil {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized); return
	}

	if hash != hashPassword(req.Password) {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized); return
	}

	u.GroupID = gid
	u.GroupName = gname
	u.Permissions = loadPermissions(gid)
	u.Role = ""

	token := genToken()
	tokenMu.Lock()
	tokens[token] = &u
	tokenMu.Unlock()

	// Persist to DB for restart survival
	data, _ := json.Marshal(u)
	db.DB.Exec(`INSERT INTO admin_sessions (token, admin_id, data, expires_at) VALUES ($1,$2,$3,NOW()+INTERVAL '24 hours')`,
		token, u.ID, string(data))

	http.SetCookie(w, &http.Cookie{
		Name: "admin_token", Value: token, Path: "/", HttpOnly: true, MaxAge: 86400,
	})
	writeJSON(w, u)
}

// POST /api/admin/logout
func AdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_token")
	if err == nil {
		tokenMu.Lock()
		delete(tokens, cookie.Value)
		tokenMu.Unlock()
		db.DB.Exec(`DELETE FROM admin_sessions WHERE token=$1`, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "admin_token", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, map[string]string{"status": "ok"})
}

// GET /api/admin/me
func AdminMe(w http.ResponseWriter, r *http.Request) {
	admin := getAdmin(r)
	if admin == nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	writeJSON(w, admin)
}

// GET /api/admin/users
func AdminListUsers(w http.ResponseWriter, r *http.Request) {
	admin := getAdmin(r)
	if admin == nil || (admin.Role != "super_admin" && admin.Role != "organizer") {
		http.Error(w, "forbidden", http.StatusForbidden); return
	}

	rows, err := db.DB.Query(`
		SELECT au.id, au.username, COALESCE(au.display_name,''), au.role,
			COALESCE(au.group_id,0), COALESCE(ag.name,''),
			COALESCE(au.player_id,0), COALESCE(p.name,''),
			COALESCE(cu.display_name, cu.username),
			COALESCE(uu.display_name, uu.username),
			au.created_at
		FROM admin_users au
		LEFT JOIN admin_groups ag ON ag.id = au.group_id
		LEFT JOIN players p ON p.id = au.player_id
		LEFT JOIN admin_users cu ON cu.id = au.created_by
		LEFT JOIN admin_users uu ON uu.id = au.updated_by
		ORDER BY au.id`)
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		u := loadAdminUser(rows)
		// Remove role for new schema
		users = append(users, u)
	}
	if users == nil { users = []AdminUser{} }
	writeJSON(w, users)
}

// POST /api/admin/users
func AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "manage_admins")
	if admin == nil { http.Error(w, "forbidden", http.StatusForbidden); return }

	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest); return
	}
	if req.Role == "" { req.Role = "participant" }

	var id int
	err := db.DB.QueryRow(
		`INSERT INTO admin_users (username, password_hash, display_name, role) VALUES ($1,$2,$3,$4) RETURNING id`,
		req.Username, hashPassword(req.Password), req.DisplayName, req.Role,
	).Scan(&id)
	if err != nil {
		http.Error(w, "用户名已存在", http.StatusConflict); return
	}

	writeLog(admin.ID, "添加操作人员", "admin_user", id,
		`{"username":"`+req.Username+`","role":"`+req.Role+`"}`, r.RemoteAddr)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": id})
}

// PUT /api/admin/users/{id}
func AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "manage_admins")
	if admin == nil { http.Error(w, "forbidden", http.StatusForbidden); return }

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	uid, _ := strconv.Atoi(idStr)

	var req struct {
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}

	if req.Password != "" {
		db.DB.Exec(`UPDATE admin_users SET password_hash=$1 WHERE id=$2`, hashPassword(req.Password), uid)
	}
	if req.DisplayName != "" || req.Role != "" {
		db.DB.Exec(`UPDATE admin_users SET display_name=$1, role=$2, updated_at=NOW() WHERE id=$3`, req.DisplayName, req.Role, uid)
	}

	writeLog(admin.ID, "修改操作人员", "admin_user", uid,
		`{"role":"`+req.Role+`","name":"`+req.DisplayName+`"}`, r.RemoteAddr)
	writeJSON(w, map[string]string{"status": "ok"})
}

// DELETE /api/admin/users/{id}
func AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "manage_admins")
	if admin == nil { http.Error(w, "forbidden", http.StatusForbidden); return }

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	uid, _ := strconv.Atoi(idStr)

	// Don't delete self
	if uid == admin.ID { http.Error(w, "不能删除自己", http.StatusBadRequest); return }

	db.DB.Exec(`DELETE FROM admin_users WHERE id=$1`, uid)
	writeLog(admin.ID, "删除操作人员", "admin_user", uid, "", r.RemoteAddr)
	writeJSON(w, map[string]string{"status": "ok"})
}

// GET /api/admin/logs
func AdminGetLogs(w http.ResponseWriter, r *http.Request) {
	admin := getAdmin(r)
	if admin == nil || (admin.Role != "super_admin" && admin.Role != "organizer") {
		http.Error(w, "forbidden", http.StatusForbidden); return
	}

	rows, err := db.DB.Query(`
		SELECT al.id, al.admin_id, COALESCE(au.display_name, au.username), al.action,
			COALESCE(al.target_type,''), COALESCE(al.target_id,0), COALESCE(al.detail,''),
			COALESCE(al.ip,''), al.created_at
		FROM admin_logs al
		LEFT JOIN admin_users au ON au.id = al.admin_id
		ORDER BY al.id DESC LIMIT 100`)
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	defer rows.Close()

	var logs []AdminLog
	for rows.Next() {
		var l AdminLog
		rows.Scan(&l.ID, &l.AdminID, &l.AdminName, &l.Action, &l.TargetType, &l.TargetID, &l.Detail, &l.IP, &l.CreatedAt)
		logs = append(logs, l)
	}
	if logs == nil { logs = []AdminLog{} }
	writeJSON(w, logs)
}

// GET /api/admin/groups
func AdminListGroups(w http.ResponseWriter, r *http.Request) {
	admin := getAdmin(r)
	if admin == nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	rows, err := db.DB.Query(`SELECT id, name, COALESCE(description,''), manage_admins, manage_players, manage_sessions, manage_funmatch, score_matches, edit_ratings, view_logs, view_data, participate FROM admin_groups ORDER BY id`)
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	defer rows.Close()

	var groups []AdminGroup
	for rows.Next() {
		var g AdminGroup
		rows.Scan(&g.ID, &g.Name, &g.Description,
			&g.Permissions.ManageAdmins, &g.Permissions.ManagePlayers, &g.Permissions.ManageSessions,
			&g.Permissions.ManageFunmatch, &g.Permissions.ScoreMatches, &g.Permissions.EditRatings,
			&g.Permissions.ViewLogs, &g.Permissions.ViewData, &g.Permissions.Participate)
		groups = append(groups, g)
	}
	if groups == nil { groups = []AdminGroup{} }
	writeJSON(w, groups)
}

// POST /api/admin/rating
func AdminAdjustRating(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "edit_ratings")
	if admin == nil { http.Error(w, "forbidden: only super_admin can edit ratings", http.StatusForbidden); return }

	var req struct {
		PlayerID int    `json:"player_id"`
		NewRating int   `json:"new_rating"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}

	var oldRating int
	db.DB.QueryRow(`SELECT current_rating FROM players WHERE id=$1`, req.PlayerID).Scan(&oldRating)
	db.DB.Exec(`UPDATE players SET current_rating=$1, initial_rating=CASE WHEN initial_rating=0 THEN $1 ELSE initial_rating END WHERE id=$2`, req.NewRating, req.PlayerID)

	writeLog(admin.ID, "修改球员积分", "player", req.PlayerID,
		`{"old":`+strconv.Itoa(oldRating)+`,"new":`+strconv.Itoa(req.NewRating)+`,"reason":"`+req.Reason+`"}`, r.RemoteAddr)
	writeJSON(w, map[string]interface{}{"status":"ok","old_rating":oldRating,"new_rating":req.NewRating})
}

// Update CreateUser to use group_id
func AdminCreateUserV2(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "manage_admins")
	if admin == nil { http.Error(w, "forbidden", http.StatusForbidden); return }

	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		GroupID     int    `json:"group_id"`
		PlayerID    int    `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest); return
	}
	if req.GroupID == 0 { req.GroupID = 4 } // default participant

	var id int
	err := db.DB.QueryRow(
		`INSERT INTO admin_users (username, password_hash, display_name, group_id, player_id, created_by) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		req.Username, hashPassword(req.Password), req.DisplayName, req.GroupID, req.PlayerID, admin.ID,
	).Scan(&id)
	if err != nil { http.Error(w, "用户名已存在", http.StatusConflict); return }

	writeLog(admin.ID, "添加操作人员", "admin_user", id,
		`{"username":"`+req.Username+`","group_id":`+strconv.Itoa(req.GroupID)+`}`, r.RemoteAddr)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": id})
}

// Update UpdateUser to use group_id
func AdminUpdateUserV2(w http.ResponseWriter, r *http.Request) {
	admin := checkPerm(r, "manage_admins")
	if admin == nil { http.Error(w, "forbidden", http.StatusForbidden); return }

	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	uid, _ := strconv.Atoi(idStr)

	var req struct {
		DisplayName string `json:"display_name"`
		GroupID     int    `json:"group_id"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest); return
	}

	if req.Password != "" {
		db.DB.Exec(`UPDATE admin_users SET password_hash=$1 WHERE id=$2`, hashPassword(req.Password), uid)
	}
	if req.DisplayName != "" || req.GroupID > 0 {
		db.DB.Exec(`UPDATE admin_users SET display_name=$1, group_id=$2, updated_by=$3, updated_at=NOW() WHERE id=$4`, req.DisplayName, req.GroupID, admin.ID, uid)
	}

	writeLog(admin.ID, "修改操作人员", "admin_user", uid,
		`{"group_id":`+strconv.Itoa(req.GroupID)+`}`, r.RemoteAddr)
	writeJSON(w, map[string]string{"status": "ok"})
}
