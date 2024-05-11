package models

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	bolt "go.etcd.io/bbolt"
)

type StoredProcModel struct {
	DB *bolt.DB
}

func (m *StoredProcModel) getTableName() []byte {
	return []byte("storedprocs")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) Save(u *storedProc.StoredProc) (string, error) {
	var id string
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))

		u.SetNameSpace()

		// generate new ID if id is blank else use the old one to update
		if u.ID == "" {
			u.ID = u.Slug() //uuid.NewString()
			//u.AllowedOnServers = make([]*ServerRecord, 0)
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))
		u.Lib = strings.ToUpper(strings.TrimSpace(u.Lib))
		u.EndPointName = strings.ToLower(strings.TrimSpace(u.EndPointName))
		id = u.ID
		// Marshal user data into bytes.
		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(u.ID) // + string(itob(u.ID))

		return bucket.Put([]byte(key), buf)
	})

	return id, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) Delete(id string) error {

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
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) DeleteByName(name string, method string) error {

	for _, sp := range m.List(true) {
		if strings.EqualFold(sp.EndPointName, name) && strings.EqualFold(sp.HttpMethod, method) {
			err := m.Delete(sp.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) Exists(id string) bool {

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
func (m *StoredProcModel) Duplicate(u *storedProc.StoredProc) bool {
	exists := false
	for _, sp := range m.List(true) {

		if sp.ID != u.ID && strings.EqualFold(sp.EndPointName, u.EndPointName) && strings.EqualFold(sp.HttpMethod, u.HttpMethod) && strings.EqualFold(sp.GetNamespace(), u.GetNamespace()) {
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
func (m *StoredProcModel) DuplicateByName(name string, method string, namespace string) bool {
	exists := false
	for _, sp := range m.List(true) {

		if strings.EqualFold(sp.EndPointName, name) && strings.EqualFold(sp.HttpMethod, method) && strings.EqualFold(sp.GetNamespace(), namespace) {
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
func (m *StoredProcModel) Get(id string) (*storedProc.StoredProc, error) {

	if id == "" {
		return nil, errors.New("SavedQuery blank id not allowed")
	}
	var savedQueryJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		savedQueryJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	savedQuery := storedProc.StoredProc{}
	if err != nil {
		return &savedQuery, err
	}

	// log.Println("savedQueryJSON >2 >>", savedQueryJSON)

	if savedQueryJSON != nil {
		err := json.Unmarshal(savedQueryJSON, &savedQuery)
		return &savedQuery, err

	}

	return &savedQuery, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) List(loadSpecial bool) []*storedProc.StoredProc {
	savedQueries := make([]*storedProc.StoredProc, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			savedQuery := storedProc.StoredProc{}
			err := json.Unmarshal(v, &savedQuery)
			if err == nil {

				if !loadSpecial && savedQuery.IsSpecial {
					continue
				}

				savedQueries = append(savedQueries, &savedQuery)
			}
		}

		return nil
	})
	return savedQueries

}
