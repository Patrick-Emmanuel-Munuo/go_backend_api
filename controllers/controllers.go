package controllers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var db *sql.DB

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

// mysql read controller
func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return fmt.Errorf("mysql open error: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("mysql ping error: %w", err)
	}
	return nil
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

/*// Read fetches data from the DB with optional conditions.
func Read(ctx context.Context, options map[string]interface{}) (map[string]interface{}, error) {
	table, ok := options["table"].(string)
	if !ok || table == "" {
		return map[string]interface{}{"success": false, "message": "Table name is required"}, nil
	}

	condition, _ := options["condition"].(map[string]interface{})
	orCondition, _ := options["or_condition"].(map[string]interface{})
	selectFieldsMap, _ := options["select"].(map[string]interface{})

	if len(condition) == 0 && len(orCondition) == 0 {
		return map[string]interface{}{"success": false, "message": "At least one of 'condition' or 'or_condition' is required"}, nil
	}

	selectFields := "*"
	if len(selectFieldsMap) > 0 {
		fields := make([]string, 0, len(selectFieldsMap))
		for k := range selectFieldsMap {
			fields = append(fields, k)
		}
		selectFields = ""
		for i, f := range fields {
			if i > 0 {
				selectFields += ", "
			}
			selectFields += f
		}
	}

	var whereClause string
	hasCondition := len(condition) > 0
	hasOrCondition := len(orCondition) > 0

	if hasCondition && hasOrCondition {
		whereClause = fmt.Sprintf("( %s ) AND ( %s )", helpers.Where(condition), helpers.WhereOr(orCondition))
	} else if hasCondition {
		whereClause = helpers.Where(condition)
	} else if hasOrCondition {
		whereClause = helpers.WhereOr(orCondition)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, table, whereClause)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("Query error:", err)
		return map[string]interface{}{"success": false, "message": err.Error()}, nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Println("Failed to get columns:", err)
		return map[string]interface{}{"success": false, "message": err.Error()}, nil
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		cols := make([]interface{}, len(columns))
		colPointers := make([]interface{}, len(columns))
		for i := range cols {
			colPointers[i] = &cols[i]
		}

		if err := rows.Scan(colPointers...); err != nil {
			log.Println("Row scan failed:", err)
			return map[string]interface{}{"success": false, "message": err.Error()}, nil
		}

		rowMap := map[string]interface{}{}
		for i, colName := range columns {
			val := colPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}

	if len(results) == 0 {
		return map[string]interface{}{"success": false, "message": "not found data"}, nil
	}
	return map[string]interface{}{"success": true, "message": results}, nil
}*/
