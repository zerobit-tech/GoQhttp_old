package database

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/onlysumitg/GoQhttp/go_ibm_db"
)

type ColumnType struct {
	IndexName string
	Name      string

	HasNullable       bool
	HasLength         bool
	HasPrecisionScale bool

	Nullable     bool
	Length       int64
	DatabaseType string
	Precision    int64
	Scale        int64

	IsLink bool
}

type DBServer interface {
	GetConnectionID() string
	GetConnectionType() string
	GetConnectionString() string
	MaxOpenConns() int
	MaxIdleConns() int
	ConnMaxIdleTime() time.Duration
	ConnMaxLifetime() time.Duration
}

var connectionMap map[string]*sql.DB = make(map[string]*sql.DB)

func ClearCache(server DBServer) {
	delete(connectionMap, server.GetConnectionID())
}

func GetConnection(server DBServer) (*sql.DB, error) {
	connectionID := server.GetConnectionID()
	db, found := connectionMap[connectionID]
	if found && db != nil {

		err := db.Ping()

		// error occured in ping
		if err != nil {
			db.Close()
			delete(connectionMap, connectionID)
		} else {
			return db, nil
		}

	}

	//fmt.Println((" ========================== BUILDING NEW CONNECTION ===================================="))
	db, err := sql.Open(strings.ToLower(server.GetConnectionType()), server.GetConnectionString())

	if err == nil {
		db.SetMaxOpenConns(server.MaxOpenConns())
		db.SetMaxIdleConns(server.MaxIdleConns())
		db.SetConnMaxIdleTime(server.ConnMaxIdleTime())
		db.SetConnMaxLifetime(server.ConnMaxLifetime())

		connectionMap[connectionID] = db

	} else {

		//log.Println(" connetion errror 1>>>>>>>>>>>>", err)
	}

	//db.Ping()

	return db, err
}

func GetSingleConnection(server DBServer) (*sql.DB, error) {

	db, err := sql.Open(strings.ToLower(server.GetConnectionType()), server.GetConnectionString())

	db.SetMaxOpenConns(1)
	db.Ping()

	return db, err
}
