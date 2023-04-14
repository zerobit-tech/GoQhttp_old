package rbac

import (
	"github.com/mikespook/gorbac"
	bolt "go.etcd.io/bbolt"
)

type RBAC struct {
	RBAC  *gorbac.RBAC
	DB    *bolt.DB
	Model *RbacModel
}

var rbac *RBAC

// ----------------------------------------------------
//
// ---------------------------------------------------
func (r *RBAC) AssignDB(db *bolt.DB) {
	r.DB = db

}

// ----------------------------------------------------
//
// ---------------------------------------------------
func GetRbac(db *bolt.DB) *RBAC {

	if rbac == nil {
		rbac = &RBAC{
			RBAC:  gorbac.New(),
			DB:    db,
			Model: &RbacModel{DB: db},
		}

		rbac.LoadRbacRoles()
	}

	return rbac

}

//----------------------------------------------------
//
//---------------------------------------------------

func (r *RBAC) LoadRbacRoles() {
	roles := r.Model.ListRolePermissions()

	for _, v := range roles {
		rl := gorbac.NewStdRole(v.Role)

		r.RBAC.Remove(rl.ID())

		for _, p := range v.Permissions {
			rl.Assign(gorbac.NewStdPermission(p))
		}

		r.RBAC.Add(rl)
	}
}
