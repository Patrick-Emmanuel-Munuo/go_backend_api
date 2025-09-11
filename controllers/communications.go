package controllers

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"vartrick/helpers"

	serial "go.bug.st/serial.v1"
	"gopkg.in/gomail.v2"
)

// Generate OTP
func GenerateOTP(options map[string]interface{}) map[string]interface{} {
	// Default OTP length
	otpLength := 6
	// Check if custom length is provided and valid
	if lenVal, ok := options["length"]; ok {
		switch v := lenVal.(type) {
		case int:
			if v > 0 {
				otpLength = v
			}
		case string:
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				otpLength = parsed
			}
		}
	}
	const otpCharset = "0123456789"
	otp := make([]byte, otpLength)

	for i := 0; i < otpLength; i++ {
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
	// Format: Split in half if even length; otherwise just return as is
	var otpFormatted string
	otpFormatted = string(otp)
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"status": "OTP generated successfully",
			"otp":    otpFormatted,
		},
	}
}

// send OTP via either mail or phone
func SendOTP(options map[string]interface{}) map[string]interface{} {
	otpLength := 4
	// Handle OTP length
	if lenVal, ok := options["length"]; ok {
		switch v := lenVal.(type) {
		case int:
			if v > 0 {
				otpLength = v
			}
		case string:
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				otpLength = parsed
			}
		}
	}
	// Generate OTP
	result := GenerateOTP(map[string]interface{}{
		"length": otpLength,
	})
	// Check if generation was successful
	if success, ok := result["success"].(bool); ok && success {
		msgData, ok := result["message"].(map[string]interface{})
		if !ok {
			return map[string]interface{}{
				"success": false,
				"message": "OTP generation response format error",
				"error":   result,
			}
		}
		otpCode, ok := msgData["otp"].(string)
		message := fmt.Sprintf(
			"Dear user,\n\nYour One-Time Password (OTP) is: %s\n\n"+
				"Please use this code to complete your verification. "+
				"Thank you,",
			otpCode,
		)
		if !ok {
			return map[string]interface{}{
				"success": false,
				"message": "OTP value missing",
				"error":   result,
			}
		}
		var emailStatus, phoneStatus map[string]interface{}
		// Send via email
		if emailVal, ok := options["email"]; ok {
			if email, ok := emailVal.(string); ok && email != "" {
				emailStatus = SendMail(map[string]interface{}{
					"to":      email,
					"message": message,
					"subject": "Your One-Time Password (OTP)",
				})
			} else {
				emailStatus = map[string]interface{}{
					"success": false,
					"message": "Invalid email format",
				}
			}
		} else {
			emailStatus = map[string]interface{}{
				"success": false,
				"message": "Email not provided",
			}
		}
		// Send via phone
		if phoneVal, ok := options["phone"]; ok {
			if phone, ok := phoneVal.(string); ok && phone != "" {
				phoneStatus = SendMessage(map[string]interface{}{
					"to":      phone,
					"message": message,
				})
			} else {
				phoneStatus = map[string]interface{}{
					"success": false,
					"message": "Invalid phone format",
				}
			}
		} else {
			phoneStatus = map[string]interface{}{
				"success": false,
				"message": "Phone number not provided",
			}
		}
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"status": "OTP generated and sent successfully",
				"otp":    message,
				"email":  emailStatus,
				"phone":  phoneStatus,
			},
		}
	}
	// OTP generation failed
	return map[string]interface{}{
		"success": false,
		"message": "Failed to generate or send OTP",
		"error":   result,
	}
}

// SendMail sends an email using SMTP with gomail
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
	subject, _ := options["subject"].(string)
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

// SendMessage sends an SMS using Africa's Talking API
func SendMessage(options map[string]interface{}) map[string]interface{} {
	// Validate token field
	message := options["message"].(string)
	toRaw, toOk := options["to"]
	toSlice, ok := toRaw.([]string)
	if !toOk || !ok || len(toSlice) == 0 || strings.TrimSpace(message) == "" {
		return map[string]interface{}{
			"success": false,
			"message": "both 'to' and 'message' are required and must be valid",
		}
	}
	formData := url.Values{}
	formData.Set("username", helpers.SmsUserName)
	formData.Set("to", strings.Join(toSlice, ","))
	formData.Set("message", message)
	formData.Set("from", helpers.SmsSenderId)

	req, err := http.NewRequest("POST", "https://api.africastalking.com/version1/messaging", strings.NewReader(formData.Encode()))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", helpers.SmsApiKey)

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
			"success": false,
			"message": userMessage,
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("SMS send failed with status %d. Response body: %s\n", resp.StatusCode, string(bodyBytes))
		return map[string]interface{}{
			"success": false,
			"message": "SMS send failed. Please try again later.",
		}
	}
	// Use helper function to convert XML response body to JSON map
	result := helpers.XMLtoJSON(resp)
	if success, ok := result["success"].(bool); ok && success {
		data := result["message"].(map[string]interface{})
		results := data["AfricasTalkingResponse"].(map[string]interface{})["SMSMessageData"].(map[string]interface{})
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"status":     results["Message"],
				"recipients": results["Recipients"].(map[string]interface{})["Recipient"],
			},
		}
	} else {
		return map[string]interface{}{
			"success": false,
			"message": result["message"],
		}
	}
}

