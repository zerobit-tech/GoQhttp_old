package models

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"runtime/debug"
// 	"sync"
// 	"time"

// 	"github.com/onlysumitg/GoQhttp/internal/database"
// 	"github.com/onlysumitg/GoQhttp/internal/validator"
// 	"github.com/onlysumitg/GoQhttp/utils/concurrent"
// 	"github.com/onlysumitg/GoQhttp/utils/stringutils"
// )

// // keep the length same
// const MySecret string = "Ang&1*~U^2^#s0^=)^^7#b34"

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// // Define a new User type. Notice how the field names and types align
// // with the columns in the database "users" table?
// type Server struct {
// 	mux sync.Mutex `json:"-" db:"-" form:"-"`

// 	ID   string `json:"id" db:"id" form:"id"`
// 	Name string `json:"server_name" db:"server_name" form:"name"`
// 	IP   string `json:"ip" db:"ip" form:"ip"`
// 	Port uint16 `json:"port" db:"port" form:"port"`
// 	Ssl  bool   `json:"ssl" db:"ssl" form:"ssl"`

// 	UserName string `json:"un" db:"un" form:"user_name"`
// 	Password string `json:"pwd" db:"pwd" form:"password"`
// 	//WorkLib           string    `json:"wlib" db:"wlib" form:"worklib"`
// 	CreatedAt       time.Time `json:"c_at" db:"c_at" form:"-"`
// 	UpdatedAt       time.Time `json:"u_at" db:"u_at" form:"-"`
// 	ConnectionsOpen int       `json:"conn" db:"conn" form:"connections"`
// 	ConnectionsIdle int       `json:"iconn" db:"iconn" form:"idleconnections"`

// 	ConnectionMaxAge  int    `json:"cage" db:"cage" form:"cage"`
// 	ConnectionIdleAge int    `json:"icage" db:"icage" form:"icage"`
// 	PingTimeout       int    `json:"pingtout" db:"pingtout" form:"pingtout"`
// 	PingQuery         string `json:"pingquery" db:"pingquery" form:"pingquery"`

// 	OnHold        bool   `json:"oh" db:"oh" form:"onhold"`
// 	OnHoldMessage string `json:"ohm" db:"ohm" form:"onholdmessage"`

// 	ConfigFileLib string `json:"configfilelib" db:"configfilelib" form:"configfilelib"`
// 	ConfigFile    string `json:"configfile" db:"configfile" form:"configfile"`

// 	AutoPromotePrefix string `json:"autopromoteprefix" db:"autopromoteprefix" form:"autopromoteprefix"`

// 	UserTokenFileLib string `json:"usertokenfilelib" db:"usertokenfilelib" form:"usertokenfilelib"`
// 	UserTokenFile    string `json:"usertokenfile" db:"usertokenfile" form:"usertokenfile"`

// 	LastAutoPromoteDate string `json:"lastautopromotecheck" db:"lastautopromotecheck" form:"lastautopromotecheck"`

// 	validator.Validator `json:"-" db:"-" form:"-"`
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetConnectionString() string {
// 	driver := "IBM i Access ODBC Driver"
// 	ssl := 0
// 	if s.Ssl {
// 		ssl = 1
// 	}
// 	pwd := s.GetPassword()
// 	connectionString := fmt.Sprintf("DRIVER=%s;SYSTEM=%s; UID=%s;PWD=%s;DBQ=*USRLIBL;UNICODESQL=1;XDYNAMIC=1;EXTCOLINFO=0;PKG=A/DJANGO,2,0,0,1,512;PROTOCOL=TCPIP;NAM=1;CMT=0;SSL=%d;ALLOWUNSCHAR=1", driver, s.IP, s.UserName, pwd, ssl)

// 	//connectionString := fmt.Sprintf("DSN=pub400; UID=%s;PWD=%s", s.UserName, s.Password)

// 	return connectionString
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------

// func (s *Server) GetSQLToPing() string {
// 	return s.PingQuery
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetPassword() string {
// 	pwd, err := stringutils.Decrypt(s.Password, MySecret)
// 	if err != nil {
// 		log.Println("Unable to decrypt password")
// 		return ""
// 	}
// 	return pwd
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetConnectionType() string {
// 	return "go_ibm_db" //"odbc"
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) MaxOpenConns() int {
// 	if s.ConnectionsOpen <= 0 {
// 		return 2
// 	}
// 	return s.ConnectionsOpen
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) MaxIdleConns() int {
// 	if s.ConnectionsIdle <= 0 {
// 		return 2
// 	}
// 	return s.ConnectionsIdle
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) ConnMaxIdleTime() time.Duration {
// 	age := 10
// 	if s.ConnectionIdleAge > 0 {
// 		age = s.ConnectionIdleAge
// 	}

// 	return time.Duration(age) * time.Second
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) ConnMaxLifetime() time.Duration {
// 	age := 10
// 	if s.ConnectionMaxAge > 0 {
// 		age = s.ConnectionMaxAge
// 	}

// 	return time.Duration(age) * time.Second
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) PingTimeoutDuration() time.Duration {
// 	age := 3
// 	if s.PingTimeout > 0 {
// 		age = s.PingTimeout
// 	}

// 	return time.Duration(age) * time.Second
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetConnectionID() string {
// 	return s.ID
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) ClearCache() {
// 	defer concurrent.Recoverer("ClearCache")
// 	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

// 	database.ClearCache(s)
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetConnection() (*sql.DB, error) {
// 	if s.OnHold {
// 		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
// 	}

// 	db, err := database.GetConnection(s)

// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().InUse ", db.Stats().InUse)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().OpenConnections ", db.Stats().OpenConnections)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().Idle ", db.Stats().Idle)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitCount ", db.Stats().WaitCount)
// 	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitDuration ", db.Stats().WaitDuration)

// 	return db, err
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetSingleConnection() (*sql.DB, error) {
// 	if s.OnHold {
// 		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
// 	}

// 	return database.GetSingleConnection(s)
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
// func (s *Server) GetMux() *sync.Mutex {
// 	return &s.mux
// }

// // ------------------------------------------------------------
// //
// // ------------------------------------------------------------
