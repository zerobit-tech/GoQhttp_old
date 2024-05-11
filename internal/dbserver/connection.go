package dbserver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/microsoft/go-mssqldb"

	//_ "github.com/zerobit-tech/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
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
var connectionInvalidCache concurrent.MapInterface = concurrent.NewSuperEfficientSyncMap(0)

//var connectionMap2 sync.Map

// ---------------------------------------------------
//
// ---------------------------------------------------
func ClearCache(server DBServer) error {
	connectionInvalidCache.Store(server.GetConnectionID(), true)
	return nil

}

// ---------------------------------------------------
//
// ---------------------------------------------------
func CloseConnections() {
	//delete(connectionMap, server.GetConnectionID())
	connectionMap.Range(func(key, value interface{}) bool {

		dbY, ok := value.(*sql.DB)
		if ok {
			dbY.Close()
		}
		return true
	})
}

// ---------------------------------------------------
//
// ---------------------------------------------------
func getConnectionFromCache(server DBServer) (_ *sql.DB, inuse bool) {
	mapLock.Lock()

	defer mapLock.Unlock()
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	connectionID := server.GetConnectionID()

	_, found := connectionInvalidCache.Load(connectionID)
	if found {
		connectionInvalidCache.Delete(connectionID)
		return nil, false
	}

	dbX, found := connectionMap.Load(connectionID)
	if !found || dbX == nil {
		return nil, false
	}

	db, ok := dbX.(*sql.DB)
	if !ok {
		return nil, false
	}

	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().InUse ", db.Stats().InUse)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().OpenConnections ", db.Stats().OpenConnections)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().Idle ", db.Stats().Idle)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().WaitCount ", db.Stats().WaitCount)
	//fmt.Println(" >>>>>>>>>>>>>>>>>>>>> db.Stats().WaitDuration ", db.Stats().WaitDuration)

	if db.Stats().InUse > 0 {
		return db, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), server.PingTimeoutDuration())

	sqlToPing := server.GetSQLToPing()
	if sqlToPing != "" {
		ctx = context.WithValue(ctx, godbc.SQL_TO_PING, sqlToPing)
	}

	defer func() {

		cancel()

	}()

	err := db.PingContext(ctx)

	// error occured in ping
	if err != nil {

		log.Println("Closing connections .....", err)
		db.Close()
		connectionMap.Delete(connectionID)
		return nil, false
	} else {
		return db, false
	}

}

// ---------------------------------------------------
//
// ---------------------------------------------------
func GetConnection(server DBServer) (*sql.DB, error) {

	db, _ := getConnectionFromCache(server)
	if db != nil {
		return db, nil
	}

	connectionID := server.GetConnectionID()

	log.Println(("** Loading new DB connection **"))
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
		//fmt.Println("server.MaxIdleConns(", server.MaxIdleConns())
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
	if err != nil {
		return db, err
	}
	db.SetMaxOpenConns(1)
	err = db.PingContext(context.TODO())

	return db, err
}
