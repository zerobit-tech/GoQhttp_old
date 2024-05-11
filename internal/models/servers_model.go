package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"

	bolt "go.etcd.io/bbolt"
)

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
func (m *ServerModel) Insert(u *ibmiServer.Server) (string, error) {
	var id string = uuid.NewString()
	u.ID = id
	err := m.Update(u, false)

	u.LastAutoPromoteDate = time.Now().Format(godbc.TimestampFormat)

	return id, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ServerModel) Update(u *ibmiServer.Server, clearCache bool) error {

	u.ManageLibList()

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))
		// if strings.TrimSpace(u.Namespace) == "" {
		// 	u.Namespace = "V1"
		// }
		// u.Namespace = strings.ToUpper(u.Namespace)

		// u.Namespace = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(u.Namespace))

		if !u.OnHold {
			u.OnHoldMessage = ""
		} else {
			go u.ClearCache() //goroutine
		}

		if clearCache {
			go u.ClearCache() //goroutine
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
func (m *ServerModel) DuplicateName(serverToCheck *ibmiServer.Server) bool {
	exists := false
	for _, server := range m.List() {
		//fmt.Println(">>>>duplucate name<<<", server.Name, "<>", serverToCheck.Name, "||", server.ID, "<>", serverToCheck.ID)
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
func (m *ServerModel) Get(id string) (*ibmiServer.Server, error) {

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
	server := ibmiServer.Server{}
	if err != nil {
		return &server, err
	}

	// log.Println("serverJSON >2 >>", serverJSON)

	if serverJSON != nil {
		err := json.Unmarshal(serverJSON, &server)
		return &server, err

	}
	//server.Load()
	return &server, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ServerModel) List() []*ibmiServer.Server {
	servers := make([]*ibmiServer.Server, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			server := ibmiServer.Server{}
			err := json.Unmarshal(v, &server)
			if err == nil {
				//server.Load()
				servers = append(servers, &server)
			}
		}

		return nil
	})
	return servers

}
