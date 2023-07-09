package dbserver

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new User type. Notice how the field names and types align
// with the columns in the database "users" table?
type Server struct {
	Mux sync.Mutex `json:"-" db:"-" form:"-"`

	ID   string `json:"id" db:"id" form:"id"`
	Type string `json:"type" db:"type" form:"type"`

	Name string `json:"server_name" db:"server_name" form:"name"`
	IP   string `json:"ip" db:"ip" form:"ip"`
	Port uint16 `json:"port" db:"port" form:"port"`
	Ssl  bool   `json:"ssl" db:"ssl" form:"ssl"`

	UserName string `json:"un" db:"un" form:"user_name"`
	Password string `json:"pwd" db:"pwd" form:"password"`

	

	//WorkLib           string    `json:"wlib" db:"wlib" form:"worklib"`
	CreatedAt       time.Time `json:"c_at" db:"c_at" form:"-"`
	UpdatedAt       time.Time `json:"u_at" db:"u_at" form:"-"`
	ConnectionsOpen int       `json:"conn" db:"conn" form:"connections"`
	ConnectionsIdle int       `json:"iconn" db:"iconn" form:"idleconnections"`

	ConnectionMaxAge  int    `json:"cage" db:"cage" form:"cage"`
	ConnectionIdleAge int    `json:"icage" db:"icage" form:"icage"`
	PingTimeout       int    `json:"pingtout" db:"pingtout" form:"pingtout"`
	PingQuery         string `json:"pingquery" db:"pingquery" form:"pingquery"`

	OnHold        bool   `json:"oh" db:"oh" form:"onhold"`
	OnHoldMessage string `json:"ohm" db:"ohm" form:"onholdmessage"`

	ConfigFileLib string `json:"configfilelib" db:"configfilelib" form:"configfilelib"`
	ConfigFile    string `json:"configfile" db:"configfile" form:"configfile"`

	AutoPromotePrefix string `json:"autopromoteprefix" db:"autopromoteprefix" form:"autopromoteprefix"`

	UserTokenFileLib string `json:"usertokenfilelib" db:"usertokenfilelib" form:"usertokenfilelib"`
	UserTokenFile    string `json:"usertokenfile" db:"usertokenfile" form:"usertokenfile"`

	LastAutoPromoteDate string `json:"lastautopromotecheck" db:"lastautopromotecheck" form:"lastautopromotecheck"`

	validator.Validator `json:"-" db:"-" form:"-"`

	dbDriver DbDriver `json:"-" db:"-" form:"-"`
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionString() string {
	return s.GetDbDriver().GetConnectionString()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) GetSQLToPing() string {
	return s.GetDbDriver().GetSQLToPing()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetPassword() string {
	return s.GetDbDriver().GetPassword()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionType() string {
	return s.GetDbDriver().GetConnectionType()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) MaxOpenConns() int {
	//return s.GetDbDriver().MaxOpenConns()
	if s.ConnectionsOpen <= 0 {
		return 2
	}
	return s.ConnectionsOpen

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) MaxIdleConns() int {
	//return s.GetDbDriver().MaxIdleConns()
	if s.ConnectionsIdle <= 0 {
		return 2
	}
	return s.ConnectionsIdle

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ConnMaxIdleTime() time.Duration {
	//return s.GetDbDriver().ConnMaxIdleTime()
	age := 10
	if s.ConnectionIdleAge > 0 {
		age = s.ConnectionIdleAge
	}

	return time.Duration(age) * time.Second

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ConnMaxLifetime() time.Duration {
	age := 10
	if s.ConnectionMaxAge > 0 {
		age = s.ConnectionMaxAge
	}

	return time.Duration(age) * time.Second

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PingTimeoutDuration() time.Duration {
	return s.GetDbDriver().PingTimeoutDuration()

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ClearCache() {
	defer concurrent.Recoverer("ClearCache")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	// database.ClearCache(s) //TODO
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	db, err := GetConnection(s)

	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().InUse ", db.Stats().InUse)
	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().OpenConnections ", db.Stats().OpenConnections)
	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().MaxOpenConnections ", db.Stats().MaxOpenConnections)
	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().Idle ", db.Stats().Idle)
	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitCount ", db.Stats().WaitCount)
	//fmt.Println(s.Name, " >>>>>>>>>>>>>>>>>>>>> dbY.Stats().WaitDuration ", db.Stats().WaitDuration)

	return db, err
}

	// ------------------------------------------------------------
	//
	// ------------------------------------------------------------
func (s *Server) GetSingleConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	return GetSingleConnection(s)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetMux() *sync.Mutex {
	return &s.Mux
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetSecretKey() string {
	return s.GetDbDriver().GetSecretKey()

	//return "Ang&1*~U^2^#s0^=)^^7#b34"
}
