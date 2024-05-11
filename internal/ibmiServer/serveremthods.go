package ibmiServer

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/dbserver"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (s *Server) GetSQLToPing() string {
	return s.PingQuery
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetUserName() string {
	user := s.UserName
	if strings.ToUpper(s.UserName) == "*ENV" {
		user = env.GetServerUserName(s.Name)

	}

	return user

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
func (s *Server) GetConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	db, err := dbserver.GetConnection(s)

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
func (s *Server) ClearCache() error {
	defer concurrent.Recoverer("ClearCache")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	return dbserver.ClearCache(s)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetSingleConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	return dbserver.GetSingleConnection(s)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetMux() *sync.Mutex {
	return &s.Mux
}
