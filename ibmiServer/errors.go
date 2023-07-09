package ibmiServer

import (
	"errors"
	"net/http"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
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

func (s *IBMiServer) ErrorToHttpStatus(inerr error) (int, string, string, bool) {
	var odbcError *go_ibm_db.Error

	if errors.As(inerr, &odbcError) {

		if len(odbcError.Diag) > 0 {
			code := odbcError.Diag[0].NativeError
			switch code {
			case -420:
				return http.StatusBadRequest, "Please check the values.", odbcError.Error(), true
			case -204:
				return http.StatusNotFound, "OD0204[42S02]", odbcError.Error(), true
			case 8001:
				return http.StatusInternalServerError, "OD8001", odbcError.Error(), true
			case 10060:
				return http.StatusInternalServerError, "OD10060", odbcError.Error(), true
			case 30038:
				return http.StatusInternalServerError, "OD30038", odbcError.Error(), true
			case 30189:
				return http.StatusInternalServerError, "OD30189", odbcError.Error(), true // {HYT00} [IBM][System i Access ODBC Driver]Connection login timed out.
			case 10065:
				return http.StatusInternalServerError, "OD10065", odbcError.Error(), true // "[IBM][System i Access ODBC Driver]Communication link failure. comm rc=10065 - CWBCO1003 - Sockets error, function  returned 10065 "
			case 8002:
				return http.StatusInternalServerError, "OD8002", odbcError.Error(), true // SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8002 - CWBSY0002 - Password for user SUMITG33 on system PUB400.COM is not correct, Password length = 10, Prompt Mode = Never, System IP Address = 185.113.5.134
			}

		}

		return http.StatusBadRequest, odbcError.Error(), odbcError.Error(), true
	}
	return 0, "", "", false

}

//"Message": "SQLExecute: {42S22} [IBM][System i Access ODBC Driver][DB2 for i5/OS]SQL0206 - Column or global variable OCLOBFIELD not found.",
//SQLPrepare: {HY000} [IBM][System i Access ODBC Driver][DB2 for i5/OS]SQL0301 - Input variable *N or argument 1 not valid.
//"SQLExecute: {22001} [IBM][System i Access ODBC Driver]Column 1: CWB0111 - Input data is too big to fit into field\n{22001} [IBM][System i Access ODBC Driver]Column 1: Character data right truncation.",
//"SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8015 - CWBSY1006 - User ID is invalid, Password length = 0, Prompt Mode = Never, System IP Address = 185.113.5.134",
//SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8002 - CWBSY0002 - Password for user SGOYAL on system PUB400.COM is not correct, Password length = 0, Prompt Mode = Never, System IP Address = 185.113.5.134
//SQLExecute: {07002} [IBM][System i Access ODBC Driver]SQLBindParameter has not been called for parameter 4.
