package controllers

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	"vartrick/helpers"

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

// Generate OTP
func GenerateOTP() map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in GenerateOTP:", r)
		}
	}()
	const otpCharset = "0123456789"
	const otpLength = 6
	numberGroups := otpLength / 2 // Integer division
	otp := make([]byte, otpLength)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpCharset))))
		if err != nil {
			log.Println("Error generating secure OTP:", err)
			return map[string]interface{}{
				"success": false,
				"message": map[string]interface{}{
					"status": "Failed to generate secure OTP",
					"otp":    nil,
				},
			}
		}
		otp[i] = otpCharset[num.Int64()]
	}
	otpFormatted := string(otp[:numberGroups]) + "-" + string(otp[numberGroups:])

	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"status": "OTP generated successfully",
			"otp":    otpFormatted,
		},
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
	// Calculate issue time from BaseDate and tidMinutes
	issueTime := helpers.BaseDate.Add(time.Duration(tidMinutes) * time.Minute)

	// Calculate expiry time (1 year after issue time)
	expiryTime := issueTime.AddDate(1, 0, 0)

	// Check if the token is expired
	if timeNow.After(expiryTime) {
		return map[string]interface{}{
			"success": false,
			"message": "Token issue date has expired",
		}
	}
	// Check if the issue time is invalid (before BaseDate or too far in the future)
	if issueTime.Before(helpers.BaseDate) || issueTime.After(timeNow.Add(24*time.Hour)) {
		return map[string]interface{}{
			"success": false,
			"message": "update Change meter base date for",
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

// encript token
func EncriptToken(options map[string]interface{}) map[string]interface{} {
	// Validate required fields
	amountRaw, ok := options["amount"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Amount field is required.",
		}
	}
	// First, assert that amountRaw is a string
	amountStr, ok := amountRaw.(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Amount must be a string.",
		}
	}
	// Then parse the string to float64
	amountNumber, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Amount must be a valid number.",
		}
	}
	// Use amountNumber as needed
	amount := math.Floor(amountNumber*100) / 100
	// Use current time or provided time
	issueTime := time.Now()
	if t, ok := options["issued_time"]; ok {
		if tStr, ok := t.(string); ok {
			parsedTime, err := time.Parse(time.RFC3339, tStr)
			if err == nil {
				issueTime = parsedTime
			}
		}
	}
	// Calculate TID (minutes since base date)
	tidMinutes := int64(issueTime.Sub(helpers.BaseDate).Minutes())
	tidBin := fmt.Sprintf("%022b", tidMinutes)
	// Encode units (amount → binary string of 23 bits)
	amtRes := helpers.EncodeUnits(amount)
	if !amtRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to encode units: " + amtRes["message"].(string),
		}
	}
	amtBlock := amtRes["message"].(string)
	// Generate 3-bit random block
	randRes := helpers.GenerateRandomBits(3)
	if !randRes["success"].(bool) {
		// handle error here
	}
	randomBits := randRes["message"].(string)
	// Construct binary string: random(3) + tid(22) + amount(23) = 48 bits
	dataBin := randomBits + tidBin + amtBlock
	// Convert to hex for CRC
	dataHex := helpers.BinToHex(dataBin)
	if len(dataHex) < 14 {
		dataHex = fmt.Sprintf("%014s", dataHex)
	}
	dataBytes, err := helpers.HexToBytes(dataHex)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to convert data to bytes: " + err.Error(),
		}
	}
	// Calculate CRC16 on first 48 bits
	crcRes := helpers.CalculateCRC16(dataBytes)
	if !crcRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "CRC16 calculation failed: " + crcRes["message"].(string),
		}
	}
	crcBin := crcRes["message"].(string)
	if len(crcBin) < 16 {
		crcBin = fmt.Sprintf("%016s", crcBin)
	}
	// Full binary block (64-bit)
	fullBin := dataBin + crcBin
	// Convert to byte array
	fullBytes := helpers.BinStrToBytes(fullBin)
	// Generate decoder key
	keyRes := helpers.GenerateDecoderKey()
	if !keyRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to generate decoder key: " + keyRes["message"].(string),
		}
	}
	keyBin := keyRes["message"].(string)
	keyBytes := helpers.BinStrToBytes(keyBin)
	// Encrypt with 3DES
	encRes := helpers.Encrypt3DES(fullBytes, keyBytes)
	if !encRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Encryption failed: " + encRes["message"].(string),
		}
	}
	encBytes := encRes["message"].([]byte)
	// Convert encrypted bytes to binary string
	encBin := helpers.BytesToBinStr(encBytes)
	if len(encBin) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Encrypted binary is less than 64 bits.",
		}
	}

	// Add class bits (assume class provided in options or default to 0)
	classInt := 0
	if c, ok := options["class"]; ok {
		if cInt, ok := c.(int); ok {
			classInt = cInt
		} else if cFloat, ok := c.(float64); ok {
			classInt = int(cFloat)
		}
	}
	classBitsRes := helpers.GenerateClassBits(classInt)
	if !classBitsRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to generate class bits: " + classBitsRes["message"].(string),
		}
	}
	classBits := classBitsRes["message"].(string)

	// Apply transposition
	transpositionRes := helpers.TranspositionAndAddClassBits(encBin, classBits)
	if !transpositionRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed in transposition: " + transpositionRes["message"].(string),
		}
	}
	finalBin := transpositionRes["message"].(string)
	// Format for display (e.g. spaced groups)
	tokenResponce := helpers.FormatTokenDisplay(finalBin)
	if !tokenResponce["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to format token display: " + tokenResponce["message"].(string),
		}
	}
	token := tokenResponce["message"].(string)
	// Return response
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			//"token":         tokenStr,
			"token":            token,
			"issued_date":      issueTime.Format(time.RFC3339),
			"expired_datetime": issueTime.AddDate(1, 0, 0).Format(time.RFC3339),
			"identifier":       tidMinutes,
			"units":            amount,
			"random_bits":      randomBits,
			"class_bits":       classBits,
			"crc_block":        crcBin,
		},
	}
}

