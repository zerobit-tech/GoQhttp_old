package dbserver

import (
	"context"
	"sync"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

type DbDriver interface {
	Load(*Server)

	APICall(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error)
	ErrorToHttpStatus(inerr error) (int, string, string, bool)

	GetConnectionString() string
	GetSQLToPing() string
	GetPassword() string
	GetConnectionType() string
	PingTimeoutDuration() time.Duration
	GetSecretKey() string
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
	Refresh(ctx context.Context, sp *storedProc.StoredProc) error
	PreapreToSave(ctx context.Context, sp *storedProc.StoredProc) error
	DummyCall(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error)
	Exists(ctx context.Context, sp *storedProc.StoredProc) (bool, error)

	// Promotions
	ListPromotion(withupdate bool) ([]*storedProc.PromotionRecord, error)
	UpdateStatusForPromotionRecord(p storedProc.PromotionRecord)
	PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc

	// User Tokens
	UpdateStatusUserTokenTable(p storedProc.UserTokenSyncRecord)
	SyncUserTokenRecords(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error)
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
