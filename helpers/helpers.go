package helpers

import (
	"fmt"
	"strings"
)

func Where(cond map[string]interface{}) string {
	var clauses []string
	for k, v := range cond {
		clauses = append(clauses, fmt.Sprintf("`%s`='%v'", k, v))
	}
	return strings.Join(clauses, " AND ")
}

func WhereOr(cond map[string]interface{}) string {
	var clauses []string
	for k, v := range cond {
		clauses = append(clauses, fmt.Sprintf("`%s`='%v'", k, v))
	}
	return strings.Join(clauses, " OR ")
}

func JoinFields(fields []string) string {
	var wrapped []string
	for _, f := range fields {
		wrapped = append(wrapped, "`"+f+"`")
	}
	return strings.Join(wrapped, ", ")
}