// SendMessageLocal sends an SMS using a local GSM modem via serial port
func SendMessageLocal(options map[string]interface{}) map[string]interface{} {
	messageRaw, messageOk := options["message"].(string)
	toRaw, toOk := options["to"]
	toSlice, ok := toRaw.([]string)
	if !messageOk || !toOk || !ok || len(toSlice) == 0 || strings.TrimSpace(messageRaw) == "" {
		return map[string]interface{}{
			"success": false,
			"message": "both 'to' and 'message' are required and must be valid",
		}
	}
	recipients := []map[string]interface{}{}
	sentCount := 0
	totalCost := 0.0 // pretend cost calculation

	portName := "COM8" //"/dev/ttyUSB0" // or "COM3" on Windows
	baudRate := 9600

	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatalf("Error listing serial ports: %v", err)
		}
		return map[string]interface{}{
			"success": false,
			"ports":   ports,
			"message": fmt.Sprintf("failed to open port %s: %v", portName, err),
		}
	}
	defer port.Close()

	_, err = port.Write([]byte("AT+CMGF=1\r"))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("failed to set text mode: %v", err),
		}
	}
	time.Sleep(50 * time.Millisecond)

	statusMap := make(map[string]string)

	for _, phone := range toSlice {
		_, err := port.Write([]byte(fmt.Sprintf("AT+CMGS=\"%s\"\r", phone)))
		if err != nil {
			statusMap[phone] = "failed to initiate send command"
			continue
		}
		time.Sleep(50 * time.Millisecond)

		_, err = port.Write([]byte(messageRaw + "\r"))
		if err != nil {
			statusMap[phone] = "failed to write message text"
			continue
		}
		time.Sleep(50 * time.Millisecond)

		_, err = port.Write([]byte{26})
		if err != nil {
			statusMap[phone] = "failed to send message terminator"
			continue
		}
		time.Sleep(50 * time.Millisecond)
		statusMap[phone] = "message sent"
		//log.Printf("Sent SMS to %s", phone)
		sentCount++
	}

	for _, phone := range toSlice {
		status, exists := statusMap[phone]
		statusCode := "200"
		if !exists || status != "message sent" {
			status = "InvalidPhoneNumber"
			statusCode = "403"
		}

		recipient := map[string]interface{}{
			"cost":         "0",
			"message_id":   helpers.GenerateUniqueID(), // your unique ID func
			"messageParts": "0",
			"number":       phone,
			"message":      messageRaw,
			"method":       "local_via_moderm",
			"status":       status,
			"sender_id":    "+255760449295",
			"statusCode":   statusCode,
		}
		recipients = append(recipients, recipient)
	}
	statusSummary := fmt.Sprintf("Sent to %d/%d Total Cost: %.2f", sentCount, len(toSlice), totalCost)
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"status":     statusSummary,
			"recipients": recipients,
		},
	}
}

// ReadMessageLocal reads SMS messages from a local GSM modem via serial port
func ReadMessageLocal() map[string]interface{} {
	portName := "COM8"
	baudRate := 9600
	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to open port %s: %v", portName, err),
		}
	}
	defer port.Close()

	// Set text mode
	_, err = port.Write([]byte("AT+CMGF=1\r"))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to set text mode: %v", err),
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Request all messages
	_, err = port.Write([]byte("AT+CNMI=1,2,0,0,0\r"))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to request messages: %v", err),
		}
	}
	time.Sleep(500 * time.Millisecond)

	type readResult struct {
		n   int
		err error
		buf []byte
	}
	buffer := make([]byte, 1024)
	fullResponse := make([]byte, 0)

readLoop:
	for {
		resultChan := make(chan readResult)
		go func() {
			n, err := port.Read(buffer)
			resultChan <- readResult{n: n, err: err, buf: buffer[:n]}
		}()
		select {
		case res := <-resultChan:
			if res.err != nil {
				break readLoop
			}
			if res.n == 0 {
				break readLoop
			}
			fullResponse = append(fullResponse, res.buf...)
			if res.n < len(buffer) {
				break readLoop
			}
		case <-time.After(300 * time.Millisecond):
			break readLoop
		}
	}
	responseStr := string(fullResponse)
	parsedMessages := helpers.ParseMessages(responseStr)
	if parsedMessages["success"].(bool) {
		return map[string]interface{}{
			"success": true,
			"message": parsedMessages["message"],
		}
	}
	return map[string]interface{}{
		"success": false,
		"message": parsedMessages["message"],
	}
}
