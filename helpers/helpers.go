package helpers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/clbanning/mxj"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// This assumes a global DB variable
var DB *sql.DB // exported so other packages can use it
var (
	ServerSecurity          string
	ServerDomain            string
	ServerPort              int
	ServerEnvironment       string
	SslCertificate          string
	SslKey                  string
	DatabaseHost            string
	DatabaseUser            string
	DatabasePassword        string
	DatabaseName            string
	DatabasePort            string
	Mailsender              string
	Mailhost                string
	Mailusername            string
	Mailpassword            string
	Mailport                int
	SmsUserName             string
	SmsApiKey               string
	SmsSenderId             string
	JwtKey                  string
	EnableEncripted         bool
	EncryptionKey           string
	initializationVector    string
	EncryptionAlgorithm     string
	EncryptionInitializatin string
	// encryptionKey and initializationVector are the binary forms used by crypto
	EncryptionKey_Byte        []byte
	initializationVector_Byte []byte
)

func UpdateEnvVars() {
	ServerSecurity = getEnvValue("SECURITY", "http").(string)
	ServerDomain = getEnvValue("DOMAIN", "localhost").(string)
	ServerEnvironment = getEnvValue("SERVER_ENVIROMENT", "development").(string)
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
	SmsUserName = getEnvValue("SMS_USERNAME", "").(string)
	SmsApiKey = getEnvValue("SMS_API_KEY", "").(string)
	SmsSenderId = getEnvValue("SMS_SENDER_ID", "").(string)
	JwtKey = getEnvValue("JWT_KEY", "").(string)
	EnableEncripted = getEnvValue("EnableEncripted", false).(bool)
	EncryptionKey = (getEnvValue("EncryptionKey", "1234567890123456").(string))
	EncryptionAlgorithm = (getEnvValue("EncryptionAlgorithm", "aes-128-cbc").(string))
	EncryptionInitializatin = (getEnvValue("EncryptionInitializatin", "2d52550dc714656b").(string))
}

func getEnvValue(key string, fallback interface{}) interface{} {
	val := os.Getenv(key)
	if val == "" {
		//LogJSON(false, fmt.Sprintf("Environment variable %q not set", key))
		// Return zero value based on type
		switch fallback.(type) {
		case int:
			return 0
		case bool:
			return false
		case string:
			return ""
		default:
			return nil
		}
	}

	switch fallback.(type) {
	case int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			//LogJSON(false, fmt.Sprintf("Invalid int for environment variable %q: %q", key, val))
			return 0
		}
		return intVal
	case bool:
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			//LogJSON(false, fmt.Sprintf("Invalid bool for environment variable %q: %q", key, val))
			return false
		}
		return boolVal
	case string:
		return val
	default:
		LogJSON(false, fmt.Sprintf("Unsupported type for environment variable %q", key))
		return nil
	}
}

// LogJSON  prints logs in JSON format with optional colors
func LogJSON(success bool, message string) {
	entry := map[string]interface{}{
		//"timestamp": time.Now().UTC().Format("2006-01-02 15:04:05"),
		"success": success,
		"message": message,
	}

	if data, err := json.Marshal(entry); err == nil {
		if success {
			color.New(color.FgGreen).Println(string(data))
		} else {
			color.New(color.FgRed).Println(string(data))
		}
	} else {
		color.New(color.FgRed).Println("Error marshaling log JSON:", err)
	}
}

