package controllers

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"vartrick-server/configurations"
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
type ReadResult struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
}
type SMSOptions struct {
	To      []string
	Message string
}

type SMSRecipient struct {
	PhoneNumber  string `json:"phoneNumber"`
	MessageID    string `json:"messageId"`
	Cost         string `json:"cost"`
	Status       string `json:"status"`
	StatusCode   string `json:"statusCode"`
	MessageParts int    `json:"messageParts"`
}
type AfricaTalkingResponse struct {
	SMSMessageData struct {
		Message    string `json:"Message"`
		Recipients []struct {
			Number       string `json:"number"`
			MessageID    string `json:"messageId"`
			Cost         string `json:"cost"`
			Status       string `json:"status"`
			StatusCode   string `json:"statusCode"`
			MessageParts int    `json:"messageParts"`
		} `json:"Recipients"`
	} `json:"SMSMessageData"`
}
type SMSResult struct {
	UniqueID     string         `json:"unique_id"`
	CreatedDate  string         `json:"created_date"`
	CreatedBy    string         `json:"created_by"`
	UpdatedBy    string         `json:"updated_by"`
	UpdatedDate  string         `json:"updated_date"`
	ScheduleTime string         `json:"schedule_time"`
	SenderID     string         `json:"sender_id"`
	Text         string         `json:"text"`
	Summary      string         `json:"summary"`
	TotalCost    string         `json:"totalCost"`
	Recipients   []SMSRecipient `json:"recipients"`
}
type MailOptions struct {
	To          string
	Subject     string
	Message     string
	HTML        string
	Attachments []string
}

type SmsPayload struct {
	Username string `json:"username"`
	To       string `json:"to"`
	Message  string `json:"message"`
	From     string `json:"from,omitempty"`
}

type JsonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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

// sendsms
func SendMessage(options SMSOptions) map[string]interface{} {
	if len(options.To) == 0 || options.Message == "" {
		return map[string]interface{}{"success": false, "message": "both 'to' and 'message' are required and cannot be empty"}
	}
	formData := url.Values{}
	formData.Set("username", os.Getenv("AFRICAS_TALKING_USERNAME"))
	formData.Set("to", strings.Join(options.To, ","))
	formData.Set("message", options.Message)
	formData.Set("from", os.Getenv("AFRICAS_TALKING_SENDER_ID"))
	req, err := http.NewRequest("POST", "https://api.africastalking.com/version1/messaging", strings.NewReader(formData.Encode()))
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("failed to create request: %v", err)}
	}
	// Use form content-type and add API key
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", os.Getenv("AFRICAS_TALKING_API_KEY"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("failed to send request: %v", err)}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{"success": false, "message": "failed to read response body"}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("failed to send SMS: %s", string(body))}
	}
	//fmt.Println("response:", resp)
	// Just return the raw string response
	return map[string]interface{}{"success": true, "message": string(body)}
}

// send main
func SendMail(options MailOptions) map[string]interface{} {
	if options.Message == "" || options.To == "" {
		return map[string]interface{}{"success": false, "message": "both mail to and message required can't be empty"}
	}
	subject := options.Subject
	if subject == "" {
		subject = options.Message
	}
	html := options.HTML
	if html == "" {
		html = fmt.Sprintf("<h2>%s</h2>", options.Message)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", configurations.MailSender)
	m.SetHeader("To", options.To)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", options.Message)
	m.AddAlternative("text/html", html)

	for _, attachment := range options.Attachments {
		m.Attach(attachment)
	}

	d := gomail.NewDialer(configurations.MailHost, configurations.MailPort, configurations.MailUsername, configurations.MailPassword)

	if err := d.DialAndSend(m); err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}

	return map[string]interface{}{"success": true, "message": "Mail sent successfully"}
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