// SendMessage sends an SMS using Africa's Talking API
func SendMessage(options map[string]interface{}) map[string]interface{} {
	// Validate token field
	message := options["message"].(string)
	toRaw, toOk := options["to"]
	toSlice, ok := toRaw.([]string)
	if !toOk || !ok || len(toSlice) == 0 || strings.TrimSpace(message) == "" {
		return map[string]interface{}{
			"Success": false,
			"Message": "both 'to' and 'message' are required and must be valid",
		}
	}
	formData := url.Values{}
	formData.Set("username", os.Getenv("AFRICAS_TALKING_USERNAME"))
	formData.Set("to", strings.Join(toSlice, ","))
	formData.Set("message", message)
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
		"": map[string]interface{}{
			"status": "",
			"data":   parsedResponse,
		},
	}
}

func SendMail(options map[string]interface{}) map[string]interface{} {
	// Validate required fields
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	msg, msgOk := options["message"].(string)
	toRaw, toOk := options["to"]
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
	if helpers.Mailsender == "" || helpers.Mailhost == "" || helpers.Mailusername == "" || helpers.Mailpassword == "" || helpers.Mailport == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Missing required environment variables for SMTP configuration.",
		}
	}
	// Compose the message
	m := gomail.NewMessage()
	m.SetHeader("From", helpers.Mailsender)
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
	//d := gomail.NewDialer(host, port, username, password)
	d := gomail.NewDialer(helpers.Mailhost, helpers.Mailport, helpers.Mailusername, helpers.Mailpassword)

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
func Backup(options map[string]interface{}) map[string]interface{} {
	email, emailOk := options["email"].(string)
	if !emailOk || email == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Email is required and must be a string.",
		}
	}
	timeNow := time.Now()
	fileName := fmt.Sprintf("mysql_backup_%d.sql", timeNow.Unix())
	publicDir := filepath.Join(".", "public")
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		log.Println("Failed to create public dir:", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	filePath := filepath.Join(publicDir, fileName)
	if helpers.DatabasePassword == "" {
		log.Println("MYSQL_PASSWORD not set in environment or helpers")
		return map[string]interface{}{
			"success": false,
			"message": "MySQL password not configured.",
		}
	}
	//cmd := exec.Command("mysqldump", "-h", helpers.DatabaseHost, "-u", helpers.DatabaseUser, "-p"+helpers.DatabasePassword, helpers.DatabaseName)
	cmd := exec.Command("mysqldump",
		"-h", helpers.DatabaseHost,
		"-u", helpers.DatabaseUser,
		"-p"+helpers.DatabasePassword,
		helpers.DatabaseName,
	)
	outfile, err := os.Create(filePath)
	if err != nil {
		log.Println("Failed to create dump file:", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	if err := cmd.Run(); err != nil {
		log.Println("mysqldump failed:", err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	// Send backup via email
	response := SendMail(map[string]interface{}{
		"to":          email,
		"message":     "Email for backup database",
		"Attachments": filePath,
	})
	if success, ok := response["success"].(bool); !ok || !success {
		return map[string]interface{}{
			"success": false,
			"message": map[string]interface{}{
				"status": "Backup created successfully but failed to send email",
				"error":  response["message"],
			},
		}
	}
	// Optional: Cleanup old backups older than 7 days
	helpers.CleanupOldBackups(publicDir, 7*24*time.Hour)
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"status": "Backup created and sent successfully to " + email,
			"error":  "",
		},
	}
}

// read mysql
func Read(options map[string]interface{}) map[string]interface{} {
	// Extract and validate required fields
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing or invalid table name",
		}
	}

	// Optional: SELECT fields
	selectFields := "*"
	if selectSlice, ok := options["select"].([]interface{}); ok && len(selectSlice) > 0 {
		fields := make([]string, len(selectSlice))
		for i, f := range selectSlice {
			fields[i], _ = f.(string)
		}
		selectFields = helpers.JoinFields(fields)
	}

	// Parse "condition" and "or_condition"
	condition := make(map[string]interface{})
	orCondition := make(map[string]interface{})
	if cond, ok := options["condition"].(map[string]interface{}); ok {
		condition = cond
	}
	if orCond, ok := options["or_condition"].(map[string]interface{}); ok {
		orCondition = orCond
	}

	if len(condition) == 0 && len(orCondition) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Missing condition(s)",
		}
	}

	// Build WHERE clause
	var whereClause string
	var params []interface{}
	if len(condition) > 0 && len(orCondition) > 0 {
		where1, params1 := helpers.Where(condition)
		where2, params2 := helpers.WhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	} else if len(condition) > 0 {
		whereClause, params = helpers.Where(condition)
	} else {
		whereClause, params = helpers.WhereOr(orCondition)
	}

	// Build and execute query
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// bulk read
func ReadBulk(options []map[string]interface{}) map[string]interface{} {
	var (
		data      []map[string]interface{}
		errorData []map[string]interface{}
	)

	if len(options) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Body can't be empty",
		}
	}

	for _, opt := range options {
		tableName := ""
		if t, ok := opt["table"].(string); ok {
			tableName = t
		}
		result := Read(opt)

		entry := map[string]interface{}{
			"table":   tableName,
			"message": result["message"],
		}

		if result["success"] == true {
			data = append(data, entry)
		} else {
			errorData = append(errorData, entry)
		}
	}

	// Determine the final status based on results
	if len(errorData) == 0 && len(data) > 0 {
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"error": errorData,
				"data":  data,
			},
		}
	} else if len(errorData) > 0 && len(data) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": map[string]interface{}{
				"error": errorData,
				"data":  data,
			},
		}
	} else if len(errorData) > 0 && len(data) > 0 {
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"error": errorData,
				"data":  data,
			},
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "Data not found",
	}
}

