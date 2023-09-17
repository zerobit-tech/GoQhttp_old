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
type RpgProgramModel struct {
	DB *bolt.DB
}

func (m *RpgProgramModel) getTableName() []byte {
	return []byte("rpgprogram")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RpgProgramModel) Save(u *Program) (string, error) {
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
func (m *RpgProgramModel) Update(u *Program, clearCache bool) error {

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
func (m *RpgProgramModel) Delete(id string) error {

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
func (m *RpgProgramModel) Exists(id string) bool {

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
func (m *RpgProgramModel) DuplicateName(rpgprogramToCheck *Program) bool {
	exists := false
	for _, rpgprogram := range m.List() {
		//fmt.Println(">>>>duplucate name<<<", rpgprogram.Name, "<>", rpgprogramToCheck.Name, "||", rpgprogram.ID, "<>", rpgprogramToCheck.ID)
		if strings.EqualFold(rpgprogram.Name, rpgprogramToCheck.Name) && !strings.EqualFold(rpgprogram.ID, rpgprogramToCheck.ID) {
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
func (m *RpgProgramModel) Get(id string) (*Program, error) {

	if id == "" {
		return nil, errors.New("RpgProgram blank id not allowed")
	}
	var rpgprogramJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		rpgprogramJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	rpgprogram := Program{}
	if err != nil {
		return &rpgprogram, err
	}

	// log.Println("rpgprogramJSON >2 >>", rpgprogramJSON)

	if rpgprogramJSON != nil {
		err := json.Unmarshal(rpgprogramJSON, &rpgprogram)
		return &rpgprogram, err

	}
	//rpgprogram.Load()
	return &rpgprogram, errors.New("Not Found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RpgProgramModel) List() []*Program {
	rpgprograms := make([]*Program, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			rpgprogram := Program{}
			err := json.Unmarshal(v, &rpgprogram)
			if err == nil {
				//rpgprogram.Load()
				rpgprograms = append(rpgprograms, &rpgprogram)
			}
		}

		return nil
	})
	return rpgprograms

}
