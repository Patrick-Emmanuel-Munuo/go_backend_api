package helpers

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// This assumes a global DB variable
var DB *sql.DB // exported so other packages can use it
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
		parts = append(parts, fmt.Sprintf("BINARY %s = ?", EscapeId(k)))
		params = append(params, v)
	}
	return strings.Join(parts, " AND "), params
}

// WhereOr builds WHERE condition with OR
func GenerateWhereOr(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("BINARY %s = ?", EscapeId(k)))
		params = append(params, v)
	}
	return strings.Join(parts, " OR "), params
}

// Like builds WHERE with LIKE conditions
func GenerateLike(like map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var params []interface{}
	for key, val := range like {
		conditions = append(conditions, fmt.Sprintf("BINARY %s LIKE ?", EscapeId(key)))
		params = append(params, fmt.Sprintf("%%%v%%", val))
	}
	return strings.Join(conditions, " AND "), params
}

// UpdateSet builds on create data`SET field1=?, field2=?`
func GenerateSet(set map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for key, val := range set {
		parts = append(parts, fmt.Sprintf("%s = ?", EscapeId(key)))
		params = append(params, val)
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
func ScanRowss(rows *sql.Rows) ([]map[string]interface{}, error) {
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

// buildWhere generates the WHERE clause and parameter slice from conditions
func BuildWhere(condition, orCondition map[string]interface{}) (string, []interface{}, error) {
	// Validate: at least one condition must exist
	if len(condition) == 0 && len(orCondition) == 0 {
		return "", nil, fmt.Errorf("Missing condition(s)")
	}

	var whereParts []string
	var params []interface{}

	if len(condition) > 0 {
		w, p := GenerateWhere(condition)
		whereParts = append(whereParts, w)
		params = append(params, p...)
	}

	if len(orCondition) > 0 {
		w, p := GenerateWhereOr(orCondition)
		whereParts = append(whereParts, w)
		params = append(params, p...)
	}

	// Join conditions with AND
	return strings.Join(whereParts, " AND "), params, nil
}

// scanRows converts sql.Rows into []map[string]interface{} with byte -> string conversion
func ScanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			if b, ok := (*val).([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = *val
			}
		}
		results = append(results, rowMap)
	}
	return results, nil
}

// executeSelect runs SELECT query with params and returns results
func ExecuteSelect(query string, params ...interface{}) map[string]interface{} {
	rows, err := DB.Query(query, params...)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	defer rows.Close()

	results, err := ScanRows(rows)
	if err != nil {
		return map[string]interface{}{"success": false, "message": err.Error()}
	}
	if len(results) == 0 {
		return map[string]interface{}{"success": false, "message": "No data found"}
	}
	return map[string]interface{}{"success": true, "message": results}
}
