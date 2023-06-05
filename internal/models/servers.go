package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/onlysumitg/GoQhttp/internal/database"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"

	bolt "go.etcd.io/bbolt"
)

// keep the length same
const MySecret string = "Ang&1*~U^2^#s0^=)^^7%b34"

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new User type. Notice how the field names and types align
// with the columns in the database "users" table?
type Server struct {
	ID   string `json:"id" db:"id" form:"id"`
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

	ConnectionMaxAge  int `json:"cage" db:"cage" form:"cage"`
	ConnectionIdleAge int `json:"icage" db:"icage" form:"icage"`

	OnHold        bool   `json:"oh" db:"oh" form:"onhold"`
	OnHoldMessage string `json:"ohm" db:"ohm" form:"onholdmessage"`

	ConfigFileLib string `json:"configfilelib" db:"configfilelib" form:"configfilelib"`
	ConfigFile string `json:"configfile" db:"configfile" form:"configfile"`

	validator.Validator `json:"-" db:"-" form:"-"`
}

 




// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) GetConnectionString() string {
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
func (s Server) GetPassword() string {
	pwd, err := stringutils.Decrypt(s.Password, MySecret)
	if err != nil {
		log.Println("Unable to decrypt password")
		return ""
	}
	return pwd
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) GetConnectionType() string {
	return "go_ibm_db" //"odbc"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) MaxOpenConns() int {
	if s.ConnectionsOpen <= 0 {
		return 2
	}
	return s.ConnectionsOpen
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) MaxIdleConns() int {
	if s.ConnectionsIdle <= 0 {
		return 2
	}
	return s.ConnectionsIdle
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) ConnMaxIdleTime() time.Duration {
	age := 10
	if s.ConnectionIdleAge > 0 {
		age = s.ConnectionIdleAge
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) ConnMaxLifetime() time.Duration {
	age := 10
	if s.ConnectionMaxAge > 0 {
		age = s.ConnectionMaxAge
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) GetConnectionID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) ClearCache() {
	database.ClearCache(s)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s Server) GetConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	return database.GetConnection(s)
}
func (s Server) GetSinglaConnection() (*sql.DB, error) {
	if s.OnHold {
		return nil, fmt.Errorf("Server is on hold due to %s", s.OnHoldMessage)
	}

	return database.GetSingleConnection(s)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type ServerModel struct {
	DB *bolt.DB
}

func (m *ServerModel) getTableName() []byte {
	return []byte("servers")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ServerModel) Insert(u *Server) (string, error) {
	var id string = uuid.NewString()
	u.ID = id
	err := m.Update(u, false)

	return id, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ServerModel) Update(u *Server, clearCache bool) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))
		u.Password, _ = stringutils.Encrypt(u.Password, MySecret)

		if !u.OnHold {
			u.OnHoldMessage = ""
		} else {
			go u.ClearCache()
		}

		if clearCache {
			go u.ClearCache()
		}

		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(u.ID) // + string(itob(u.ID))

		return bucket.Put([]byte(key), buf)
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ServerModel) Delete(id string) error {

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
func (m *ServerModel) Exists(id string) bool {

	var userJson []byte

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		key := strings.ToUpper(id)

		userJson = bucket.Get([]byte(key))

		return nil

	})

	return (userJson != nil)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ServerModel) DuplicateName(serverToCheck *Server) bool {
	exists := false
	for _, server := range m.List() {
		fmt.Println(">>>>duplucate name<<<", server.Name, "<>", serverToCheck.Name, "||", server.ID, "<>", serverToCheck.ID)
		if strings.EqualFold(server.Name, serverToCheck.Name) && !strings.EqualFold(server.ID, serverToCheck.ID) {
			exists = true
			break
		}
	}

	return exists
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ServerModel) Get(id string) (*Server, error) {

	if id == "" {
		return nil, errors.New("Server blank id not allowed")
	}
	var serverJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		serverJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	server := Server{}
	if err != nil {
		return &server, err
	}

	// log.Println("serverJSON >2 >>", serverJSON)

	if serverJSON != nil {
		err := json.Unmarshal(serverJSON, &server)
		return &server, err

	}

	return &server, ErrServerNotFound

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ServerModel) List() []*Server {
	servers := make([]*Server, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			server := Server{}
			err := json.Unmarshal(v, &server)
			if err == nil {
				servers = append(servers, &server)
			}
		}

		return nil
	})
	return servers

}
