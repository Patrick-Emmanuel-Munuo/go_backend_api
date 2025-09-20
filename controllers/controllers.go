package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"vartrick/helpers"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
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
	now := time.Now()
	fileName := fmt.Sprintf("mysql_backup_%d.sql", now.Unix())
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
	//helpers.CleanupOldBackups(publicDir, 7*24*time.Hour)
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
			"message": map[string]interface{}{
				"error": "Body can't be empty",
				"data":  nil,
			},
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
		"message": map[string]interface{}{
			"error": "Data not found",
			"data":  data,
		},
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

//bulk count

// Query executes a raw SQL query and returns the results
func Query(options map[string]interface{}) map[string]interface{} {
	// Validate query input
	queryRaw, ok := options["query"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Query is required",
		}
	}
	queryStr, ok := queryRaw.(string)
	if !ok || queryStr == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid query string",
		}
	}
	// Optional: Get query params
	var params []interface{}
	if p, ok := options["params"].([]interface{}); ok {
		params = p
	}
	rows, err := db.Query(queryStr, params...)
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
		row := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			switch v := (*val).(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		results = append(results, row)
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

//
