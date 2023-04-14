package rbac

import (
	"errors"
	"strings"

	"github.com/mikespook/gorbac"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	bolt "go.etcd.io/bbolt"
)

type RbackRoleForm struct {
	Role string `json:"role" db:"role" form:"role"`
	validator.Validator
}

type RbacModel struct {
	DB *bolt.DB
}

func (m *RbacModel) getRolesTable() []byte {
	return []byte("rbacroles")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RbacModel) SaveRole(roleName string) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getRolesTable())
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(roleName) // + string(itob(u.ID))

		return bucket.Put([]byte(key), []byte(key))
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RbacModel) ListRoles() gorbac.Roles {

	rbackRoles := make(gorbac.Roles)

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getRolesTable())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			rbackRoles[string(v)] = gorbac.NewStdRole(string(v))
		}

		return nil
	})

	return rbackRoles

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RbacModel) ListRolesAsString() []string {

	rbackRoles := make([]string, 0)

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getRolesTable())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			rbackRoles = append(rbackRoles, string(v))
		}

		return nil
	})

	return rbackRoles

}
