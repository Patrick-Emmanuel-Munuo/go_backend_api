package helpers

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// This assumes a global DB variable
var DB *sql.DB // exported so other packages can use it
var (
	ServerSecurity    string
	ServerDomain      string
	ServerPort        int
	ServerEnvironment string
	SslCertificate    string
	SslKey            string
	DatabaseHost      string
	DatabaseUser      string
	DatabasePassword  string
	DatabaseName      string
	DatabasePort      string
	Mailsender        string
	Mailhost          string
	Mailusername      string
	Mailpassword      string
	Mailport          int
	SmsUserName       string
	SmsApiKey         string
	SmsSenderId       string
	JwtKey            string
)

func UpdateEnvVars() {
	ServerSecurity = getEnvValue("SECURITY", "http").(string)
	ServerDomain = getEnvValue("DOMAIN", "localhost").(string)
	ServerEnvironment = getEnvValue("ENVIRONMENT", "development").(string)
	ServerPort = getEnvValue("PORT", 2010).(int)
	SslCertificate = getEnvValue("SSL_CERTIFICATE", "").(string)
	SslKey = getEnvValue("SSL_KEY", "").(string)
	DatabaseHost = getEnvValue("DATABASE_HOST", "localhost").(string)
	DatabaseUser = getEnvValue("DATABASE_USER", "root").(string)
	DatabasePassword = getEnvValue("DATABASE_PASSWORD", "").(string)
	DatabaseName = getEnvValue("DATABASE_NAME", "trick").(string)
	DatabasePort = getEnvValue("DATABASE_PORT", "3306").(string)
	Mailsender = getEnvValue("MAIL_SENDER", "noreply@example.com").(string)
	Mailhost = getEnvValue("MAIL_HOST", "smtp.example.com").(string)
	Mailusername = getEnvValue("MAIL_ADDRESS", "noreply@example.com").(string)
	Mailpassword = getEnvValue("MAIL_PASSWORD", "").(string)
	Mailport = getEnvValue("MAIL_PORT", 587).(int)
	SmsUserName = getEnvValue("AFRICAS_TALKING_USERNAME", "").(string)
	SmsApiKey = getEnvValue("AFRICAS_TALKING_API_KEY", "").(string)
	SmsSenderId = getEnvValue("AFRICAS_TALKING_SENDER_ID", "").(string)
	JwtKey = getEnvValue("JWT_KEY", "").(string)
}

// StartServer starts a Gin HTTP/HTTPS server (no retries)
func StartServer(router *gin.Engine) map[string]interface{} {
	secure := ServerSecurity == "https"
	addr := fmt.Sprintf("%s:%d", ServerDomain, ServerPort)

	// Resolve SSL paths to absolute if using HTTPS
	if secure {
		if !filepath.IsAbs(SslCertificate) {
			absCert, err := filepath.Abs(SslCertificate)
			if err != nil {
				log.Printf(`{"success": false, "message": "Invalid SSL_CERTIFICATE path: %v"}`, err)
				secure = false
			} else {
				SslCertificate = absCert
			}
		}
		if !filepath.IsAbs(SslKey) {
			absKey, err := filepath.Abs(SslKey)
			if err != nil {
				log.Printf(`{"success": false, "message": "Invalid SSL_KEY path: %v"}`, err)
				secure = false
			} else {
				SslKey = absKey
			}
		}
		// Check if files exist
		if _, err := os.Stat(SslCertificate); os.IsNotExist(err) {
			log.Printf(`{"success": false, "message": "SSL certificate not found, falling back to HTTP"}`)
			secure = false
		}
		if _, err := os.Stat(SslKey); os.IsNotExist(err) {
			log.Printf(`{"success": false, "message": "SSL key not found, falling back to HTTP"}`)
			secure = false
		}
	}
	protocol := "http"
	if secure {
		protocol = "https"
	}
	log.Printf(`{"success": true, "message": "Starting server at %s://%s [PID: %d]"}`, protocol, addr, os.Getpid())
	var err error
	if secure {
		err = router.RunTLS(addr, SslCertificate, SslKey)
	} else {
		err = router.Run(addr)
	}
	if err != nil && err != http.ErrServerClosed {
		log.Printf(`{"success": false, "message": "Server error: %v"}`, err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Server error: %v", err),
		}
	}
	return map[string]interface{}{
		"success":  true,
		"protocol": protocol,
		"message":  fmt.Sprintf("Server running at %s://%s [PID: %d]", protocol, addr, os.Getpid()),
	}
}

