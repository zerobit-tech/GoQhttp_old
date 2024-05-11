package models

import (
	"encoding/json"
	"errors"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/endpoints"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	bolt "go.etcd.io/bbolt"
)

type LogEntry struct {
	LogID    string `json:"logid" db:"logid" form:"-"`
	CalledAt time.Time
}

type SPCallLog struct {
	SpID string `json:"spid" db:"spid" form:"-"`

	Logs []LogEntry `json:"logid" db:"logid" form:"-"`
}
type SPCallLogEntry struct {
	EndPoint endpoints.Endpoint

	LogId string
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type SPCallLogModel struct {
	DB       *bolt.DB
	dbmux    sync.Mutex
	DataChan chan SPCallLogEntry
}

func (m *SPCallLogModel) getTableName() []byte {
	return []byte("spcalllogs")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *SPCallLogModel) AddLogid() {

	defer concurrent.Recoverer("AddLogid")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	for {
		logE, ok := <-m.DataChan
		if !ok {
			return
		}

		m.dbmux.Lock()

		logEntry := LogEntry{
			LogID:    logE.LogId,
			CalledAt: time.Now().Local(),
		}

		splog, err := m.Get(logE.EndPoint.EPID())

		if err != nil {
			logEntries := make([]LogEntry, 0)
			splog = &SPCallLog{SpID: logE.EndPoint.EPID(), Logs: logEntries}
		}

		//Prepend
		splog.Logs = append([]LogEntry{logEntry}, splog.Logs...)

		maxEntries, err := strconv.Atoi(env.GetEnvVariable("MAX_LOG_ENTRIES_FOR_ONE_ENDPOINT", "1000"))
		if err != nil || maxEntries <= 0 {
			maxEntries = 1000
		}

		maxEntriesByEP := logE.EndPoint.EPMaxLogEntries()
		if maxEntries > maxEntriesByEP {
			maxEntries = maxEntriesByEP
		}

		if len(splog.Logs) > maxEntries {

			//delete extra log entries
			entriesToDelete := splog.Logs[maxEntries:]

			for _, ed := range entriesToDelete {
				DeleteLog(m.DB, ed.LogID)
			}

			splog.Logs = splog.Logs[0:maxEntries]
		}

		m.Save(splog)

		m.dbmux.Unlock()
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *SPCallLogModel) Save(u *SPCallLog) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}

		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(u.SpID) // + string(itob(u.ID))

		return bucket.Put([]byte(key), buf)
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *SPCallLogModel) Delete(id string) error {

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		key := strings.ToUpper(id)
		dbDeleteError := bucket.Delete([]byte(key))
		return dbDeleteError
	})

	return err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *SPCallLogModel) Get(id string) (*SPCallLog, error) {

	if id == "" {
		return nil, errors.New("Server blank id not allowed")
	}
	var calllogJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		calllogJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	calllog := SPCallLog{}
	if err != nil {
		return &calllog, err
	}

	// log.Println("calllogJSON >2 >>", calllogJSON)

	if calllogJSON != nil {
		err := json.Unmarshal(calllogJSON, &calllog)
		return &calllog, err

	}

	return &calllog, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *SPCallLogModel) List() []*SPCallLog {
	calllogs := make([]*SPCallLog, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			calllog := SPCallLog{}
			err := json.Unmarshal(v, &calllog)
			if err == nil {
				calllogs = append(calllogs, &calllog)
			}
		}

		return nil
	})
	return calllogs

}