// search mysql
func Search(options map[string]interface{}) map[string]interface{} {
	// Validate table name
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing or invalid table name",
		}
	}

	// Parse select fields
	selectFields := "*"
	if selectVal, ok := options["select"].([]interface{}); ok && len(selectVal) > 0 {
		fields := make([]string, len(selectVal))
		for i, f := range selectVal {
			fields[i], _ = f.(string)
		}
		selectFields = helpers.JoinFields(fields)
	}

	// Parse condition
	condition := make(map[string]interface{})
	if cond, ok := options["condition"].(map[string]interface{}); ok {
		condition = cond
	}
	if len(condition) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Missing search condition(s)",
		}
	}

	// Build LIKE where clause and params
	whereClause, params := helpers.Like(condition)
	if whereClause == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid search condition(s)",
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}
		results = append(results, rowMap)
	}

	// Convert []uint8 to string for JSON compatibility
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// SearchBetween finds records where a field's value is between two values
func SearchBetween(options map[string]interface{}) map[string]interface{} {
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing or invalid table name",
		}
	}
	field, ok := options["field"].(string)
	if !ok || field == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing or invalid field name",
		}
	}
	start, startExists := options["start"]
	end, endExists := options["end"]
	if !startExists || !endExists {
		return map[string]interface{}{
			"success": false,
			"message": "Start and end values are required",
		}
	}
	condition := map[string]interface{}{}
	if condRaw, exists := options["condition"]; exists {
		if condMap, ok := condRaw.(map[string]interface{}); ok {
			condition = condMap
		}
	}
	whereClause := fmt.Sprintf("%s BETWEEN ? AND ?", field)
	params := []interface{}{start, end}
	if len(condition) > 0 {
		additionalWhere, additionalParams := helpers.Where(condition)
		whereClause += " AND " + additionalWhere
		params = append(params, additionalParams...)
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// List returns paginated results from a table
func List(options map[string]interface{}) map[string]interface{} {
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}
	pageFloat, ok := options["page"].(float64) // JSON numbers decode to float64
	if !ok || pageFloat <= 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Page number must be a positive integer",
		}
	}
	page := int(pageFloat)
	pageSizeFloat, ok := options["page_size"].(float64)
	if !ok || pageSizeFloat <= 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Page size must be a positive integer",
		}
	}
	pageSize := int(pageSizeFloat)
	offset := (page - 1) * pageSize
	// Extract condition if provided, expect a map[string]interface{}
	condition := map[string]interface{}{}
	if condRaw, exists := options["condition"]; exists {
		if condMap, ok := condRaw.(map[string]interface{}); ok {
			condition = condMap
		}
	}
	whereClause, params := helpers.Where(condition)
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT ? OFFSET ?", table, whereClause)
	params = append(params, pageSize, offset)
	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// ListAll returns all records from a table
