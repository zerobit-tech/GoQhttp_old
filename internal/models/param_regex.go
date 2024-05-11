package models

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	bolt "go.etcd.io/bbolt"
)

type ParamRegex struct {
	Name                string `json:"name" db:"name" form:"name"`
	Regex               string `json:"regex" db:"regex" form:"regex"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type ParamRegexModel struct {
	DB *bolt.DB
}

func (m *ParamRegexModel) getTableName() []byte {
	return []byte("paramregex")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ParamRegexModel) Save(u *ParamRegex) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		u.Name = strings.ToUpper(u.Name)
		u.Name = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(u.Name))

		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}

		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		key := strings.ToUpper(u.Name)

		return bucket.Put([]byte(key), buf)
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *ParamRegexModel) Delete(id string) error {

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
func (m *ParamRegexModel) Get(id string) (*ParamRegex, error) {

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
	pr := ParamRegex{}
	if err != nil {
		return &pr, err
	}

	// log.Println("calllogJSON >2 >>", calllogJSON)

	if calllogJSON != nil {
		err := json.Unmarshal(calllogJSON, &pr)
		return &pr, err

	}

	return &pr, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ParamRegexModel) List() []*ParamRegex {
	prs := make([]*ParamRegex, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			pr := ParamRegex{}
			err := json.Unmarshal(v, &pr)
			if err == nil {
				prs = append(prs, &pr)
			}
		}

		return nil
	})
	return prs

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *ParamRegexModel) Map() map[string]string {
	rmap := make(map[string]string)
	for _, r := range m.List() {
		rmap[r.Name] = r.Regex
	}

	return rmap

}