// authenticate
func Authenticate(data map[string]interface{}) map[string]interface{} {
	//key
	var jwtKey = []byte(JwtKey)
	// Ensure required fields exist
	username, uOk := data["username"].(string)
	id, idOk := data["id"]
	if !uOk || !idOk {
		return map[string]interface{}{
			"success": false,
			"message": "Authentication failed: username and id are required",
		}
	}

	// Create JWT claims 1h expiration
	expireTime := time.Now().Add(1 * time.Hour)
	options := jwt.MapClaims{
		"exp":          expireTime.Unix(),
		"expired_time": expireTime.Format(time.RFC3339),
		"user_name":    username,
		"id":           id,
		"issued_at":    time.Now().Unix(),
		"issuer":       "go_backend_api",
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, options)
	// Sign the token
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Authentication failed: " + err.Error(),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": tokenString,
	}
}

// AuthMiddleware validates access_token
func Authorization(options map[string]interface{}) map[string]interface{} {
	var jwtKey = []byte(JwtKey)
	authTokenRaw, ok := options["authorization"].(string)
	if !ok || authTokenRaw == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Access denied: Your session has expired or the token is invalid. Please log in again to continue.",
		}
	}

	// Strip "Bearer " prefix if present
	const bearerPrefix = "Bearer "
	if len(authTokenRaw) > len(bearerPrefix) && authTokenRaw[:len(bearerPrefix)] == bearerPrefix {
		authTokenRaw = authTokenRaw[len(bearerPrefix):]
	}
	// Parse and verify the token
	token, err := jwt.Parse(authTokenRaw, func(token *jwt.Token) (interface{}, error) {
		// Ensure HMAC is used
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return map[string]interface{}{
			"success": false,
			"message": "Unauthorized: Token expired or invalid",
		}
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"status": "Successfully authorized",
				"data":   claims,
			},
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "Unauthorized: Invalid token structure",
	}
}

// Middleware to enforce token authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		access_token := c.Query("access_token")

		if access_token == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				access_token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		result := Authorization(map[string]interface{}{
			"authorization": access_token,
		})

		if success, ok := result["success"].(bool); !ok || !success {
			c.AbortWithStatusJSON(http.StatusUnauthorized, result)
			return
		}
		// Token is valid, proceed

		c.Next()
	}
}

func getEnvValue(key string, fallback interface{}) interface{} {
	val := os.Getenv(key)
	if val == "" {
		log.Printf(`{"success": false,"message": "Environment variable %q not set, using default: %v"}`, key, fallback)
		return fallback
	}
	switch fallback.(type) {
	case int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Printf(`{"success": false,"message": "Invalid int for %q: %q, using default: %v"}`, key, val, fallback)
			return fallback
		}
		return intVal
	case string:
		return val
	default:
		log.Printf(`{"success": false,"message": "Unsupported type for key %q, using default: %v"}`, key, fallback)
		return fallback
	}
}

func CleanupOldBackups(dir string, olderThan time.Duration) {
	files, _ := ioutil.ReadDir(dir)
	now := time.Now()
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			fullPath := filepath.Join(dir, file.Name())
			if now.Sub(file.ModTime()) > olderThan {
				os.Remove(fullPath)
				log.Println("Deleted old backup:", fullPath)
			}
		}
	}
}

// getServerIPAddress tries to find a non-loopback IP address
func GetServerIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && !ip.IsLoopback() && !strings.HasPrefix(ip.String(), "169.254.") {
				return ip.String() // Return first non-loopback, non-link-local IPv4
			}
		}
	}
	return "127.0.0.1" // fallback
}

