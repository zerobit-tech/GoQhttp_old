package storedProc

import "github.com/zerobit-tech/GoQhttp/internal/validator"

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type UserTokenSyncRecord struct {
	Rowid               string
	Username            string // D: Delete   R:Refresh   I:Insert
	Token               string
	Status              string
	StatusMessage       string
	validator.Validator `json:"-" db:"-" form:"-"`
}