func ListAll(options map[string]interface{}) map[string]interface{} {
	tableRaw, ok := options["table"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}
	table, ok := tableRaw.(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid table name",
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", table)
	rows, err := db.Query(query)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// create mysql
func Create(options map[string]interface{}) map[string]interface{} {
	tableRaw, ok := options["table"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}
	table, ok := tableRaw.(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid table name",
		}
	}

	dataRaw, ok := options["data"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Data is required",
		}
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok || len(data) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid data format",
		}
	}

	// Generate unique_id and add to data
	uniqueID := helpers.GenerateUniqueID()
	data["unique_id"] = uniqueID

	// Build query: INSERT INTO table SET col1=?, col2=?, ...
	columns := []string{}
	values := []interface{}{}

	for col, val := range data {
		columns = append(columns, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}
	query := fmt.Sprintf("INSERT INTO %s SET %s", table, strings.Join(columns, ", "))

	// Execute query
	result, err := db.Exec(query, values...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	if rowsAffected > 0 {
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"unique_id": uniqueID,
				"data":      data,
			},
		}
	}

	return map[string]interface{}{
		"success": false,
		"message": "Failed to insert data",
	}
}

// CreateBulk inserts multiple records in a single query
func CreateBulk(options map[string]interface{}) map[string]interface{} {
	dataCreate := []map[string]interface{}{}
	errorCreate := []map[string]interface{}{}

	// Validate options
	dataRaw, dataExists := options["data"]
	tableRaw, tableExists := options["table"]

	if !dataExists || !tableExists {
		return map[string]interface{}{
			"success": false,
			"message": "Missing required fields: data and table",
		}
	}

	dataSlice, ok := dataRaw.([]interface{})
	table, ok2 := tableRaw.(string)
	if !ok || !ok2 || len(table) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid data or table format",
		}
	}

	if len(dataSlice) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Body can't be empty",
		}
	}

	// Loop through each data item and call Create
	for _, item := range dataSlice {
		dataMap, ok := item.(map[string]interface{})
		if !ok {
			errorCreate = append(errorCreate, map[string]interface{}{
				"success": false,
				"message": "Invalid data item format",
			})
			continue
		}

		result := Create(map[string]interface{}{
			"table": table,
			"data":  dataMap,
		})

		if success, ok := result["success"].(bool); ok && success {
			dataCreate = append(dataCreate, result)
		} else {
			errorCreate = append(errorCreate, result)
		}
	}

	// Build response
	switch {
	case len(errorCreate) == 0 && len(dataCreate) > 0:
		return map[string]interface{}{
			"success": true,
			"message": dataCreate,
		}
	case len(errorCreate) > 0 && len(dataCreate) == 0:
		return map[string]interface{}{
			"success": false,
			"message": errorCreate,
		}
	case len(errorCreate) > 0 && len(dataCreate) > 0:
		return map[string]interface{}{
			"success": false,
			"message": append(errorCreate, dataCreate...),
		}
	default:
		return map[string]interface{}{
			"success": false,
			"message": "Data not found",
		}
	}
}

