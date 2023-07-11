package ibmiServer

import (
	"errors"
)

type ServerConnectionError struct {
	StatusCode int
	Err        error
}

func (m *ServerConnectionError) Error() string {
	return m.Err.Error()
}

var (
	ErrNoRecord = errors.New("models: no matching record found")
	// Add a new ErrInvalidCredentials error. We'll use this later if a user
	// tries to login with an incorrect email address or password.
	ErrInvalidCredentials = errors.New("Invalid user credentials")
	// Add a new ErrDuplicateEmail error. We'll use this later if a user
	// tries to signup with an email address that's already in use.
	ErrDuplicateEmail = errors.New("models: duplicate email")
	ErrNotFound       = errors.New("models: Not found")

	ErrUserNotFound       = errors.New("User not found")
	ErrServerNotFound     = errors.New("Not Found")
	ErrSavedQueryNotFound = errors.New("models: Saved query not found")

	SpNotFound = errors.New("Stored procedure not found. ")
)

//"Message": "SQLExecute: {42S22} [IBM][System i Access ODBC Driver][DB2 for i5/OS]SQL0206 - Column or global variable OCLOBFIELD not found.",
//SQLPrepare: {HY000} [IBM][System i Access ODBC Driver][DB2 for i5/OS]SQL0301 - Input variable *N or argument 1 not valid.
//"SQLExecute: {22001} [IBM][System i Access ODBC Driver]Column 1: CWB0111 - Input data is too big to fit into field\n{22001} [IBM][System i Access ODBC Driver]Column 1: Character data right truncation.",
//"SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8015 - CWBSY1006 - User ID is invalid, Password length = 0, Prompt Mode = Never, System IP Address = 185.113.5.134",
//SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8002 - CWBSY0002 - Password for user SGOYAL on system PUB400.COM is not correct, Password length = 0, Prompt Mode = Never, System IP Address = 185.113.5.134
//SQLExecute: {07002} [IBM][System i Access ODBC Driver]SQLBindParameter has not been called for parameter 4.
// SQLPrepare: {HY000} [IBM][System i Access ODBC Driver][DB2 for i5/OS]SQ20484 - Parameter 3 required for routine SPNUM2 in SUMITG1."
