package mssqlserver

import (
	"fmt"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
)

func init() {

	// Recover from panic to avoid stop an application when can't get the db2 cli
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s\nThe MS SQL Server driver cannot be registered \n", err)
		}
	}()

	msSqlServer := &MSSqlServer{}
	//go's to databse/sql/sql.go 43 line
	dbserver.Register("MS SQL Server", msSqlServer)

}
