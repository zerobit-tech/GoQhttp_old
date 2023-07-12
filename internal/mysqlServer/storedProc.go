package mysqlserver

import (
	"database/sql"
	"fmt"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
)

func RowsToResultsets(rows *sql.Rows, dummyCall bool) map[string][]map[string]any {
	resultsets := make(map[string][]map[string]any, 0)

	counter := 1

	resultSet, _ := dbserver.ToMap(rows, -1, dummyCall)

	if resultSet == nil {
		return resultsets
	}
	resultsets[fmt.Sprintf("DATASET_%d", counter)] = resultSet

	for rows.NextResultSet() {
		counter += 1
		resultSet, _ := dbserver.ToMap(rows, -1, dummyCall)
		resultsets[fmt.Sprintf("DATASET_%d", counter)] = resultSet

	}
	return resultsets

}
