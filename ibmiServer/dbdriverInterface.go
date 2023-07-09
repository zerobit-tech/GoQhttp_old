package ibmiServer

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
func (s *IBMiServer) GetConnectionString() string {
	driver := "IBM i Access ODBC Driver"
	ssl := 0
	if s.Ssl {
		ssl = 1
	}
	pwd := s.GetPassword()
	connectionString := fmt.Sprintf("DRIVER=%s;SYSTEM=%s; UID=%s;PWD=%s;DBQ=*USRLIBL;UNICODESQL=1;XDYNAMIC=1;EXTCOLINFO=0;PKG=A/DJANGO,2,0,0,1,512;PROTOCOL=TCPIP;NAM=1;CMT=0;SSL=%d;ALLOWUNSCHAR=1", driver, s.IP, s.UserName, pwd, ssl)

	//connectionString := fmt.Sprintf("DSN=pub400; UID=%s;PWD=%s", s.UserName, s.Password)

	return connectionString
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *IBMiServer) GetSQLToPing() string {
	return s.PingQuery
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *IBMiServer) GetPassword() string {
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
func (s *IBMiServer) GetConnectionType() string {
	return "go_ibm_db" //"odbc"
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
func (s *IBMiServer) PingTimeoutDuration() time.Duration {
	age := 3
	if s.PingTimeout > 0 {
		age = s.PingTimeout
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *IBMiServer) GetConnectionID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *IBMiServer) ClearCache() {
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
func (s *IBMiServer) GetMux() *sync.Mutex {
	return &s.Mux
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *IBMiServer) GetSecretKey() string {
	return "Ang&1*~U^2^#s0^=)^^7#b34"
}
