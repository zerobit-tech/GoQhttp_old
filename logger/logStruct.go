package logger

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"

	bolt "go.etcd.io/bbolt"
)

// var InfoLog *log.Logger = log.New(os.Stderr, "INFO \t", log.Ldate|log.Ltime)
// var ErrorLog *log.Logger = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
// var RequestLog *log.Logger = log.New(os.Stderr, "Request\t", log.Ldate|log.Ltime)
// var ResponseLog *log.Logger = log.New(os.Stderr, "Response\t", log.Ldate|log.Ltime)

// TODO move to app level

var LoggerChan chan *LogEvent = make(chan *LogEvent, 5000)

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
type LogEvent struct {
	EventTime  time.Time
	Id         string
	Type       string
	Message    string
	ScrubeData bool
}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------

func (l *LogEvent) String() string {

	return fmt.Sprintf("%s\t%s\t%s", l.EventTime.Format(stringutils.TimestampFormat), l.Type, l.Message)

}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
func GetLogEvent(logType string, id string, message string, scrubeData bool) *LogEvent {

	return &LogEvent{

		EventTime:  time.Now(),
		Id:         id,
		Type:       logType,
		Message:    message,
		ScrubeData: scrubeData,
	}

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func getLogTableName() []byte {
	return []byte("apilogs")
}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
func StartLogging(db *bolt.DB) {
	defer concurrent.Recoverer("SaveLogs")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	counter := 0
	for {
		logEvent, ok := <-LoggerChan
		if !ok {
			continue
		}
		counter += 1

		message := logEvent.String()
		scrubed := message

		if logEvent.ScrubeData {
			scrubed = RemoveNonLogData(scrubed)

		}

		db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists(getLogTableName())
			if err != nil {
				return err
			}
			key := fmt.Sprintf("%s_%d", logEvent.Id, counter)
			bucket.Put([]byte(key), []byte(fmt.Sprintf("%s", scrubed)))
			return nil
		})

	}
}
