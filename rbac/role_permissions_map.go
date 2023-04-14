package rbac

import (
	"encoding/json"
	"errors"
	"strings"

	bolt "go.etcd.io/bbolt"
)

type RolePermissionMaper struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func (m *RbacModel) getRolePermissionTable() []byte {
	return []byte("rbackrolepermission")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *RbacModel) SaveRolePermissions(rp RolePermissionMaper) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getRolePermissionTable())
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(rp.Role) // + string(itob(u.ID))
		buf, err := json.Marshal(rp)

		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), buf)
	})

	return err

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RbacModel) ListRolePermissions() []RolePermissionMaper {

	rps := make([]RolePermissionMaper, 0)

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getRolePermissionTable())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			rp := RolePermissionMaper{}
			err := json.Unmarshal(v, &rp)
			if err == nil {
				rps = append(rps, rp)
			}
		}

		return nil
	})

	return rps

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *RbacModel) ListRolePermission(role string) (RolePermissionMaper, error) {

	var rpJson []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getRolePermissionTable())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		rpJson = bucket.Get([]byte(strings.ToUpper(role)))

		return nil

	})
	rp := RolePermissionMaper{}

	if rpJson != nil {
		err = json.Unmarshal(rpJson, &rp)
	}

	return rp, err
}
