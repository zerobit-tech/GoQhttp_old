package storedProc

import "github.com/onlysumitg/GoQhttp/internal/validator"

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type UserTokenSyncRecord struct {
	Rowid               int
	Username            string // D: Delete   R:Refresh   I:Insert
	Token               string
	Status              string
	StatusMessage       string
	validator.Validator `json:"-" db:"-" form:"-"`
}
