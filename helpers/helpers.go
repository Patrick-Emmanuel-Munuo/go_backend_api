package helpers

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

func Where(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " AND "), params
}

func WhereOr(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " OR "), params
}

func JoinFields(fields []string) string {
	var wrapped []string
	for _, f := range fields {
		wrapped = append(wrapped, "`"+f+"`")
	}
	return strings.Join(wrapped, ", ")
}

func Like(like map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var params []interface{}

	for key, val := range like {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", key))
		params = append(params, fmt.Sprintf("%%%v%%", val))
	}

	return strings.Join(conditions, " AND "), params
}

// Generate a 24-char hex string similar to MongoDB ObjectId
func GenerateUniqueID() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err) // or handle error as you like
	}
	return fmt.Sprintf("%x", b)
}

// UpdateSet generates a SQL SET clause from a map of fields to update
func UpdateSet(set map[string]interface{}) string {
	var parts []string
	for key := range set {
		parts = append(parts, key+" = ?")
	}
	return strings.Join(parts, ", ")
}
