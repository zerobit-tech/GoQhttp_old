package dbserver

import (
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func prepareColumnType(sql_column sql.ColumnType, index int) (column_type ColumnType) {

	name := sql_column.Name()
	length, hasLength := sql_column.Length()
	precision, scale, hasPrecisionScale := sql_column.DecimalSize()
	nullable, hasNullable := sql_column.Nullable()
	databaseType := sql_column.DatabaseTypeName()

	column_type = ColumnType{
		IndexName:         fmt.Sprintf("%d_%s", index, name),
		Name:              name,
		Length:            length,
		HasLength:         hasLength,
		Precision:         precision,
		Scale:             scale,
		HasPrecisionScale: hasPrecisionScale,
		Nullable:          nullable,
		HasNullable:       hasNullable,
		DatabaseType:      databaseType,
	}
	return
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func prepareColumnTypes(rows *sql.Rows) (columnTypes []ColumnType) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("prepareColumnTypes", r)
		}
	}()
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	column_types_p, _ := rows.ColumnTypes()

	columnTypes = make([]ColumnType, 0)
	for index, sql_column := range column_types_p {
		columnTypes = append(columnTypes, prepareColumnType(*sql_column, index))
	}

	return columnTypes
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func processRow(scans []interface{}, fields []string, dummyCall bool, columnTypes []ColumnType) map[string]interface{} {
	row := make(map[string]interface{})
	for i, v := range scans {

		//fmt.Println(">>>>>>>", fields[i], " type = ", reflect.TypeOf(v), reflect.ValueOf(v).Kind())

		if dummyCall {
			row[fields[i]] = columnTypes[i].DatabaseType
			continue
		}

		switch v.(type) {
		case []uint, []uint8:

			row[fields[i]] = fmt.Sprintf("%s", v)
		default:
			row[fields[i]] = v
		}
		// if reflect.TypeOf(v) == []byte {
		// 	row[fields[i]] = string(v)
		// }
		// else {
		// row[fields[i]] = v
		// }

	}
	return row
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func ToMap(rows *sql.Rows, maxRows int, dummyCall bool) (return_rows []map[string]interface{}, column_types []ColumnType) {

	fieldsX, _ := rows.Columns()
	// fmt.Println("fields >>>", fieldsX)

	fields := make([]string, 0)
	fields = append(fields, fieldsX...)

	colch := make(chan []ColumnType)
	defer close(colch)

	if dummyCall {

		column_types = prepareColumnTypes(rows) //goroutine
	}

	//fmt.Println("ToMap rows.JumpToRow2(3)", rows.JumpToRow2(scrollTo))
	for rows.Next() {
		//rows.JumpToRow(3)

		scans := make([]interface{}, len(fields))

		for i := range scans {
			scans[i] = &scans[i]
		}

		err := rows.Scan(scans...)
		if err != nil {
			log.Println("ToMap Scan....:", err.Error())
		}

		return_rows = append(return_rows, processRow(scans, fields, dummyCall, column_types))
		if maxRows > 0 && len(return_rows) >= maxRows {
			break
		}
	}

	return
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func processRow2(scans []interface{}, fields []string, rowch chan<- map[string]interface{}, wg *sync.WaitGroup) { //map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Println("processRow2", r)
		}
	}()
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	wg.Add(1)
	row := make(map[string]interface{})
	for i, v := range scans {

		// fmt.Println(">>>>>>>", fields[i], " type = ", reflect.TypeOf(v), reflect.ValueOf(v).Kind())

		switch v.(type) {
		case []uint, []uint8:

			row[fields[i]] = fmt.Sprintf("%s", v)
		default:
			row[fields[i]] = v
		}
		// if reflect.TypeOf(v) == []byte {
		// 	row[fields[i]] = string(v)
		// }
		// else {
		// row[fields[i]] = v
		// }

	}
	wg.Done()
	rowch <- row
}
