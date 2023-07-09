package mssqlserver

import (
	"fmt"

	"github.com/onlysumitg/GoQhttp/dbserver"
)

func init() {

	// Recover from panic to avoid stop an application when can't get the db2 cli
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("%s\nThe go_ibm_db driver cannot be registered", err))
		}
	}()

	msSqlServer := &MSSqlServer{}
	//go's to databse/sql/sql.go 43 line
	dbserver.Register("MS SQL Server", msSqlServer)

}
