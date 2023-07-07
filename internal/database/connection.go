package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	_ "github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

var mapLock sync.Mutex

// ---------------------------------------------------
//
// ---------------------------------------------------
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

// ---------------------------------------------------
//
// ---------------------------------------------------
type DBServer interface {
	GetConnectionID() string
	GetConnectionType() string
	GetConnectionString() string
	MaxOpenConns() int
	MaxIdleConns() int
	ConnMaxIdleTime() time.Duration
	ConnMaxLifetime() time.Duration
	PingTimeoutDuration() time.Duration
	GetSQLToPing() string
	GetMux() *sync.Mutex
}

// ---------------------------------------------------
//
// ---------------------------------------------------
var connectionMap concurrent.MapInterface = concurrent.NewSuperEfficientSyncMap(0)

//var connectionMap2 sync.Map

// ---------------------------------------------------
//
// ---------------------------------------------------
func ClearCache(server DBServer) {
	//delete(connectionMap, server.GetConnectionID())
	concurrent.EraseSyncMap(connectionMap)
}

// ---------------------------------------------------
//
// ---------------------------------------------------
func GetConnectionFromCache(server DBServer) *sql.DB {
	mapLock.Lock()

	defer mapLock.Unlock()

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	connectionID := server.GetConnectionID()
	dbX, found := connectionMap.Load(connectionID)
	if !found || dbX == nil {
		return nil
	}

	db, ok := dbX.(*sql.DB)
	if !ok {
		return nil
	}

	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().InUse ", db.Stats().InUse)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().OpenConnections ", db.Stats().OpenConnections)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().Idle ", db.Stats().Idle)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().WaitCount ", db.Stats().WaitCount)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().WaitDuration ", db.Stats().WaitDuration)

	if db.Stats().InUse > 0 {
		return db
	}

	ctx, cancel := context.WithTimeout(context.Background(), server.PingTimeoutDuration())

	sqlToPing := server.GetSQLToPing()
	if sqlToPing != "" {
		ctx = context.WithValue(ctx, go_ibm_db.SQL_TO_PING, sqlToPing)
	}

	defer func() {

		cancel()

	}()

	err := db.PingContext(ctx)

	// error occured in ping
	if err != nil {

		fmt.Println("Closing connections .....", err)
		db.Close()
		connectionMap.Delete(connectionID)
		return nil
	} else {
		return db
	}

}

// ---------------------------------------------------
//
// ---------------------------------------------------
func GetConnection(server DBServer) (*sql.DB, error) {

	db := GetConnectionFromCache(server)
	if db != nil {
		return db, nil
	}

	connectionID := server.GetConnectionID()

	fmt.Println((" ========================== BUILDING NEW CONNECTION ===================================="))
	db, err := sql.Open(strings.ToLower(server.GetConnectionType()), server.GetConnectionString())

	if err == nil {
		mapLock.Lock()
		dboldX, found := connectionMap.Load(connectionID)
		if found && dboldX != nil {
			dbY, ok := dboldX.(*sql.DB)
			if ok {
				fmt.Println((" still in map ===================================="))
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().InUse ", db.Stats().InUse)
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().OpenConnections ", db.Stats().OpenConnections)
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().Idle ", db.Stats().Idle)
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitCount ", db.Stats().WaitCount)
				//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitDuration ", db.Stats().WaitDuration)
				dbY.Close()
			}
		}
		mapLock.Unlock()
		connectionMap.Store(connectionID, db)
		db.SetMaxOpenConns(server.MaxOpenConns())
		fmt.Println("server.MaxIdleConns(", server.MaxIdleConns())
		db.SetMaxIdleConns(server.MaxIdleConns())
		db.SetConnMaxIdleTime(server.ConnMaxIdleTime())
		db.SetConnMaxLifetime(server.ConnMaxLifetime())

	} else {

		log.Println(" connetion errror 1>>>>>>>>>>>>", err)
	}

	//db.Ping()

	return db, err
}

// ---------------------------------------------------
//
// ---------------------------------------------------
func GetSingleConnection(server DBServer) (*sql.DB, error) {

	db, err := sql.Open(strings.ToLower(server.GetConnectionType()), server.GetConnectionString())

	db.SetMaxOpenConns(1)
	db.Ping()

	return db, err
}
