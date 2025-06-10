package controllers

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"vartrick-server/helpers"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

//var db *sql.DB

type Options struct {
	Table       string                 `json:"table"`
	Select      []string               `json:"select"`
	Condition   map[string]interface{} `json:"condition"`
	OrCondition map[string]interface{} `json:"or_condition"`
}

// SMSOptions defines SMS parameters
type SMSOptions struct {
	To      []string
	Message string
}

// AfricaTalkingXMLResponse is for parsing XML response from Africa's Talking
type AfricaTalkingXMLResponse struct {
	XMLName        xml.Name `xml:"AfricasTalkingResponse"`
	SMSMessageData struct {
		Message    string `xml:"Message"`
		Recipients struct {
			Recipient struct {
				Number       string `xml:"number"`
				Cost         string `xml:"cost"`
				Status       string `xml:"status"`
				StatusCode   string `xml:"statusCode"`
				MessageID    string `xml:"messageId"`
				MessageParts string `xml:"messageParts"`
			} `xml:"Recipient"`
		} `xml:"Recipients"`
	} `xml:"SMSMessageData"`
}
type JsonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MailOptions struct {
	To          string
	Subject     string
	Message     string
	HTML        string
	Attachments []string
}

// Generate OTP
func GenerateOTP() map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in GenerateOTP:", r)
		}
	}()
	const otpCharset = "0123456789"
	const otpLength = 6
	otp := make([]byte, otpLength)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpCharset))))
		if err != nil {
			log.Println("Error generating secure OTP:", err)
			return map[string]interface{}{
				"success": false,
				"message": "Failed to generate secure OTP",
				"otp":     nil,
			}
		}
		otp[i] = otpCharset[num.Int64()]
	}
	otp_formated := string(otp[:3]) + "-" + string(otp[3:])
	return map[string]interface{}{
		"success": true,
		"message": "OTP generated successfully",
		"otp":     otp_formated,
	}
}

// send OTP via either mail or phone
func SendOTP(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in SendOTP:", r)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		}
	}()
	otp := GenerateOTP()
	if !otp["success"].(bool) {
		c.JSON(http.StatusInternalServerError, otp)
		return
	}
	// Here you would typically send the OTP via email or SMS
	log.Println("Generated OTP:", otp["otp"])
	c.JSON(http.StatusOK, otp)
}

