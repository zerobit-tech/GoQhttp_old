package dbserver

import (
	"context"
	"sync"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

type DbDriver interface {
	LoadX(*Server)

	APICallX(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error)
	ErrorToHttpStatusX(inerr error) (int, string, string, bool)

	GetConnectionStringX() string
	//GetSQLToPing() string
	GetPasswordX() string
	GetConnectionTypeX() string
	PingTimeoutDurationX() time.Duration
	GetSecretKeyX() string
	//MaxOpenConns() int
	//MaxIdleConns() int
	//ConnMaxIdleTime() time.Duration
	//ConnMaxLifetime() time.Duration

	//GetConnectionID() string
	//ClearCache()
	//GetConnection() (*sql.DB, error)
	//GetSingleConnection() (*sql.DB, error)

	//GetMux() *sync.Mutex

	// StoredPrcd
	RefreshX(ctx context.Context, sp *storedProc.StoredProc) error
	PrepareToSaveX(ctx context.Context, sp *storedProc.StoredProc) error
	DummyCallX(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error)
	ExistsX(ctx context.Context, sp *storedProc.StoredProc) (bool, error)

	// Promotions
	ListPromotionX(withupdate bool) ([]*storedProc.PromotionRecord, error)
	UpdateStatusForPromotionRecordX(p storedProc.PromotionRecord)
	//PromotionRecordToStoredProcX(p storedProc.PromotionRecord) *storedProc.StoredProc

	// User Tokens
	UpdateStatusUserTokenTableX(p storedProc.UserTokenSyncRecord)
	SyncUserTokenRecordsX(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var (
	driversMu sync.RWMutex
	drivers   = make(map[string]DbDriver)
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver DbDriver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("DbDriver: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("DbDriver: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func GetRegisterDrivers() []string {
	returnList := make([]string, 0)
	driversMu.Lock()
	defer driversMu.Unlock()
	for k, _ := range drivers {
		returnList = append(returnList, k)
	}
	return returnList
}
