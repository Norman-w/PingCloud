package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"pingpong/db"
)

// In-memory code store: phone → {code, expires, playerId}
var (
	codeStore = map[string]struct {
		code     string
		expires  time.Time
		attempts int
	}{}
	codeStoreMu sync.Mutex

	// IP rate limiter: track SMS sends per IP
	ipRateMap   = map[string][]time.Time{}
	ipRateMapMu sync.Mutex
)

// IP rate limit: max 5 SMS per hour per IP
func checkIPRate(ip string) bool {
	ipRateMapMu.Lock()
	defer ipRateMapMu.Unlock()
	now := time.Now()
	// Clean old entries
	var recent []time.Time
	for _, t := range ipRateMap[ip] {
		if now.Sub(t) < time.Hour {
			recent = append(recent, t)
		}
	}
	if len(recent) >= 5 {
		ipRateMap[ip] = recent
		return false // rate limited
	}
	recent = append(recent, now)
	ipRateMap[ip] = recent
	return true
}

func genCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(9000))
	return fmt.Sprintf("%04d", n.Int64()+1000)
}

// Alibaba Cloud SMS signature v1
func smsSignature(params map[string]string, method, accessSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canon string
	for _, k := range keys {
		canon += "&" + percentEncode(k) + "=" + percentEncode(params[k])
	}
	canon = canon[1:]

	str := method + "&" + percentEncode("/") + "&" + percentEncode(canon)
	mac := hmac.New(sha1.New, []byte(accessSecret+"&"))
	mac.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percentEncode(s string) string {
	return url.QueryEscape(strings.ReplaceAll(strings.ReplaceAll(s, "+", "%20"), "*", "%2A"))
}

func sendSMS(phone, code string) error {
	accessKey := os.Getenv("ALIBABA_ACCESS_KEY")
	accessSecret := os.Getenv("ALIBABA_ACCESS_SECRET")
	signName := os.Getenv("ALIBABA_SMS_SIGN")
	templateCode := os.Getenv("ALIBABA_SMS_TEMPLATE")
	if accessKey == "" || accessSecret == "" {
		fmt.Println("[SMS] credentials not configured, code for", phone, ":", code)
		return nil // mock mode
	}
	if signName == "" || templateCode == "" {
		return fmt.Errorf("SMS sign name or template code not configured")
	}

	params := map[string]string{
		"AccessKeyId":      accessKey,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"SignName":         signName,
		"TemplateCode":     templateCode,
		"TemplateParam":    `{"code":"` + code + `","product":"乒云"}`,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}
	params["Signature"] = smsSignature(params, "GET", accessSecret)

	query := ""
	for k, v := range params {
		query += "&" + percentEncode(k) + "=" + percentEncode(v)
	}
	resp, err := http.Get("https://dysmsapi.aliyuncs.com/?" + query[1:])
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("[SMS]", string(body))
	return nil
}

// POST /api/auth/send-code
func AuthSendCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if len(phone) < 11 {
		http.Error(w, "请输入正确的手机号", http.StatusBadRequest)
		return
	}

	// IP rate limit: max 5 per hour
	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" { clientIP = r.Header.Get("X-Forwarded-For") }
	if clientIP == "" { clientIP = r.RemoteAddr }
	if idx := strings.LastIndex(clientIP, ":"); idx > 0 && !strings.Contains(clientIP, "[") { clientIP = clientIP[:idx] }
	if !checkIPRate(clientIP) {
		http.Error(w, "操作太频繁，请稍后再试", http.StatusTooManyRequests)
		return
	}

	// Find player by phone
	var pid int
	var pname string
	db.DB.QueryRow(`SELECT id, name FROM players WHERE phone=$1`, phone).Scan(&pid, &pname)
	if pid == 0 {
		http.Error(w, "该手机号未绑定球员", http.StatusNotFound)
		return
	}

	// Rate limit: max 3 attempts in 5 min, 1 per 60s
	codeStoreMu.Lock()
	if s, ok := codeStore[phone]; ok && s.attempts >= 3 && time.Since(s.expires) < 5*time.Minute {
		codeStoreMu.Unlock()
		http.Error(w, "发送次数过多，请5分钟后再试", http.StatusTooManyRequests)
		return
	}
	if s, ok := codeStore[phone]; ok && time.Since(s.expires) < 60*time.Second {
		codeStoreMu.Unlock()
		http.Error(w, "验证码已发送，请60秒后重试", http.StatusTooManyRequests)
		return
	}
	codeStoreMu.Unlock()

	code := genCode()
	if err := sendSMS(phone, code); err != nil {
		http.Error(w, "短信发送失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	codeStoreMu.Lock()
	attempts := 0
	if s, ok := codeStore[phone]; ok {
		attempts = s.attempts
	}
	codeStore[phone] = struct {
		code     string
		expires  time.Time
		attempts int
	}{code, time.Now(), attempts + 1}
	codeStoreMu.Unlock()

	writeJSON(w, map[string]string{"status": "ok"})
}

// POST /api/auth/verify
func AuthVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	codeStoreMu.Lock()
	s, ok := codeStore[req.Phone]
	if !ok || s.code != req.Code || time.Now().After(s.expires.Add(5*time.Minute)) {
		codeStoreMu.Unlock()
		http.Error(w, "验证码错误或已过期", http.StatusUnauthorized)
		return
	}
	delete(codeStore, req.Phone) // one-time use
	codeStoreMu.Unlock()

	// Get player
	var pid int
	var pname string
	db.DB.QueryRow(`SELECT id, name FROM players WHERE phone=$1`, req.Phone).Scan(&pid, &pname)

	// Check if player has an admin account
	var hasAccount bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM admin_users WHERE player_id=$1 AND deleted=false)`, pid).Scan(&hasAccount)

	// Auto-claim: if no admin account yet, create one (needs username setup)
	if !hasAccount {
		db.DB.Exec(`INSERT INTO admin_users (username, display_name, group_id, player_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			"player_"+strconv.Itoa(pid), pname, 4, pid) // participant group
	}

	// Set identity cookie
	token := hex.EncodeToString([]byte(fmt.Sprintf("%d:%d", pid, time.Now().Unix())))
	http.SetCookie(w, &http.Cookie{
		Name: "ping_id", Value: token, Path: "/", MaxAge: 86400 * 30, HttpOnly: false,
	})
	writeJSON(w, map[string]interface{}{
		"player_id": pid, "player_name": pname,
		"need_setup": !hasAccount,
	})
}

// GET /api/auth/me — current identity from cookie
func AuthMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("ping_id")
	if err != nil {
		http.Error(w, "not logged in", http.StatusUnauthorized)
		return
	}

	var pid int
	fmt.Sscanf(cookie.Value, "%d:", &pid)
	if pid == 0 {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var name string
	db.DB.QueryRow(`SELECT name FROM players WHERE id=$1`, pid).Scan(&name)
	writeJSON(w, map[string]interface{}{"player_id": pid, "player_name": name})
}