// decript token
func DecriptToken(options map[string]interface{}) map[string]interface{} {
	// Validate token field
	tokenRaw, ok := options["token"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Token field is required.",
		}
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Token must be a string.",
		}
	}
	// Remove dashes
	tokenStr = strings.ReplaceAll(tokenStr, "-", "")
	// Validate format: must be 20 digits
	if len(tokenStr) != 20 || !helpers.IsAllDigits(tokenStr) {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid token format. Must be 20 digits (dashes allowed).",
		}
	}
	// Convert token string to integer
	// Convert to *big.Int
	tokenBigInt := new(big.Int)
	_, success := tokenBigInt.SetString(tokenStr, 10)
	if !success {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to parse token as big integer.",
		}
	}
	// Convert to binary with 66 bits
	tokenBin := fmt.Sprintf("%066b", tokenBigInt)
	//tokenBin := helpers.DecToBin(token,)
	// Generate decoder key (already returns bin string)
	keyRes := helpers.GenerateDecoderKey()
	if !keyRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": keyRes["message"].(string),
		}
	}

	keyBin := keyRes["message"].(string)
	keyBytes := helpers.BinStrToBytes(keyBin)

	// Perform transposition and extract class bits
	tokRes := helpers.TranspositionAndRemoveClassBits(tokenBin)
	if !tokRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to extract token blocks",
		}
	}
	tokData := tokRes["message"].(map[string]interface{})
	restored := tokData["data"].(string)
	classBits := tokData["class"].(string)
	if len(restored) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Token block must be at least 64 bits.",
		}
	}

	// Decrypt the first 64-bit block (8 bytes)
	encBytes := helpers.BinStrToBytes(restored[:64])
	decRes := helpers.Decrypt3DES(encBytes, keyBytes)
	if !decRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Decryption failed: " + decRes["message"].(string),
		}
	}
	decBytes := decRes["message"].([]byte)

	// Parse decrypted binary string
	binStr := helpers.BytesToBinStr(decBytes)
	if len(binStr) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Decrypted data is less than 64 bits.",
		}
	}
	rndBlock := binStr[0:3]
	tidBlock := binStr[3:25]
	amtBlock := binStr[25:48]
	crcBlock := binStr[48:64]

	//fmt.Println("amtBlock:", amtBlock)
	//fmt.Println("tidBlock:", tidBlock)
	//fmt.Println("binStr:", binStr)
	//crc validation
	dataBin := binStr[:48]
	crcInTokenBin := binStr[48:64]
	// Convert dataBin to hex string, padded to 14 hex chars (48 bits = 12 bytes = 14 hex chars with leading zeros)
	dataHex := helpers.BinToHex(dataBin)
	if len(dataHex) < 14 {
		dataHex = fmt.Sprintf("%014s", dataHex)
	}
	// Convert hex string to bytes
	dataBytes, err := helpers.HexToBytes(dataHex)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to convert data hex to bytes: " + err.Error(),
		}
	}
	// Calculate CRC16 of dataBytes
	crcRes := helpers.CalculateCRC16(dataBytes)
	if !crcRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "CRC calculation failed: " + crcRes["message"].(string),
		}
	}
	crcCalcBin := crcRes["message"].(string)
	// Pad CRC binary string to 16 bits if needed
	if len(crcCalcBin) < 16 {
		crcCalcBin = fmt.Sprintf("%016s", crcCalcBin)
	}
	// Validate CRC match
	if crcCalcBin != crcInTokenBin {
		return map[string]interface{}{
			"success": false,
			"message": "CRC mismatch - invalid token data",
		}
	}

	// Time validations
	// Parse timestamp (in minutes since base date)
	tidMinutes, err := strconv.ParseInt(tidBlock, 2, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to parse timestamp block: " + err.Error(),
		}
	}
	timeNow := time.Now()
	// Assuming you have parsed tidMinutes from earlier as int64
	issueTime := helpers.BaseDate.Add(time.Duration(tidMinutes) * time.Minute)

	if timeNow.Sub(issueTime) > 365*24*time.Hour {
		return map[string]interface{}{
			"success": false,
			"message": "Token expired",
		}
	}
	if issueTime.Before(helpers.BaseDate) || issueTime.After(timeNow.Add(24*time.Hour)) {
		return map[string]interface{}{
			"success": false,
			"message": "Change meter base date",
		}
	}
	// Decode units block (23 bits)
	unitsRes := helpers.DecodeUnits(amtBlock)
	if !unitsRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to decode units: " + unitsRes["message"].(string),
		}
	}
	units := unitsRes["message"]
	// Calculate timesx
	//issueTime := helpers.BaseDate.Add(time.Duration(tidMinutes) * time.Minute)
	expiryTime := issueTime.AddDate(1, 0, 0)

	// Assemble result
	result := map[string]interface{}{
		"crc":                helpers.BinStrToDecimal(crcBlock),
		"class":              helpers.BinStrToDecimal(classBits),
		"identifier_minutes": tidMinutes,
		"units":              units,
		"issued_date":        issueTime.Format(time.RFC3339),
		"expiry_date":        expiryTime.Format(time.RFC3339),
		"base_date":          helpers.BaseDate,
		"random":             helpers.BinStrToDecimal(rndBlock),
		"status":             "Token successfully decrypted and parsed.",
	}
	return map[string]interface{}{
		"success": true,
		"message": result,
	}
}

