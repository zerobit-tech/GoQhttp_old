package models

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/jwtutils"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

type ContextKey string

const ContextUserKey ContextKey = "userid"
const ContextUserName ContextKey = "username"

type UserVerification struct {
	UserID         string
	VerificationID string
	Active         bool
}

type UserTroubleForm struct {
	Email               string `json:"email" db:"email" form:"email"`
	Option              string `json:"option" db:"option" form:"option"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

type UserPasswordResetForm struct {
	Password            string `json:"-" db:"-" form:"password"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

type UserSignUpForm struct {
	Name                string `json:"name" db:"name" form:"name"`
	Email               string `json:"email" db:"email" form:"email"`
	Password            string `json:"-" db:"-" form:"password"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

type UserLoginForm struct {
	Email               string `json:"email" db:"email" form:"email"`
	Password            string `json:"-" db:"-" form:"password"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

type AddOwnerForm struct {
	Email               string `json:"email" db:"email" form:"email"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new User type. Notice how the field names and types align
// with the columns in the database "users" table?
type User struct {
	ID                  string    `json:"id" db:"id" form:"id"`
	Name                string    `json:"name" db:"name" form:"name"`
	Email               string    `json:"email" db:"email" form:"email"`
	Password            string    `json:"-" db:"-" form:"password"`
	HashedPassword      []byte    `json:"passwordhashed" db:"passwordhashed" form:"-"`
	Created             time.Time `json:"created" db:"created" form:"-"`
	ResetPassword       bool
	IsSuperUser         bool `json:"issuperuser" db:"issuperuser" form:"issuperuser"`
	IsStaff             bool `json:"isstaff" db:"isstaff" form:"isstaff"`
	HasVerified         bool `json:"hasverified" db:"hasverified" form:"hasverified"`
	MaxAllowedEndpoints int  `json:"maxallowedendpoints" db:"maxallowedendpoints" form:"maxallowedendpoints"`
	MaxInvalidAttempts  int  `json:"maxinvalidattempts" db:"maxinvalidattempts" form:"maxinvalidattempts"`
	InvalidAttempts     int  `json:"invalidattempts" db:"invalidattempts" form:"-"`
	Disabled            bool `json:"disabled" db:"disabled" form:"-"`
	//Role           string `json:"role" db:"role" form:"role"`
	validator.Validator `json:"-" db:"-" form:"-"`

	OwnedEndPoints []string `json:"endpoints" db:"endpoints" form:"-"`

	Token string `json:"token" db:"token" form:"-"`

	ServerId    string `json:"serverid" db:"serverid" form:"serverid"`
	LandingPage string `json:"landingpage" db:"landingpage" form:"landingpage"`
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (u *User) GetNewToken(expiaryDuration time.Duration) (string, error) {
	jwtClaims := map[string]any{
		"sub": u.ID,
		//	"useremail": u.Email,
		//	"isadmin":   u.IsSuperUser,
		//	"isstaff":   u.IsStaff,
	}

	jwtClaims["exp"] = time.Now().UTC().Add(expiaryDuration).Unix()
	jwtString, err := jwtutils.Get(jwtClaims)
	return jwtString, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (u *User) OwnsEndPoint(id string) bool {

	if u.IsSuperUser {
		return true
	}
	for _, eID := range u.OwnedEndPoints {
		if strings.EqualFold(eID, id) {
			return true
		}
	}

	return false
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (u *User) AssignOwnedEndPoint(id string) {

	for _, eID := range u.OwnedEndPoints {
		if strings.EqualFold(eID, id) {
			return
		}
	}

	u.OwnedEndPoints = append(u.OwnedEndPoints, id)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (u *User) setEncruptedPassword(password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		u.HashedPassword = []byte(password)
	} else {
		u.HashedPassword = hash
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type UserModel struct {
	DB *bolt.DB
}

func (m *UserModel) getTableName() []byte {
	return []byte("users")
}

func (m *UserModel) GetVerificationTableName() []byte {
	return []byte("usersverification")
}
func (m *UserModel) GetPasswordResetTableName() []byte {
	return []byte("userspasword")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (m *UserModel) ClearEndPointowners(endpointID string) {
	users := m.GetEndPointowners(endpointID)

	for _, u := range users {
		newOwnList := make([]string, 0)
		for _, epid := range u.OwnedEndPoints {
			if !strings.EqualFold(epid, endpointID) {
				newOwnList = append(newOwnList, epid)
			}
		}

		u.OwnedEndPoints = newOwnList

		m.Save(u, false)
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (m *UserModel) GetEndPointowners(endpointID string) []*User {
	users := make([]*User, 0)

	for _, u := range m.List() {
		for _, epid := range u.OwnedEndPoints {
			if strings.EqualFold(epid, endpointID) {
				users = append(users, u)
			}
		}
	}

	return users
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (m *UserModel) Authenticate(email, password string) (*User, error) {
	// Retrieve the id and hashed password associated with the given email. If
	// no matching email exists we return the ErrInvalidCredentials error.
	user, err := m.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(strings.TrimSpace(password)))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		} else {
			return nil, err
		}
	}
	// Otherwise, the password is correct. Return the user ID.
	return user, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) Save(u *User, updatePasword bool) error {

	newUser := false
	if u.ID == "" {
		var id string = strings.ToUpper(uuid.NewString())
		u.ID = id
		newUser = true
		if u.MaxAllowedEndpoints == 0 {
			u.MaxAllowedEndpoints = 5
		}
	}

	if u.Token == "" {

		// 10 year token
		token, err := u.GetNewToken(24 * 3650 * time.Hour)
		if err != nil {
			return err
		}
		u.Token = token
	}

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}

		if updatePasword && strings.TrimSpace(u.Password) != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(u.Password)), 12)
			if err != nil {
				return err
			}

			u.HashedPassword = hashedPassword
		}

		// Generate ID for the user.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		//id, _ := bucket.NextSequence()
		//u.ID = int(id)
		// Marshal user data into bytes.
		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		u.Email = strings.ToLower(u.Email)

		// key = > user.name+ user.id
		key := u.ID // + string(itob(u.ID))

		return bucket.Put([]byte(key), buf)
	})

	if err == nil && newUser {
		// send verification email
	}

	return err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) Exists(u *User) bool {
	found := false
	for _, user := range m.List() {
		if strings.EqualFold(user.Email, u.Email) && !strings.EqualFold(user.ID, u.ID) {
			found = true
			break
		}
	}
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) Get(id string) (*User, error) {

	var userJson []byte // = make([]byte, 0)

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		key := id
		//strings.ToUpper(id)

		userJson = bucket.Get([]byte(key))

		return nil

	})

	user := User{}

	if userJson != nil {
		err := json.Unmarshal(userJson, &user)
		if err == nil {
			return &user, nil
		}
	}

	return nil, errors.New("User ID: Not found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) IsDuplicate(u *User) bool {

	u2, err := m.GetByEmail(u.Email)
	if err != nil {
		return false
	}

	if u2.ID != u.ID {
		return true
	}

	return false

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) GetByEmail(email string) (*User, error) {

	for _, u := range m.List() {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}

	return nil, errors.New("User email: Not found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) GetByUserName(username string) (*User, error) {

	for _, u := range m.List() {
		if strings.EqualFold(u.Name, username) {
			return u, nil
		}
	}

	return nil, errors.New("User name: Not found")

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) GetByToken(token string) (*User, error) {

	claims, err := jwtutils.Parse(token)

	if err != nil {
		// check non JWT toketns
		for _, u := range m.List() {
			if strings.EqualFold(u.Token, token) {
				return u, nil
			}
		}
		return nil, errors.New("User token: No User found")

	}

	usedID, found := claims["sub"]
	if !found {
		return nil, fmt.Errorf("invalid token")
	}

	usedIDString, ok := usedID.(string)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	user, err := m.Get(usedIDString)

	if err != nil {
		return nil, err
	}

	if user.Token != token {
		return nil, fmt.Errorf("invalid token")

	}

	return user, nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) List() []*User {
	users := make([]*User, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			user := User{}
			err := json.Unmarshal(v, &user)
			if err == nil {
				users = append(users, &user)
			}

		}

		return nil
	})
	return users

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) Delete(id string) error {

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		key := id //strings.ToUpper(id)
		dbDeleteError := bucket.Delete([]byte(key))
		return dbDeleteError
	})

	return err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) BuildVerificationEmail(user *User, host string) *EmailRequest {
	if !user.HasVerified {

		v, _ := m.CreateNewVerification(user, m.GetVerificationTableName())

		d := &AccountEmailTemplateData{
			User:            user,
			VerficationId:   v,
			VerficationLink: fmt.Sprintf("%s/user/verify/%s/%s", host, user.ID, v),
		}

		return &EmailRequest{
			To:       []string{user.Email},
			Subject:  "Please Verify your account",
			Body:     "",
			Template: "email_verify_email.tmpl",
			Data:     d,
		}

	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) BuildPasswordEmail(user *User, host string) *EmailRequest {

	v, _ := m.CreateNewVerification(user, m.GetPasswordResetTableName())

	d := &AccountEmailTemplateData{
		User:            user,
		VerficationId:   v,
		VerficationLink: fmt.Sprintf("%s/user/resetpwd/%s/%s", host, user.ID, v),
	}

	return &EmailRequest{
		To:       []string{user.Email},
		Subject:  "Password reset",
		Body:     "",
		Template: "email_password_reset.tmpl",
		Data:     d,
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) CreateNewVerification(u *User, table []byte) (string, error) {

	VerificationID := strings.ToUpper(uuid.NewString())

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(table)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(u.ID), []byte(VerificationID))
	})

	return VerificationID, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) Verify(u *User, verificationId string, table []byte) bool {

	userActualVerificationId := ""

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(table)
		if bucket == nil {
			return errors.New("table does not exits")
		}
		key := strings.ToUpper(u.ID)

		userActualVerificationId = string(bucket.Get([]byte(key)))

		return nil

	})

	return strings.EqualFold(verificationId, userActualVerificationId)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) DeleteVerificationRecord(u *User, table []byte) error {
	concurrent.Recoverer("DeleteVerificationRecord")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(table)
		if err != nil {
			return err
		}
		key := u.ID
		dbDeleteError := bucket.Delete([]byte(key))
		return dbDeleteError
	})

	// TODO ==> delete request and response params

	return err
}
