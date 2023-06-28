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
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

type MyLicence struct {
	Client string    `json:"client"`
	Email  string    `json:"email"`
	End    time.Time `json:"end"`
}

type LicenseFile struct {
	Name       string
	Status     string
	ValidTill  time.Time
	ExpiryDays int64
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

			validTill, _, expiryDays, err := GetLicFileExpiryDuration(finalFileName)
			if err == nil {
				licFile.ValidTill = validTill
				licFile.ExpiryDays = expiryDays
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
func GetLicFileExpiryDuration(licKeyFile string) (time.Time, float64, int64, error) {
	b, err := os.ReadFile(licKeyFile) // just pass the file name
	if err != nil {
		return time.Now().UTC(), 0, 0, err
	}

	// fullLicString := string(b) // convert content to a 'string'

	fullLicString, err := stringutils.Decrypt(string(b), MySecret) // convert content to a 'string'
	if err != nil {
		return time.Now().UTC(), 0, 0, err
	}

	licKeyBroken := strings.Split(fullLicString, "\n")

	if len(licKeyBroken) != 2 {
		return time.Now().UTC(), 0, 0, errors.New("Invalid key stored. Can not break it")
	}

	//pubKeyString := licKeyBroken[0]

	licKeyString := licKeyBroken[1]

	return GetLicExpiryDuration(licKeyString)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLicExpiryDuration(licKeyString string) (time.Time, float64, int64, error) {
	license, err := lk.LicenseFromB64String(licKeyString)
	if err != nil {
		return time.Now().UTC(), 0, 0, err

	}

	// unmarshal the document and check the end date:
	res := MyLicence{}
	if err := json.Unmarshal(license.Data, &res); err != nil {
		return time.Now().UTC(), 0, 0, err
	}

	expiryHours := res.End.Sub(time.Now().UTC()).Hours()

	expiryDays := int64(expiryHours / 24)
	return res.End, expiryHours, expiryDays, nil

}