// SendMessage sends an SMS using Africa's Talking API
func SendMessage(options SMSOptions) map[string]interface{} {
	if len(options.To) == 0 || strings.TrimSpace(options.Message) == "" {
		return map[string]interface{}{
			"Success": false,
			"Message": "both 'to' and 'message' are required and cannot be empty",
		}
	}
	formData := url.Values{}
	formData.Set("username", os.Getenv("AFRICAS_TALKING_USERNAME"))
	formData.Set("to", strings.Join(options.To, ","))
	formData.Set("message", options.Message)
	formData.Set("from", os.Getenv("AFRICAS_TALKING_SENDER_ID"))

	req, err := http.NewRequest("POST", "https://api.africastalking.com/version1/messaging", strings.NewReader(formData.Encode()))
	if err != nil {
		return map[string]interface{}{
			"Success": false,
			"Message": fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", os.Getenv("AFRICAS_TALKING_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Log full error for debugging
		log.Println("failed to send request error : ", err)
		// Clean user-facing message
		userMessage := "failed to send request to SMS service"
		// Optional: Provide more specific feedback based on error content
		if strings.Contains(err.Error(), "no such host") {
			userMessage = "unable to reach SMS server – check network"
		} else if strings.Contains(err.Error(), "dial tcp") {
			userMessage = "connection to SMS service failed"
		}
		return map[string]interface{}{
			"Success": false,
			"Message": userMessage,
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"Success": false,
			"Message": "failed to read response body",
		}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("SMS send failed with status %d. Response body: %s\n", resp.StatusCode, string(body))
		return map[string]interface{}{
			"Success": false,
			"Message": "SMS send failed. Please try again later.",
		}
	}
	// Parse XML response
	var parsedResponse AfricaTalkingXMLResponse
	err = xml.Unmarshal(body, &parsedResponse)
	if err != nil {
		log.Printf("XML parse error: %v\n", err) // for internal logs
		return map[string]interface{}{
			"Success": false,
			"Message": "Invalid response format from SMS provider.",
		}
	}
	status := parsedResponse.SMSMessageData.Recipients.Recipient.Status
	if status != "Success" {
		log.Printf("SMS error: Status=%s, StatusCode=%s\n", status, parsedResponse.SMSMessageData.Recipients.Recipient.StatusCode)
		return map[string]interface{}{
			"Success": false,
			"Message": "SMS not sent. Please try again later.",
			//data: parsedResponse,
		}
	}
	return map[string]interface{}{
		"Success": true,
		"Message": parsedResponse.SMSMessageData.Message,
		"Data":    parsedResponse,
	}
}

func SendMail(options map[string]interface{}) map[string]interface{} {
	// Validate required fields
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	msg, msgOk := options["Message"].(string)
	toRaw, toOk := options["To"]
	if !msgOk || !toOk || msg == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Both 'To' and 'Message' fields are required, and 'Message' must be a string.",
		}
	}
	var recipients []string
	switch v := toRaw.(type) {
	case string:
		if v == "" {
			return map[string]interface{}{
				"success": false,
				"message": "'To' field cannot be empty",
			}
		}
		recipients = []string{v}
	case []string:
		if len(v) == 0 {
			return map[string]interface{}{
				"success": false,
				"message": "'To' field slice cannot be empty",
			}
		}
		recipients = v
	default:
		return map[string]interface{}{
			"success": false,
			"message": "'To' field must be a string or slice of strings",
		}
	}
	// Basic validation on all recipients (check '@')
	for _, email := range recipients {
		if !emailRegex.MatchString(email) {
			return map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Invalid email address format: %s", email),
			}
		}
	}
	// Subject fallback
	subject, _ := options["Subject"].(string)
	if subject == "" {
		subject = msg
	}
	// HTML fallback
	html, _ := options["HTML"].(string)
	if html == "" {
		html = fmt.Sprintf("<h2>%s</h2>", msg)
	}

	// Required environment variables
	sender := os.Getenv("MAIL_SENDER")
	host := os.Getenv("MAIL_HOST")
	username := os.Getenv("MAIL_ADDRESS")
	password := os.Getenv("MAIL_PASSWORD")
	portStr := os.Getenv("MAIL_PORT")
	if sender == "" || host == "" || username == "" || password == "" || portStr == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing required environment variables for SMTP configuration.",
		}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid MAIL_PORT. Ensure it's a valid number.",
		}
	}
	// Compose the message
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipients...)
	m.SetHeader("Subject", subject)
	if cc, ok := options["CC"].([]string); ok && len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if bcc, ok := options["BCC"].([]string); ok && len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}
	m.SetBody("text/plain", msg)
	m.AddAlternative("text/html", html)
	// Handle attachments
	if attachments, ok := options["Attachments"].([]string); ok {
		for _, attachment := range attachments {
			if _, err := os.Stat(attachment); err != nil {
				return map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("Failed to attach file %s: %v", attachment, err),
				}
			}
			m.Attach(attachment)
		}
	}
	// Configure SMTP
	d := gomail.NewDialer(host, port, username, password)
	// Send email
	if err := d.DialAndSend(m); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to send email: %v", err),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"status":     "Mail sent successfully",
			"data":       "response",
			"timestamp":  time.Now().Format(time.RFC3339),
			"recipients": recipients,
			"subject":    subject,
		},
	}
}

