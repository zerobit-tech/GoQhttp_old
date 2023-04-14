package rbac

import (
	"errors"
	"log"
	"strings"

	"github.com/mikespook/gorbac"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	bolt "go.etcd.io/bbolt"
)

//----------------------------------------------------
//
//---------------------------------------------------

func (r *RBAC) RegisterPermission(permission string) {
	r.Model.SavePermission(permission)
}

// ----------------------------------------------------
//
// ---------------------------------------------------
type RbackPermissionForm struct {
	Permission string `json:"permission" db:"permission" form:"permission"`
	validator.Validator
}

// ----------------------------------------------------
//
// ---------------------------------------------------
func (m *RbacModel) getPermissionTable() []byte {
	return []byte("rbacpermissions")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (m *RbacModel) DeleteAllPermissions() error {

	err := m.DB.Update(func(tx *bolt.Tx) error {

		return tx.DeleteBucket(m.getPermissionTable())

	})

	return err
}

func (m *RbacModel) PermissionExists(id string) bool {

	var userJson []byte

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getPermissionTable())
		if bucket == nil {
			return errors.New("Table not found")
		}
		key := strings.ToUpper(strings.TrimSpace(id))

		userJson = bucket.Get([]byte(key))

		return nil

	})
	if err != nil {
		log.Println("PermissionExists:", err.Error())
	}
	return (userJson != nil)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RbacModel) SavePermission(permission string) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getPermissionTable())
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(strings.TrimSpace(permission)) // + string(itob(u.ID))

		return bucket.Put([]byte(key), []byte(permission))
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RbacModel) ListPermission() gorbac.Permissions {

	rbackPermissions := make(gorbac.Permissions)

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getPermissionTable())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			rbackPermissions[string(v)] = gorbac.NewStdPermission(string(v))
		}

		return nil
	})

	return rbackPermissions

}