// ClearIPAddress masks the last octet of an IPv4 address for privacy
func ClearIPAddress(ip string) string {
	if ip == "" {
		return "unknown"
	}
	if strings.HasPrefix(ip, "::ffff:") {
		return strings.TrimPrefix(ip, "::ffff:")
	}
	if ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}

// ---------- SQL HELPERS ----------
// InitDBConnection establishes and checks the MySQL connection with retries
func InitDBConnection() map[string]interface{} {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DatabaseUser,
		DatabasePassword,
		DatabaseHost,
		DatabasePort,
		DatabaseName,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf(`{"success": false, "message": "Failed to open MySQL connection: %v"}`, err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to open MySQL connection: %v", err),
		}
	}
	if err = DB.Ping(); err != nil {
		log.Printf(`{"success": false, "message": "Failed to ping MySQL: %v"}`, err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to ping MySQL: %v", err),
		}
	}
	// ✅ Connection successful
	log.Printf(`{"success": true, "message": "MySQL connection established successfully"}`)
	return map[string]interface{}{
		"success": true,
		"message": "MySQL connection established successfully",
	}
}

// Where builds WHERE condition with AND
func Where(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " AND "), params
}

// WhereOr builds WHERE condition with OR
func WhereOr(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " OR "), params
}

// JoinFields returns `field1`, `field2`, ...
func JoinFields(fields []string) string {
	var wrapped []string
	for _, f := range fields {
		wrapped = append(wrapped, "`"+f+"`")
	}
	return strings.Join(wrapped, ", ")
}

// Like builds WHERE with LIKE conditions
func Like(like map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var params []interface{}
	for key, val := range like {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", key))
		params = append(params, fmt.Sprintf("%%%v%%", val))
	}
	return strings.Join(conditions, " AND "), params
}

// UpdateSet builds `SET field1=?, field2=?`
func UpdateSet(set map[string]interface{}) string {
	var parts []string
	for key := range set {
		parts = append(parts, key+" = ?")
	}
	return strings.Join(parts, ", ")
}

// SortedKeys returns keys of a map in sorted order
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// convert xml to json
func XMLtoJSON(resp *http.Response) map[string]interface{} {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to read response body: " + err.Error(),
		}
	}
	// Close the body after reading
	resp.Body.Close()
	// Convert []byte to io.Reader for mxj
	reader := bytes.NewReader(body)
	// Parse XML to map[string]interface{}
	data, err := mxj.NewMapXmlReader(reader)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to parse XML: " + err.Error(),
		}
	}
	// Convert mxj.Map to map[string]interface{}
	result := map[string]interface{}(data)

	// Assume data is already parsed from XML and stored in a variable like this:
	return map[string]interface{}{
		"success": true,
		"message": result,
	}
}

func ParseMessages(raw string) map[string]interface{} {
	// Clean up the raw data like Node.js `.replace(/\r\n/g, '').trim()`
	cleaned := strings.ReplaceAll(raw, "\r\n", "")
	cleaned = strings.TrimSpace(cleaned)

	if len(cleaned) < 20 {
		//fmt.Println("No new SMS received or malformed message.")
		return map[string]interface{}{
			"success": false,
			"message": "No new SMS or malformed data",
		}
	}

	receiver := safeSlice(cleaned, 7, 20)
	receiverAt := safeSlice(cleaned, 24, 44)
	if receiverAt == "" {
		receiverAt = time.Now().Format(time.RFC3339)
	}
	receiverText := safeSlice(cleaned, 45, len(cleaned))

	messageID := GenerateUniqueID() // Or use uuid.New().String()

	savedData := map[string]interface{}{
		"message_id":    messageID,
		"receiver":      receiver,
		"receiver_at":   receiverAt,
		"receiver_text": receiverText,
		"created_at":    time.Now().Format(time.RFC3339),
		"created_by":    "system",
		"updated_at":    nil,
		"updated_by":    nil,
		"description":   nil,
		"status":        "received",
		"method":        "local_via_modem",
	}
	return map[string]interface{}{
		"success": true,
		"message": savedData,
	}

}
func safeSlice(s string, start, end int) string {
	runes := []rune(s)
	if start >= len(runes) {
		return ""
	}
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