// Backup runs mysqldump and returns the result as JSON-compatible map
func Backup() (map[string]interface{}, error) {
	timeNow := time.Now()
	fileName := fmt.Sprintf("mysql_backup_%d.sql", timeNow.Unix())
	publicDir := filepath.Join(".", "public")
	// Create directory if it doesn't exist
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		log.Println("Failed to create public dir:", err)
		return map[string]interface{}{"success": false, "message": err.Error()}, nil
	}

	filePath := filepath.Join(publicDir, fileName)

	// Run mysqldump
	cmd := exec.Command("mysqldump", "-h", "localhost", "-u", "root", "-pYOURPASSWORD", "vartrick")

	outfile, err := os.Create(filePath)
	if err != nil {
		log.Println("Failed to create dump file:", err)
		return map[string]interface{}{"success": false, "message": err.Error()}, nil
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	if err := cmd.Run(); err != nil {
		log.Println("mysqldump failed:", err)
		return map[string]interface{}{"success": false, "message": err.Error()}, nil
	}

	log.Println("Backup created:", filePath)
	return map[string]interface{}{
		"success": true,
		"message": "Backup created successfully",
		"file":    filePath,
	}, nil
}

// read mysql
func Read(c *gin.Context) {
	var options Options
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || (len(options.Condition) == 0 && len(options.OrCondition) == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Missing table name or condition(s)"})
		return
	}
	selectFields := "*"
	if len(options.Select) > 0 {
		selectFields = helpers.JoinFields(options.Select)
	}
	var whereClause string
	var params []interface{}
	if len(options.Condition) > 0 && len(options.OrCondition) > 0 {
		where1, params1 := helpers.Where(options.Condition)
		where2, params2 := helpers.WhereOr(options.OrCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	} else if len(options.Condition) > 0 {
		whereClause, params = helpers.Where(options.Condition)
	} else if len(options.OrCondition) > 0 {
		whereClause, params = helpers.WhereOr(options.OrCondition)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, options.Table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	// Convert []uint8 (MySQL bytes) to string for JSON compatibility
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": results})
}

// search mysql
func Search(c *gin.Context) {
	var options Options
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}
	if options.Table == "" || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing table name or condition(s)",
		})
		return
	}
	selectFields := "*"
	if len(options.Select) > 0 {
		selectFields = helpers.JoinFields(options.Select)
	}
	// Build LIKE clause and params safely
	whereClause, params := helpers.Like(options.Condition)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, options.Table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No data found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": results,
	})
}

// SearchBetween finds records where a field's value is between two values
func SearchBetween(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Field     string                 `json:"field"`
		Start     interface{}            `json:"start"`
		End       interface{}            `json:"end"`
		Condition map[string]interface{} `json:"condition"` // Additional conditions
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || options.Field == "" || options.Start == nil || options.End == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Table, field, start and end values are required"})
		return
	}
	whereClause := fmt.Sprintf("%s BETWEEN ? AND ?", options.Field)
	params := []interface{}{options.Start, options.End}
	if len(options.Condition) > 0 {
		additionalWhere, additionalParams := helpers.Where(options.Condition)
		whereClause += " AND " + additionalWhere
		params = append(params, additionalParams...)
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", options.Table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": results})
}

// List returns paginated results from a table
func List(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Page      int                    `json:"page"`
		PageSize  int                    `json:"page_size"`
		Condition map[string]interface{} `json:"condition"` // Optional conditions
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || options.Page <= 0 || options.PageSize <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Table name, page and page_size are required"})
		return
	}
	offset := (options.Page - 1) * options.PageSize
	whereClause, params := helpers.Where(options.Condition)
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT ? OFFSET ?", options.Table, whereClause)
	params = append(params, options.PageSize, offset)
	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {

		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": results})
}