// StartServer starts Gin HTTP/HTTPS server
func StartServer(router *gin.Engine) map[string]interface{} {
	secure := ServerSecurity == "https"
	addr := fmt.Sprintf("%s:%d", ServerDomain, ServerPort)

	// Resolve SSL paths if HTTPS
	if secure {
		if !filepath.IsAbs(SslCertificate) {
			absCert, err := filepath.Abs(SslCertificate)
			if err != nil {
				LogJSON(false, fmt.Sprintf("Invalid SSL_CERTIFICATE path: %v, falling back to HTTP", err))
				secure = false
			} else {
				SslCertificate = absCert
			}
		}
		if !filepath.IsAbs(SslKey) {
			absKey, err := filepath.Abs(SslKey)
			if err != nil {
				LogJSON(false, fmt.Sprintf("Invalid SSL_CERTIFICATE path: %v, falling back to HTTP", err))
				secure = false
			} else {
				SslKey = absKey
			}
		}
		if _, err := os.Stat(SslCertificate); os.IsNotExist(err) {
			LogJSON(false, fmt.Sprintf("SSL certificate not found at %s, falling back to HTTP", SslCertificate))
			secure = false
		}
		if _, err := os.Stat(SslKey); os.IsNotExist(err) {
			LogJSON(false, fmt.Sprintf("SSL key not found at %s, falling back to HTTP", SslKey))
			secure = false
		}
	}

	protocol := "http"
	if secure {
		protocol = "https"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		LogJSON(true, fmt.Sprintf("Server running at %s://%s [PID: %d]", protocol, addr, os.Getpid()))
		var err error
		if secure {
			err = srv.ListenAndServeTLS(SslCertificate, SslKey)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			LogJSON(false, fmt.Sprintf("Server error: %v", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	LogJSON(true, "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		LogJSON(false, fmt.Sprintf("Server forced to shutdown: %v", err))
	} else {
		LogJSON(true, "Server exited gracefully")
	}
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Server running at %s://%s [PID: %d]", protocol, addr, os.Getpid()),
	}
}

// --- AES helpers ---
// Encript encrypts the JSON inside data["message"]
func Encript(data map[string]interface{}) map[string]interface{} {
	message, ok := data["message"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "message field missing",
		}
	}

	// Convert message to JSON string
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "JSON marshal error: " + err.Error(),
		}
	}

	// If encryption is disabled, return JSON string
	if !EnableEncripted {
		data["message"] = string(jsonBytes)
		return data
	}
	// Ensure key and IV are present and valid
	// Prepare binary key and IV used by crypto
	if EncryptionKey != "" {
		EncryptionKey_Byte = []byte(EncryptionKey)
		// if key length is not 16/24/32, log a warning (must be one of these)
		if !(len(EncryptionKey_Byte) == 16 || len(EncryptionKey_Byte) == 24 || len(EncryptionKey_Byte) == 32) {
			LogJSON(false, fmt.Sprintf("Invalid EncryptionKey length (%d). AES requires 16, 24 or 32 bytes.", len(EncryptionKey_Byte)))
			// You may choose to pad/truncate here — we prefer explicit error.
		}
	} else {
		EncryptionKey_Byte = nil
	}
	if EncryptionKey_Byte == nil || !(len(EncryptionKey_Byte) == 16 || len(EncryptionKey_Byte) == 24 || len(EncryptionKey_Byte) == 32) {
		return map[string]interface{}{
			"success": false,
			"message": "invalid encryption key configuration",
		}
	}

	//find initializationVector_Byte
	initializationVector_Byte = []byte(EncryptionInitializatin)
	if initializationVector_Byte == nil || len(initializationVector_Byte) != aes.BlockSize {
		return map[string]interface{}{
			"success": false,
			"message": "invalid initialization vector (IV) configuration",
		}
	}

	block, err := aes.NewCipher(EncryptionKey_Byte)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "AES error: " + err.Error(),
		}
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - (len(jsonBytes) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	padded := append(jsonBytes, padtext...)

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, initializationVector_Byte)
	mode.CryptBlocks(ciphertext, padded)

	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return map[string]interface{}{
		"success": data["success"],
		"message": encoded,
	}
}

// Decript decrypts the JSON inside data["message"]
func Decript(data map[string]interface{}) map[string]interface{} {
	messageStr, ok := data["message"].(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "message field must be a string",
		}
	}

	if !EnableEncripted {
		// If encryption disabled, attempt to parse message as JSON string into object
		var original interface{}
		if err := json.Unmarshal([]byte(messageStr), &original); err != nil {
			// not JSON, just return the raw string
			return data
		}
		data["message"] = original
		return data
	}
	// Prepare binary key and IV used by crypto
	if EncryptionKey != "" {
		EncryptionKey_Byte = []byte(EncryptionKey)
		// if key length is not 16/24/32, log a warning (must be one of these)
		if !(len(EncryptionKey_Byte) == 16 || len(EncryptionKey_Byte) == 24 || len(EncryptionKey_Byte) == 32) {
			LogJSON(false, fmt.Sprintf("Invalid EncryptionKey length (%d). AES requires 16, 24 or 32 bytes.", len(EncryptionKey_Byte)))
			// You may choose to pad/truncate here — we prefer explicit error.
		}
	} else {
		EncryptionKey_Byte = nil
	}
	if EncryptionKey_Byte == nil || !(len(EncryptionKey_Byte) == 16 || len(EncryptionKey_Byte) == 24 || len(EncryptionKey_Byte) == 32) {
		return map[string]interface{}{
			"success": false,
			"message": "invalid encryption key configuration",
		}
	}
	//find initializationVector_Byte
	initializationVector_Byte = []byte(initializationVector)
	if initializationVector_Byte == nil || len(initializationVector_Byte) != aes.BlockSize {
		return map[string]interface{}{
			"success": false,
			"message": "invalid initialization vector (IV) configuration",
		}
	}
	if initializationVector_Byte == nil || len(initializationVector_Byte) != aes.BlockSize {
		return map[string]interface{}{
			"success": false,
			"message": "invalid initialization vector (IV) configuration",
		}
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(messageStr)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Base64 decode error: " + err.Error(),
		}
	}

	block, err := aes.NewCipher(EncryptionKey_Byte)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "AES error: " + err.Error(),
		}
	}

	if len(cipherBytes)%block.BlockSize() != 0 {
		return map[string]interface{}{
			"success": false,
			"message": "ciphertext is not a multiple of the block size",
		}
	}

	plaintext := make([]byte, len(cipherBytes))
	mode := cipher.NewCBCDecrypter(block, initializationVector_Byte)
	mode.CryptBlocks(plaintext, cipherBytes)

	// PKCS7 unpadding
	length := len(plaintext)
	if length == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "decrypted plaintext is empty",
		}
	}
	padLen := int(plaintext[length-1])
	if padLen <= 0 || padLen > block.BlockSize() {
		return map[string]interface{}{
			"success": false,
			"message": "invalid padding size",
		}
	}
	unpadded := plaintext[:length-padLen]

	var original interface{}
	if err := json.Unmarshal(unpadded, &original); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "JSON unmarshal error: " + err.Error(),
		}
	}

	return map[string]interface{}{
		"success": data["success"],
		"message": original,
	}
}

