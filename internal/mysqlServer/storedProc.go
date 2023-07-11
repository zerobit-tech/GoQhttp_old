package mssqlserver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
)

func breakCatalogSchema(sp *storedProc.StoredProc) (string, string) {
	return breakStringToCatalogSchema(sp.Lib)

}

func breakStringToCatalogSchema(lib string) (string, string) {
	spCatalog := lib
	spSchema := "dbo"

	if strings.Contains(lib, ".") {
		splitS := strings.Split(lib, ".")
		spCatalog = splitS[0]
		spSchema = splitS[1]

	}

	return spCatalog, spSchema

}

func RowsToResultsets(rows *sql.Rows, dummyCall bool) map[string][]map[string]any {
	resultsets := make(map[string][]map[string]any, 0)

	counter := 1

	resultSet, _ := dbserver.ToMap(rows, -1, dummyCall)
	resultsets[fmt.Sprintf("DATASET_%d", counter)] = resultSet

	for rows.NextResultSet() {
		counter += 1
		resultSet, _ := dbserver.ToMap(rows, -1, dummyCall)
		resultsets[fmt.Sprintf("DATASET_%d", counter)] = resultSet

	}
	return resultsets

}
