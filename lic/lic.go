package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hyperboloide/lk"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

type MyLicence struct {
	Client string    `json:"client"`
	Email  string    `json:"email"`
	End    time.Time `json:"end"`
}

type LicData struct {
	End         time.Time
	ExpiryHours float64
	ExpiryDays  int64
	Client      string
	ClientEmail string
}

type LicenseFile struct {
	Name            string
	Status          string
	ValidTill       time.Time
	AssignedTo      string
	AssignedToEmail string
	ExpiryDays      int64
}

// keep the length same
const MySecret string = "BXC&1*~U#6^#s0^=^^^7=c98"

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLicFileWithStatus() []*LicenseFile {
	licFiles := make([]*LicenseFile, 0)
	files, err := getLicFileList()
	if err != nil {
		return licFiles
	}

	for _, file := range files {
		finalFileName := fmt.Sprintf("lic/%s", file.Name())

		licFile := &LicenseFile{Name: finalFileName}
		log.Println("Checking...", file.Name())
		if file.IsDir() {
			licFile.Status = "Not a file"
			licFiles = append(licFiles, licFile)
			continue
		}

		if !strings.HasSuffix(file.Name(), ".lic") {
			licFile.Status = "Not a .lic file"
			licFiles = append(licFiles, licFile)
			continue
		}

		err := VerifyLicFile(finalFileName)
		if err == nil {
			licFile.Status = "VERIFIED"

			licData, err := GetLicFileExpiryDuration(finalFileName)
			if err == nil {
				licFile.ValidTill = licData.End
				licFile.ExpiryDays = licData.ExpiryDays
				licFile.AssignedTo = licData.Client
				licFile.AssignedToEmail = licData.ClientEmail
			}
			licFiles = append(licFiles, licFile)
			continue
		} else {
			licFile.Status = err.Error()
			licFiles = append(licFiles, licFile)
			continue
		}

	}

	return licFiles

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func getLicFileList() ([]os.DirEntry, error) {
	files, err := os.ReadDir("lic")
	if err != nil {
		return nil, err
	}
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	return files, nil
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func VerifyLicFiles() (string, error) {

	files, err := getLicFileList()
	if err != nil {
		return "", err
	}

	for _, file := range files {
		log.Println("Checking...", file.Name())
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".lic") {
			continue
		}

		finalFileName := fmt.Sprintf("lic/%s", file.Name())
		err := VerifyLicFile(finalFileName)
		if err == nil {
			log.Println("Processed file:", finalFileName, ". Lic verified.")
			return finalFileName, nil
		} else {
			log.Println("Processed file:", finalFileName, " Err:", err.Error())
		}

	}

	return "", errors.New("No valid Lic file found.")

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func VerifyLicFile(licKeyFile string) error {
	b, err := os.ReadFile(licKeyFile) // just pass the file name
	if err != nil {
		return err
	}

	fullLicString, err := stringutils.Decrypt(string(b), MySecret) // convert content to a 'string'
	if err != nil {
		return err
	}

	licKeyBroken := strings.Split(fullLicString, "\n")

	if len(licKeyBroken) != 2 {
		return errors.New("Invalid key stored. Can not break it")
	}

	pubKeyString := licKeyBroken[0]

	licKeyString := licKeyBroken[1]

	err = VerifyLic(pubKeyString, licKeyString)

	return err

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLicExpiry(licKeyFile string) string {
	b, err := os.ReadFile(licKeyFile) // just pass the file name
	if err != nil {
		return ""
	}

	fullLicString, err := stringutils.Decrypt(string(b), MySecret) // convert content to a 'string'
	if err != nil {
		return ""
	}

	licKeyBroken := strings.Split(fullLicString, "\n")

	if len(licKeyBroken) != 2 {
		return ""
	}

	//pubKeyString := licKeyBroken[0]

	licKeyString := licKeyBroken[1]

	_, message := CheckLicExpiry(licKeyString)

	return message
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func VerifyLic(publicKeyString string, licKeyString string) error {

	publicKey, err := lk.PublicKeyFromB64String(publicKeyString)
	if err != nil {
		return err

	}

	license, err := lk.LicenseFromB64String(licKeyString)
	if err != nil {
		return err

	}

	// validate the license:
	if ok, err := license.Verify(publicKey); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%s", "Invalid key")
	}

	expired, message := CheckLicExpiry(licKeyString)

	if expired {
		return errors.New(message)
	}
	return nil

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func CheckLicExpiry(licKeyString string) (expired bool, message string) {

	license, err := lk.LicenseFromB64String(licKeyString)
	if err != nil {
		return true, err.Error()

	}

	// unmarshal the document and check the end date:
	res := MyLicence{}
	if err := json.Unmarshal(license.Data, &res); err != nil {
		return true, err.Error()
	} else if res.End.Before(time.Now()) {
		return true, fmt.Sprintf("License expired on: %s", res.End.String())
	} else {
		return false, fmt.Sprintf(`Licensed to %s[%s] until %s`, res.Client, res.Email, res.End.Format("2006-01-02"))
	}

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLicFileExpiryDuration(licKeyFile string) (*LicData, error) {

	licData := &LicData{
		End:         time.Now().UTC(),
		ExpiryHours: 0,
		ExpiryDays:  0,
		Client:      "",
		ClientEmail: "",
	}

	b, err := os.ReadFile(licKeyFile) // just pass the file name
	if err != nil {
		return licData, err
	}

	// fullLicString := string(b) // convert content to a 'string'

	fullLicString, err := stringutils.Decrypt(string(b), MySecret) // convert content to a 'string'
	if err != nil {
		return licData, err
	}

	licKeyBroken := strings.Split(fullLicString, "\n")

	if len(licKeyBroken) != 2 {
		return licData, errors.New("Invalid key stored. Can not break it")
	}

	//pubKeyString := licKeyBroken[0]

	licKeyString := licKeyBroken[1]

	return GetLicExpiryDuration(licKeyString)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLicExpiryDuration(licKeyString string) (*LicData, error) {

	licData := &LicData{
		End:         time.Now().UTC(),
		ExpiryHours: 0,
		ExpiryDays:  0,
		Client:      "",
		ClientEmail: "",
	}
	license, err := lk.LicenseFromB64String(licKeyString)
	if err != nil {
		return licData, err

	}

	// unmarshal the document and check the end date:
	res := MyLicence{}
	if err := json.Unmarshal(license.Data, &res); err != nil {
		return licData, err
	}

	licData.ExpiryHours = res.End.Sub(time.Now().UTC()).Hours()

	licData.ExpiryDays = int64(licData.ExpiryHours / 24)

	licData.End = res.End
	licData.Client = res.Client
	licData.ClientEmail = res.Email

	return licData, nil

}
