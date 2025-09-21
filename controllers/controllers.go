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
		selectFields = helpers.GenerateSelect(fields)
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
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	} else if len(condition) > 0 {
		whereClause, params = helpers.GenerateWhere(condition)
	} else {
		whereClause, params = helpers.GenerateWhereOr(orCondition)
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
	readed := []map[string]interface{}{}
	failed := []map[string]interface{}{}

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
		result := Read(opt)
		if success, ok := result["success"].(bool); ok && success {
			readed = append(readed, result)
		} else {
			failed = append(failed, result)
		}
	}

	switch {
	case len(failed) == 0 && len(readed) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(failed, readed...),
		}
	case len(failed) > 0 && len(readed) == 0:
		return map[string]interface{}{
			"success": false,
			"message": append(failed, readed...),
		}
	case len(failed) > 0 && len(readed) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(failed, readed...),
		}
	default:
		return map[string]interface{}{
			"success": false,
			"message": "data not found",
		}
	}
}

// Count returns the number of rows matching the condition
func Count(options map[string]interface{}) map[string]interface{} {
	// Validate table name
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}

	// Validate condition
	condition, ok := options["condition"].(map[string]interface{})
	if !ok || len(condition) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Condition is required",
		}
	}

	// Generate WHERE clause using helper
	whereClause, params := helpers.GenerateWhere(condition)

	// Prepare query
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s WHERE %s", helpers.EscapeId(table), whereClause)

	// Execute query
	var total int
	err := db.QueryRow(query, params...).Scan(&total)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": total,
	}
}

// CountBulk runs multiple count queries
func CountBulk(options []map[string]interface{}) map[string]interface{} {
	if len(options) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Options must be a non-empty array",
		}
	}

	results := []map[string]interface{}{}

	for _, opt := range options {
		table, ok1 := opt["table"].(string)
		condition, ok2 := opt["condition"].(map[string]interface{})

		if !ok1 || table == "" || !ok2 || len(condition) == 0 {
			results = append(results, map[string]interface{}{
				"table":   table,
				"success": false,
				"count":   0,
				"message": "Table name and condition required",
			})
			continue
		}

		res := Count(map[string]interface{}{
			"table":     table,
			"condition": condition,
		})

		if res["success"].(bool) {
			results = append(results, map[string]interface{}{
				"table":   table,
				"success": true,
				"count":   res["message"],
				"message": "Count retrieved",
			})
		} else {
			results = append(results, map[string]interface{}{
				"table":   table,
				"success": false,
				"count":   0,
				"message": res["message"],
			})
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// Search executes a SELECT query with AND/OR conditions
func Search(options map[string]interface{}) map[string]interface{} {
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}

	condition, _ := options["condition"].(map[string]interface{})
	orCondition, _ := options["or_condition"].(map[string]interface{})

	if len(condition) == 0 && len(orCondition) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "At least one of 'condition' or 'or_condition' is required",
		}
	}

	// Handle SELECT fields
	selectFields := "*"
	if sel, ok := options["select"].(map[string]interface{}); ok && len(sel) > 0 {
		keys := make([]string, 0, len(sel))
		for k := range sel {
			keys = append(keys, k)
		}
		selectFields = strings.Join(keys, ", ")
	}

	// Build WHERE clause
	var whereClause string
	var params []interface{}
	if len(condition) > 0 && len(orCondition) > 0 {
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	} else if len(condition) > 0 {
		whereClause, params = helpers.GenerateWhere(condition)
	} else {
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, table, whereClause)

	rows, err := db.Query(query)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		// make slice for Scan
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		rows.Scan(valPtrs...)

		rowMap := map[string]interface{}{}
		for i, col := range cols {
			rowMap[col] = vals[i]
		}
		results = append(results, rowMap)
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "No matching data found.",
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": results,
	}
}

// SearchBetween executes SELECT with BETWEEN and optional conditions
func SearchBetween(options map[string]interface{}) map[string]interface{} {
	table, ok1 := options["table"].(string)
	column, ok2 := options["column"].(string)
	start, ok3 := options["start"]
	end, ok4 := options["end"]

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return map[string]interface{}{
			"success": false,
			"message": "Parameters 'table', 'column', 'start', and 'end' are required.",
		}
	}

	// Handle SELECT fields
	selectFields := "*"
	if sel, ok := options["select"].(map[string]interface{}); ok && len(sel) > 0 {
		keys := make([]string, 0, len(sel))
		for k := range sel {
			keys = append(keys, k)
		}
		selectFields = strings.Join(keys, ", ")
	}

	// Base query and params
	params := []interface{}{start, end}
	var whereClause string

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
	if len(condition) > 0 && len(orCondition) > 0 {
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	} else if len(condition) > 0 {
		whereClause, params = helpers.GenerateWhere(condition)
	} else {
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s BETWEEN ? AND ? AND %s", selectFields, table, column, whereClause)

	rows, err := db.Query(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		rows.Scan(valPtrs...)

		rowMap := map[string]interface{}{}
		for i, col := range cols {
			rowMap[col] = vals[i]
		}
		results = append(results, rowMap)
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "No records found in the given range.",
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
	whereClause, params := helpers.GenerateWhere(condition)
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
	setClause := helpers.GenerateSet(data)
	whereClause, whereParams := helpers.GenerateWhere(condition)
	// Build parameters in correct order: data values first, then condition values
	var params []interface{}
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
	updated := []map[string]interface{}{}
	failed := []map[string]interface{}{}

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
			updated = append(updated, result)
		} else {
			failed = append(failed, result)
		}
	}

	switch {
	case len(failed) == 0 && len(updated) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(failed, updated...),
		}
	case len(failed) > 0 && len(updated) == 0:
		return map[string]interface{}{
			"success": false,
			"message": append(failed, updated...),
		}
	case len(failed) > 0 && len(updated) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(failed, updated...),
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
	whereClause, whereParams := helpers.GenerateWhere(condition)

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
		deleted []map[string]interface{}
		failed  []map[string]interface{}
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
			deleted = append(deleted, result)
		} else {
			failed = append(failed, result)
		}
	}

	// Build response
	switch {
	case len(failed) == 0 && len(deleted) > 0:
		return map[string]interface{}{
			"success": true,
			"message": deleted,
		}
	case len(failed) > 0 && len(deleted) == 0:
		return map[string]interface{}{
			"success": false,
			"message": failed,
		}
	case len(failed) > 0 && len(deleted) > 0:
		return map[string]interface{}{
			"success": true,
			"message": append(failed, deleted...),
		}
	default:
		return map[string]interface{}{
			"success": false,
			"message": "Data not found",
		}
	}
}

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