// update mysql
func Update(options map[string]interface{}) map[string]interface{} {
	// Extract and validate table name
	tableRaw, ok := options["table"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}
	table, ok := tableRaw.(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid table name",
		}
	}
	// Extract and validate data
	dataRaw, ok := options["data"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Data is required",
		}
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok || len(data) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid data format",
		}
	}
	// Extract and validate condition
	condRaw, ok := options["condition"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Condition is required",
		}
	}
	condition, ok := condRaw.(map[string]interface{})
	if !ok || len(condition) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid condition format",
		}
	}
	// Build SET clause and WHERE clause
	setClause := helpers.UpdateSet(data)
	whereClause, whereParams := helpers.Where(condition)
	// Build parameters in correct order: data values first, then condition values
	var params []interface{}
	for _, key := range helpers.SortedKeys(data) {
		params = append(params, data[key])
	}
	params = append(params, whereParams...)
	// Construct final SQL query
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setClause, whereClause)
	// Execute query
	result, err := db.Exec(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error()}
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return map[string]interface{}{
			"success": true,
			"message": "Data updated successfully", "rows_affected": rowsAffected}
	}
	return map[string]interface{}{
		"success": false,
		"message": "No data was updated. Condition may not match or data is unchanged"}
}

// UpdateBulk updates multiple records in a single query
func UpdateBulk(options []map[string]interface{}) map[string]interface{} {
	dataUpdate := []map[string]interface{}{}
	errorUpdate := []map[string]interface{}{}

	if options == nil {
		return map[string]interface{}{
			"success": false,
			"message": "function options parameter required can't be empty",
		}
	}

	if len(options) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "body can't be empty",
		}
	}

	for _, opt := range options {
		result := Update(opt)
		if success, ok := result["success"].(bool); ok && success {
			dataUpdate = append(dataUpdate, result)
		} else {
			errorUpdate = append(errorUpdate, result)
		}
	}

	switch {
	case len(errorUpdate) == 0 && len(dataUpdate) > 0:
		return map[string]interface{}{
			"success": true,
			"message": dataUpdate,
		}
	case len(errorUpdate) > 0 && len(dataUpdate) == 0:
		return map[string]interface{}{
			"success": false,
			"message": errorUpdate,
		}
	case len(errorUpdate) > 0 && len(dataUpdate) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(errorUpdate, dataUpdate...),
		}
	default:
		return map[string]interface{}{
			"success": false,
			"message": "data not found",
		}
	}
}

// Delete mysql
func Delete(options map[string]interface{}) map[string]interface{} {
	// Validate table name
	tableRaw, ok := options["table"]
	if !ok {
		return map[string]interface{}{"success": false, "message": "Table name is required"}
	}
	table, ok := tableRaw.(string)
	if !ok || table == "" {
		return map[string]interface{}{"success": false, "message": "Invalid table name"}
	}

	// Validate condition
	condRaw, ok := options["condition"]
	if !ok {
		return map[string]interface{}{"success": false, "message": "Condition is required"}
	}
	condition, ok := condRaw.(map[string]interface{})
	if !ok || len(condition) == 0 {
		return map[string]interface{}{"success": false, "message": "Invalid condition format"}
	}

	// Build WHERE clause
	whereClause, whereParams := helpers.Where(condition)

	// Final SQL delete query
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereClause)

	// Execute query
	result, err := db.Exec(query, whereParams...)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return map[string]interface{}{"success": true, "message": "Data deleted successfully", "rows_affected": rowsAffected}
	}

	return map[string]interface{}{"success": false, "message": "No data was deleted. Condition may not match"}
}

