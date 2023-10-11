package rpg

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"

	bolt "go.etcd.io/bbolt"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type RpgParamModel struct {
	DB *bolt.DB
}

func (m *RpgParamModel) getTableName() []byte {
	return []byte("rpgparams")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RpgParamModel) Save(u *Param) (string, error) {
	if u.ID == "" {
		var id string = uuid.NewString()
		u.ID = id
	}

	err := m.Update(u, false)

	return u.ID, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RpgParamModel) Update(u *Param, clearCache bool) error {

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
func (m *RpgParamModel) Delete(id string) error {

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
func (m *RpgParamModel) Exists(id string) bool {

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
func (m *RpgParamModel) DuplicateName(rpgparamToCheck *Param) bool {
	exists := false
	for _, rpgparam := range m.List() {
		//fmt.Println(">>>>duplucate name<<<", rpgparam.Name, "<>", rpgparamToCheck.Name, "||", rpgparam.ID, "<>", rpgparamToCheck.ID)
		if strings.EqualFold(rpgparam.Name, rpgparamToCheck.Name) && !strings.EqualFold(rpgparam.ID, rpgparamToCheck.ID) {
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
func (m *RpgParamModel) Get(id string) (*Param, error) {

	if id == "" {
		return nil, errors.New("RpgParam blank id not allowed")
	}
	var rpgparamJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		rpgparamJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	rpgparam := Param{}
	if err != nil {
		return &rpgparam, err
	}

	// log.Println("rpgparamJSON >2 >>", rpgparamJSON)

	if rpgparamJSON != nil {
		err := json.Unmarshal(rpgparamJSON, &rpgparam)
		if err == nil {
			m.loadChildParas(&rpgparam)

		}
		return &rpgparam, err

	}
	//pgmfields.Load()
	return &rpgparam, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RpgParamModel) List() []*Param {
	rpgparams := make([]*Param, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			rpgparam := Param{}
			err := json.Unmarshal(v, &rpgparam)
			if err == nil {
				//pgmfields.Load()
				m.loadChildParas(&rpgparam)
				rpgparams = append(rpgparams, &rpgparam)
			}
		}

		return nil
	})
	return rpgparams

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RpgParamModel) loadChildParas(p *Param) {
	if !p.IsDs {
		return
	}

	for _, f := range p.DsFields {
		f1, err := m.Get(f.ParamID)

		if err == nil {
			f.Param = f1
		}
	}

}