/* databases  */
func DatabaseHandler(options map[string]interface{}) map[string]interface{} {
	// Extract the requested action
	action, _ := options["action"].(string)
	if action == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Action is required",
		}
	}

	var query string
	var successMessage string

	switch action {

	// ---------------- DATABASE ACTIONS ----------------
	case "create_db":
		database, _ := options["database"].(string)
		if database == "" {
			return map[string]interface{}{
				"success": false,
				"message": "database_name is required",
			}
		}
		query = fmt.Sprintf("CREATE DATABASE %s", helpers.EscapeId(database))
		successMessage = fmt.Sprintf("Database '%s' created successfully", database)

	case "delete_db":
		database, _ := options["database"].(string)
		if database == "" {
			return map[string]interface{}{
				"success": false,
				"message": "database_name is required",
			}
		}
		query = fmt.Sprintf("DROP DATABASE IF EXISTS %s", helpers.EscapeId(database))
		successMessage = fmt.Sprintf("Database '%s' deleted successfully", database)

	case "update_db":
		database, _ := options["database"].(string)
		newtable, _ := options["newtable"].(string)
		if database == "" || newtable == "" {
			return map[string]interface{}{
				"success": false,
				"message": "Both old and new database_name are required",
			}
		}
		// Note: MySQL does not support rename database directly
		query = fmt.Sprintf("ALTER DATABASE %s UPGRADE DATA DIRECTORY NAME", helpers.EscapeId(database))
		successMessage = fmt.Sprintf("Database '%s' updated successfully", database)

	// ---------------- TABLE ACTIONS ----------------
	case "create_table":
		table, _ := options["table"].(string)
		cols, ok := options["columns"].([]map[string]string)
		if table == "" || !ok || len(cols) == 0 {
			return map[string]interface{}{
				"success": false,
				"message": "table_name and columns are required",
			}
		}

		colDefs := []string{}
		for _, col := range cols {
			name := helpers.EscapeId(col["name"])
			ctype := col["type"]
			if name == "" || ctype == "" {
				return map[string]interface{}{
					"success": false,
					"message": "Each column must have name and type",
				}
			}
			colDefs = append(colDefs, fmt.Sprintf("%s %s", name, ctype))
		}
		query = fmt.Sprintf("CREATE TABLE %s (%s)", helpers.EscapeId(table), strings.Join(colDefs, ", "))
		successMessage = fmt.Sprintf("Table '%s' created successfully", table)

	case "update_table":
		table, _ := options["table"].(string)
		newtable, _ := options["newtable"].(string)
		if table == "" || newtable == "" {
			return map[string]interface{}{
				"success": false,
				"message": "Both table_name and new table_name are required",
			}
		}
		query = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", helpers.EscapeId(table), helpers.EscapeId(newtable))
		successMessage = fmt.Sprintf("Table renamed from '%s' to '%s' successfully", table, newtable)

	case "delete_table":
		table, _ := options["table"].(string)
		if table == "" {
			return map[string]interface{}{
				"success": false,
				"message": "table_name is required",
			}
		}
		query = fmt.Sprintf("DROP TABLE IF EXISTS %s", helpers.EscapeId(table))
		successMessage = fmt.Sprintf("Table '%s' deleted successfully", table)

	// ---------------- COLUMN ACTIONS ----------------
	case "create_column":
		table, _ := options["table"].(string)
		column, _ := options["column"].(string)
		if table == "" || column == "" {
			return map[string]interface{}{
				"success": false,
				"message": "table_name and column are required",
			}
		}
		query = fmt.Sprintf("ALTER TABLE %s ADD %s", helpers.EscapeId(table), column)
		successMessage = fmt.Sprintf("Column added to table '%s' successfully", table)

	case "update_column":
		table, _ := options["table"].(string)
		column, _ := options["column"].(string)
		newColumn, _ := options["newColumn"].(string)
		if table == "" || column == "" || newColumn == "" {
			return map[string]interface{}{
				"success": false,
				"message": "table_name, column and newColumn are required",
			}
		}
		query = fmt.Sprintf("ALTER TABLE %s CHANGE %s %s", helpers.EscapeId(table), helpers.EscapeId(column), newColumn)
		successMessage = fmt.Sprintf("Column '%s' updated in table '%s' successfully", column, table)

	case "delete_column":
		table, _ := options["table"].(string)
		column, _ := options["column"].(string)
		if table == "" || column == "" {
			return map[string]interface{}{
				"success": false,
				"message": "Both table_name and column are required",
			}
		}
		query = fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", helpers.EscapeId(table), helpers.EscapeId(column))
		successMessage = fmt.Sprintf("Column '%s' removed from table '%s' successfully", column, table)

	// ---------------- UNKNOWN ACTION ----------------
	default:
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Unknown action: '%s'", action),
		}
	}

	// Execute query
	_, err := db.Exec(query)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": successMessage,
	}
}