// DeleteBulk deletes multiple records based on conditions
func DelateBulk(options []map[string]interface{}) map[string]interface{} {
	var (
		dataDeleted  []map[string]interface{}
		errorDeleted []map[string]interface{}
	)

	if len(options) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Body can't be empty",
		}
	}

	for _, opt := range options {
		result := Delete(opt)

		if success, ok := result["success"].(bool); ok && success {
			dataDeleted = append(dataDeleted, result)
		} else {
			errorDeleted = append(errorDeleted, result)
		}
	}

	// Build response
	switch {
	case len(errorDeleted) == 0 && len(dataDeleted) > 0:
		return map[string]interface{}{
			"success": true,
			"message": dataDeleted,
		}
	case len(errorDeleted) > 0 && len(dataDeleted) == 0:
		return map[string]interface{}{
			"success": false,
			"message": errorDeleted,
		}
	case len(errorDeleted) > 0 && len(dataDeleted) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(errorDeleted, dataDeleted...),
		}
	default:
		return map[string]interface{}{
			"success": false,
			"message": "Data not found",
		}
	}
}

// Count returns the number of rows matching the condition
func Count(options map[string]interface{}) map[string]interface{} {
	// Validate table name
	tableRaw, ok := options["table"]
	if !ok {
		return map[string]interface{}{"success": false, "message": "Table name is required"}
	}
	table, ok := tableRaw.(string)
	if !ok || table == "" {
		return map[string]interface{}{"success": false, "message": "Invalid table name"}
	}

	// Optional condition
	var whereClause string
	var params []interface{}

	if condRaw, ok := options["condition"]; ok {
		if condition, ok := condRaw.(map[string]interface{}); ok && len(condition) > 0 {
			whereClause, params = helpers.Where(condition)
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	// Execute query
	var count int
	err := db.QueryRow(query, params...).Scan(&count)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	return map[string]interface{}{
		"success": true,
		"count":   count,
	}
}

// Query executes a raw SQL query and returns the results
func Query(options map[string]interface{}) map[string]interface{} {
	// Validate query name
	query_data, ok := options["query"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "query name is required",
		}
	}
	query_name, ok := query_data.(string)
	if !ok || query_name == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid query name",
		}
	}
	var params []interface{}

	// Build query
	query := fmt.Sprintf(query_name)

	// Execute query
	//db.Query(options.Query)
	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
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
		return map[string]interface{}{
			"success": false,
			"message": "No data found",
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

/*
// file hanle
var uploadPath = "./public"

func UploadFile(c *gin.Context) map[string]interface{} {
	file, err := c.FormFile("file")
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "File upload error",
		}
	}

	os.MkdirAll(uploadPath, 0755)
	dst := filepath.Join(uploadPath, filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Unable to save file",
		}
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "File uploaded",
		"filename": file.Filename,
	}
}

func UploadMultipleFiles(c *gin.Context) map[string]interface{} {
	form, err := c.MultipartForm()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid form",
		}
	}

	files := form.File["files"]
	os.MkdirAll(uploadPath, 0755)

	var uploaded []string
	for _, file := range files {
		dst := filepath.Join(uploadPath, filepath.Base(file.Filename))
		if err := c.SaveUploadedFile(file, dst); err != nil {
			continue
		}
		uploaded = append(uploaded, file.Filename)
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "Files uploaded",
		"uploaded": uploaded,
	}
}


func DeleteFile(c *gin.Context) map[string]interface{} {
	filename := c.Param("filename")
	fullPath := filepath.Join(uploadPath, filename)
	if err := os.Remove(fullPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to delete file",
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": "File deleted",
	}
}

*/