// ListAll returns all records from a table
func ListAll(c *gin.Context) {
	var options struct {
		Table string `json:"table"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Table name is required"})
		return
	}
	query := fmt.Sprintf("SELECT * FROM %s", options.Table)
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": results})
}

// create mysql
func Create(c *gin.Context) {
	var options struct {
		Table string                 `json:"table"`
		Data  map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Both table name and data are required"})
		return
	}

	// Generate unique_id and add to data
	uniqueID := helpers.GenerateUniqueID()
	options.Data["unique_id"] = uniqueID

	// Build query: INSERT INTO table SET col1=?, col2=?, ...
	columns := []string{}
	values := []interface{}{}

	for col, val := range options.Data {
		columns = append(columns, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}
	query := fmt.Sprintf("INSERT INTO %s SET %s", options.Table, strings.Join(columns, ", "))

	// Execute query
	result, err := db.Exec(query, values...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": gin.H{"unique_id": uniqueID, "data": options.Data}})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Failed to insert data"})
		return
	}
}

// CreateBulk inserts multiple records in a single query
func CreateBulk(c *gin.Context) {
	var options struct {
		Table string                   `json:"table"`
		Data  []map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Both table name and data are required"})
		return
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	for _, row := range options.Data {
		if len(row) == 0 {
			continue // Skip empty rows
		}
		if len(columns) == 0 {
			for col := range row {
				columns = append(columns, col)
				placeholders = append(placeholders, "?")
			}
		}
		for _, val := range row {
			values = append(values, val)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", options.Table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	result, err := db.Exec(query, values...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%d rows inserted", rowsAffected)})
}

// update mysql
func Update(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Data      map[string]interface{} `json:"data"`
		Condition map[string]interface{} `json:"condition"`
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}

	// Validate required fields
	if options.Table == "" || len(options.Data) == 0 || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "table, data, and condition are required"})
		return
	}

	// Build SET and WHERE clauses
	setClause := helpers.UpdateSet(options.Data)
	whereClause, whereParams := helpers.Where(options.Condition) // FIX: get both values

	// Prepare query values in correct order: data first, then condition
	params := []interface{}{}
	for _, v := range options.Data {
		params = append(params, v)
	}
	params = append(params, whereParams...) // FIX: use extracted params from Where()

	// Final SQL update query
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", options.Table, setClause, whereClause)

	// Execute the query
	result, err := db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Evaluate update result
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Data updated successfully"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data was updated. It may be identical to the existing data or condition not matched"})
	}
}

// UpdateBulk updates multiple records in a single query
func UpdateBulk(c *gin.Context) {
	var options struct {
		Table     string                   `json:"table"`
		Data      []map[string]interface{} `json:"data"`
		Condition map[string]interface{}   `json:"condition"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Data) == 0 || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Table, data and condition are required"})
		return
	}

	setClause := helpers.UpdateSet(options.Data[0]) // Use first row to determine columns
	whereClause, whereParams := helpers.Where(options.Condition)

	params := []interface{}{}
	for _, row := range options.Data {
		for _, v := range row {
			params = append(params, v)
		}
	}
	params = append(params, whereParams...) // Add condition params

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", options.Table, setClause, whereClause)
	result, err := db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%d rows updated", rowsAffected)})
}

// Delete mysql
func Delete(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Condition map[string]interface{} `json:"condition"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Both table name and condition are required"})
		return
	}
	whereClause, params := helpers.Where(options.Condition)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", options.Table, whereClause)
	result, err := db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%d rows deleted", rowsAffected)})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No rows deleted"})
	}
}

// Count returns the number of rows matching the condition
// DeleteBulk deletes multiple records based on conditions
func DeleteBulk(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Condition map[string]interface{} `json:"condition"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Both table name and condition are required"})
		return
	}
	whereClause, params := helpers.Where(options.Condition)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", options.Table, whereClause)
	result, err := db.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%d rows deleted", rowsAffected)})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No rows deleted"})
	}
}

func Count(c *gin.Context) {
	var options struct {
		Table     string                 `json:"table"`
		Condition map[string]interface{} `json:"condition"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Table == "" || len(options.Condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Both table name and condition are required"})
		return
	}
	whereClause, params := helpers.Where(options.Condition)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", options.Table, whereClause)
	var count int
	if err := db.QueryRow(query, params...).Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Total rows: %d", count)})
}

// Query executes a raw SQL query and returns the results
func Query(c *gin.Context) {
	var options struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request format"})
		return
	}
	if options.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Query cannot be empty"})
		return
	}
	rows, err := db.Query(options.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "No data found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": results})
}
