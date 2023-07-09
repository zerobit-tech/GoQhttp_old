package mssqlserver

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/onlysumitg/GoQhttp/utils/concurrent"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetConnectionString() string {

	pwd := s.GetPassword()

	//connectionString := fmt.Sprintf("DSN=pub400; UID=%s;PWD=%s", s.UserName, s.Password)
	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", s.IP, s.UserName, pwd, s.Port, "database")
	return connectionString
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *MSSqlServer) GetSQLToPing() string {
	return s.PingQuery
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetPassword() string {
	pwd, err := stringutils.Decrypt(s.Password, s.GetSecretKey())
	if err != nil {
		log.Println("Unable to decrypt password")
		return ""
	}
	return pwd
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetConnectionType() string {
	return "sqlserver" //"odbc"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *IBMiServer) MaxOpenConns() int {
// 	if s.ConnectionsOpen <= 0 {
// 		return 2
// 	}
// 	return s.ConnectionsOpen
// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *IBMiServer) MaxIdleConns() int {
// 	if s.ConnectionsIdle <= 0 {
// 		return 2
// 	}
// 	return s.ConnectionsIdle
// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *IBMiServer) ConnMaxIdleTime() time.Duration {
// 	age := 10
// 	if s.ConnectionIdleAge > 0 {
// 		age = s.ConnectionIdleAge
// 	}

// 	return time.Duration(age) * time.Second
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *IBMiServer) ConnMaxLifetime() time.Duration {
// 	age := 10
// 	if s.ConnectionMaxAge > 0 {
// 		age = s.ConnectionMaxAge
// 	}

// 	return time.Duration(age) * time.Second
// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) PingTimeoutDuration() time.Duration {
	age := 3
	if s.PingTimeout > 0 {
		age = s.PingTimeout
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetConnectionID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) ClearCache() {
	defer concurrent.Recoverer("ClearCache")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	// database.ClearCache(s) //TODO
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *IBMiServer) GetConnection() (*sql.DB, error) {
// 	if s.OnHold {
// 		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
// 	}

// 	db, err := GetConnection(s)

// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().InUse ", db.Stats().InUse)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().OpenConnections ", db.Stats().OpenConnections)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().Idle ", db.Stats().Idle)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitCount ", db.Stats().WaitCount)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitDuration ", db.Stats().WaitDuration)

// 	return db, err
// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *IBMiServer) GetSingleConnection() (*sql.DB, error) {
// 	if s.OnHold {
// 		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
// 	}

// 	return GetSingleConnection(s)
// }

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetMux() *sync.Mutex {
	return &s.Mux
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetSecretKey() string {
	return "BhL&1*~U^2^#s0^=)^^8#b34" // keep the length
}