// authenticate
func Authenticate(data map[string]interface{}) map[string]interface{} {
	//key
	var jwtKey = []byte(JwtKey)
	// Ensure required fields exist
	user_name, uOk := data["user_name"].(string)
	id, idOk := data["id"]
	if !uOk || !idOk {
		return map[string]interface{}{
			"success": false,
			"message": "Authentication failed: user_name and id are required",
		}
	}

	// Create JWT claims 1h expiration
	expireTime := time.Now().Add(12 * time.Hour)
	options := jwt.MapClaims{
		"exp":          expireTime.Unix(),
		"expired_time": expireTime.Format(time.RFC3339),
		"user_name":    user_name,
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
			} else if authHeader != "" {
				access_token = authHeader
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
		//logger.Errorf("Failed to open MySQL connection: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to open MySQL connection: %v", err),
		}
	}

	if err = DB.Ping(); err != nil {
		//logger.Errorf("Failed to ping MySQL: %v", err)
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to ping MySQL: %v", err),
		}
	}

	// ✅ Connection successful
	//logger.Info("MySQL connection established successfully")
	return map[string]interface{}{
		"success": true,
		"message": "MySQL connection established successfully",
	}
}

// Where builds WHERE condition with AND
func GenerateWhere(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", EscapeId(k)))
		params = append(params, v)
	}
	return strings.Join(parts, " AND "), params
}

// WhereOr builds WHERE condition with OR
func GenerateWhereOr(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", EscapeId(k)))
		params = append(params, v)
	}
	return strings.Join(parts, " OR "), params
}

// Like builds WHERE with LIKE conditions
func GenerateLike(like map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var params []interface{}
	for key, val := range like {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", EscapeId(key)))
		params = append(params, fmt.Sprintf("%%%v%%", val))
	}
	return strings.Join(conditions, " AND "), params
}

// UpdateSet builds `SET field1=?, field2=?`
func GenerateSet(set map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for key := range set {
		parts = append(parts, fmt.Sprintf("%s = ?", EscapeId(key)))
	}
	return strings.Join(parts, ", "), params
}

// Select generate
func GenerateSelect(fields interface{}) string {
	var strFields []string

	switch v := fields.(type) {
	case []string:
		strFields = v
	case []interface{}:
		strFields = make([]string, len(v))
		for i, f := range v {
			strFields[i], _ = f.(string)
		}
	default:
		return "*" // fallback if unsupported type
	}

	if len(strFields) == 0 {
		return "*"
	}
	return strings.Join(EscapeIds(strFields), ", ")
}

// EscapeId safely escapes table/column names using backticks
func EscapeId(identifier string) string {
	// Strip dangerous characters except underscore and alphanumerics
	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	safe := re.ReplaceAllString(identifier, "")
	return "`" + safe + "`"
}

// EscapeIdentifiers for multiple fields
func EscapeIds(identifiers []string) []string {
	escaped := make([]string, len(identifiers))
	for i, id := range identifiers {
		escaped[i] = EscapeId(id)
	}
	return escaped
}

// Scan SQL rows into []map[string]interface{}
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := map[string]interface{}{}
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
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
