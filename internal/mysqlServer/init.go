package mysqlserver

import (
	"fmt"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
)

func init() {

	// Recover from panic to avoid stop an application when can't get the db2 cli
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("%s\nThe MySQL driver cannot be registered", err))
		}
	}()

	ibmIServer := &IBMiServer{}
	//go's to databse/sql/sql.go 43 line
	dbserver.Register("MySQL", ibmIServer)

}
