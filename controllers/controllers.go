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
	/*if helpers.DatabasePassword == "" {
		log.Println("MYSQL_PASSWORD not set in environment or helpers")
		return map[string]interface{}{
			"success": false,
			"message": "MySQL password not configured.",
		}
	}*/
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
	// Handle SELECT fields using helpers.GenerateSelect
	selectFields := "*"
	if select_data, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(select_data)
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
	// Build WHERE clause
	var whereClause string
	var params []interface{}

	switch {
	case len(condition) > 0 && len(orCondition) > 0:
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	case len(condition) > 0:
		whereClause, params = helpers.GenerateWhere(condition)
	case len(orCondition) > 0:
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	default:
		whereClause = "1=1" // fallback: no condition, selects all
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

// generate table relationship use mulitiple tables keys for table joining
func ReadJoin(options map[string]interface{}) map[string]interface{} {
	// Validate base table
	baseTable, ok := options["table"].(string)
	if !ok || baseTable == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Missing or invalid base table name",
		}
	}

	// Handle SELECT fields
	selectFields := "*"
	if sel, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(sel)
	}

	// Handle conditions
	condition := map[string]interface{}{}
	orCondition := map[string]interface{}{}
	if cond, ok := options["condition"].(map[string]interface{}); ok {
		condition = cond
	}
	if orCond, ok := options["or_condition"].(map[string]interface{}); ok {
		orCondition = orCond
	}

	// Build WHERE clause
	whereClause, params := "", []interface{}{}
	switch {
	case len(condition) > 0 && len(orCondition) > 0:
		w1, p1 := helpers.GenerateWhere(condition)
		w2, p2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", w1, w2)
		params = append(params, p1...)
		params = append(params, p2...)
	case len(condition) > 0:
		whereClause, params = helpers.GenerateWhere(condition)
	case len(orCondition) > 0:
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	default:
		whereClause = "1=1"
	}

	// Handle JOINs (expects options["joins"] as array of map[string]interface{})
	joinClause := ""
	if joins, ok := options["joins"].([]interface{}); ok {
		for _, j := range joins {
			if joinMap, ok := j.(map[string]interface{}); ok {
				joinType := "INNER JOIN"
				if jt, ok := joinMap["type"].(string); ok {
					joinType = jt
				}
				tableName, ok1 := joinMap["table"].(string)
				onClause, ok2 := joinMap["on"].(string)
				if ok1 && ok2 {
					joinClause += fmt.Sprintf(" %s %s ON %s", joinType, tableName, onClause)
				}
			}
		}
	}
	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s%s WHERE %s", selectFields, baseTable, joinClause, whereClause)

	// Execute query
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

	// Handle SELECT fields using helpers.GenerateSelect
	selectFields := "*"
	if select_data, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(select_data)
	}

	// Build WHERE clause
	var whereClause string
	var params []interface{}

	switch {
	case len(condition) > 0 && len(orCondition) > 0:
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	case len(condition) > 0:
		whereClause, params = helpers.GenerateWhere(condition)
	case len(orCondition) > 0:
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	default:
		whereClause = "1=1" // fallback: no condition, selects all
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

	// Handle SELECT fields using helpers.GenerateSelect
	selectFields := "*"
	if select_data, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(select_data)
	}

	// Build WHERE clause
	params := []interface{}{start, end}
	var whereClause string

	switch {
	case len(condition) > 0 && len(orCondition) > 0:
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	case len(condition) > 0:
		whereClause, params = helpers.GenerateWhere(condition)
	case len(orCondition) > 0:
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	default:
		whereClause = "1=1" // fallback: no condition, selects all
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

// List returns paginated results from a table with optional WHERE conditions
func List(options map[string]interface{}) map[string]interface{} {
	// Validate table
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}

	// Conditions
	condition, _ := options["condition"].(map[string]interface{})
	orCondition, _ := options["or_condition"].(map[string]interface{})

	// Pagination
	pageFloat, ok := options["page"].(float64)
	if !ok || pageFloat <= 0 {
		pageFloat = 1
	}
	pageSizeFloat, ok := options["page_size"].(float64)
	if !ok || pageSizeFloat <= 0 {
		pageSizeFloat = 10
	}
	page := int(pageFloat)
	pageSize := int(pageSizeFloat)
	offset := (page - 1) * pageSize

	// Handle SELECT fields
	selectFields := "*"
	if selectData, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(selectData)
	}

	// Build WHERE clause
	var whereClause string
	var params []interface{}
	switch {
	case len(condition) > 0 && len(orCondition) > 0:
		where1, params1 := helpers.GenerateWhere(condition)
		where2, params2 := helpers.GenerateWhereOr(orCondition)
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", where1, where2)
		params = append(params, params1...)
		params = append(params, params2...)
	case len(condition) > 0:
		whereClause, params = helpers.GenerateWhere(condition)
	case len(orCondition) > 0:
		whereClause, params = helpers.GenerateWhereOr(orCondition)
	default:
		whereClause = "1=1"
	}

	// Query total records for pagination
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereClause)
	var totalRecords int
	if err := db.QueryRow(totalQuery, params...).Scan(&totalRecords); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Query actual data with LIMIT/OFFSET
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT ? OFFSET ?", selectFields, table, whereClause)
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
			if byteVal, ok := (*val).([]uint8); ok {
				rowMap[col] = string(byteVal)
			} else {
				rowMap[col] = *val
			}
		}
		results = append(results, rowMap)
	}

	// Build pagination metadata
	totalPages := totalRecords / pageSize
	if totalRecords%pageSize > 0 {
		totalPages++
	}
	pages := make([]int, totalPages)
	for i := 0; i < totalPages; i++ {
		pages[i] = i + 1
	}
	// Calculate previous and next page
	var previous *int
	if page > 1 {
		prev := page - 1
		previous = &prev
	}

	var next *int
	if page < totalPages {
		nxt := page + 1
		next = &nxt
	}
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"data":         results,
			"totalRecords": totalRecords,
			"page":         page,
			"pageSize":     pageSize,
			"totalPages":   totalPages,
			"previous":     previous,
			"next":         next,
			"pages":        pages,
		},
	}
}

// ListAll returns all records from a table
func ListAll(options map[string]interface{}) map[string]interface{} {
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}
	// Handle SELECT fields using helpers.GenerateSelect
	selectFields := "*"
	if select_data, ok := options["select"]; ok {
		selectFields = helpers.GenerateSelect(select_data)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectFields, table)
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
	// Validate table
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
	}

	// Validate data
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

	// Build SET clause
	setClause, params := helpers.GenerateSet(data)
	query := fmt.Sprintf("INSERT INTO %s SET %s", table, setClause)

	// Execute query
	result, err := db.Exec(query, params...)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Get last insert ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Add id to the original data map
	data["id"] = lastID

	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"id":   lastID,
			"data": data,
		},
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
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
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
	setClause, setparams := helpers.GenerateSet(data)
	whereClause, whereParams := helpers.GenerateWhere(condition)
	// Build parameters in correct order: data values first, then condition values
	var params []interface{}
	params = append(params, whereParams...)
	params = append(params, setparams...)
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
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Table name is required",
		}
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
