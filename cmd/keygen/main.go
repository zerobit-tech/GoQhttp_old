package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hyperboloide/lk"
	"github.com/zerobit-tech/GoQhttp/lic"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

func main() {
	waitChan := make(chan int)

	err := os.MkdirAll("./lic", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	//--------------------------------------- Setup CLI paramters ----------------------------
	params := &parameters{}
	params.Load()

	if params.checkemail {
		go ReadEmails(waitChan)

	} else {
		time.Sleep(2 * time.Second)
		go func() {
			waitChan <- 2
		}()

		err = params.Validate()
		if err != nil {
			log.Println(err)
		} else {
			processLicRequest(params)

		}
	}

	// file, err := lic.VerifyLicFiles()
	// if err == nil {
	// 	fmt.Println("Expirt::::", lic.GetLicExpiry(file))
	// }
	// return

	<-waitChan

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func processLicRequest(params *parameters) error {

	licData := &lic.MyLicence{
		Client: params.client,
		Email:  params.email,
		End:    time.Now().UTC().Add(time.Hour * 24 * time.Duration(params.days)),
	}

	licKeyFile := generateNewLic(licData)

	fmt.Println(licKeyFile)
	err := lic.VerifyLicFile(licKeyFile)
	if err != nil {
		fmt.Println("final Error:::", err)
		return err
	}

	b, err := os.ReadFile(licKeyFile) // just pass the file name
	if err != nil {
		fmt.Println("final Error: 2::", err)

		return err
	}

	app := baseAppConfig(*params)
	fmt.Println("New lic is ready:", licKeyFile)

	emailBody := fmt.Sprintf("Please create a new qhttp.lic file in lic folder and copy the following string in that file. <br><br> %s", string(b))

	email := &EmailRequest{
		To:       []string{params.email},
		Subject:  "QHTTP Lic Key",
		Body:     emailBody,
		Template: "",
		Data:     "",
	}

	app.SendEmail(email)
	return nil
}

//------------------------------------------------------
//
//------------------------------------------------------

func generateNewLic(licData *lic.MyLicence) string {
	privateKeyFile := fmt.Sprintf("lic/%s.prvt", "master")
	licKeyFile := fmt.Sprintf("lic/%s_%s.lic", time.Now().UTC().Format("20060102_150405000000"), strings.ToUpper(licData.Client))

	if !fileExists(privateKeyFile) {
		err := generateNewPrivateKey(privateKeyFile)
		if err != nil {
			log.Fatal("Error genereating new private key:", err)

		}
	}

	b, err := os.ReadFile(privateKeyFile) // just pass the file name
	if err != nil {
		log.Fatal(err)
	}

	privateKeyString := string(b) // convert content to a 'string'

	//fmt.Println(privateKeyString) // print the content as a 'string'

	err = generateNewLicFile(privateKeyString, licKeyFile, licData)
	if err != nil {
		log.Fatal(err)
	}
	// err = verifyLic(publicKeyString, licString)
	// if err != nil {
	// 	log.Fatal("Error verifing lic:", err)
	// }

	// expired, message := checkLicExpiry(licString)

	// fmt.Println("checkLicExpiry::", expired, message)

	return licKeyFile
}

//------------------------------------------------------
//
//------------------------------------------------------

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		fmt.Println("fileExists err:", err.Error())
	}
	return !info.IsDir()
}

//------------------------------------------------------
//
//------------------------------------------------------

func generateNewPrivateKey(privateKeyFile string) error {
	f, err := os.Create(privateKeyFile)
	if err != nil {
		return err

	}

	privateKey, err := lk.NewPrivateKey()
	if err != nil {
		return err

	}

	privateKeyString, err := privateKey.ToB64String()
	if err != nil {
		return err

	}

	_, err = f.Write([]byte(privateKeyString))
	if err != nil {
		return err
	}

	f.Sync()
	f.Close()

	return nil
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func generateNewLicFile(privateKeyString string, licFileName string, licData *lic.MyLicence) error {
	if !fileExists(licFileName) {
		f, err := os.Create(licFileName)
		if err != nil {
			return err

		}

		licString, err := generateLic(privateKeyString, licData)
		if err != nil {
			return err
		}

		publicKeyString, err := generatePublicKey(privateKeyString)
		if err != nil {
			return err
		}

		err = lic.VerifyLic(publicKeyString, licString)
		if err != nil {
			return err
		}

		finalLic := fmt.Sprintf("%s\n%s", publicKeyString, licString)

		finalLic, err = stringutils.Encrypt(finalLic, lic.MySecret)
		if err != nil {
			return err
		}

		_, err = f.Write([]byte(finalLic))
		if err != nil {
			return err
		}

		f.Sync()
		f.Close()
		return nil
	}

	return errors.New("File already exists")
}

//------------------------------------------------------
//
//------------------------------------------------------

func generateLic(privateKeyString string, licData *lic.MyLicence) (string, error) {
	privateKey, err := lk.PrivateKeyFromB64String(privateKeyString)
	if err != nil {
		return "", err

	}

	// marshall the document to json bytes:
	docBytes, err := json.Marshal(licData)
	if err != nil {
		return "", err
	}

	// generate your license with the private key and the document:
	license, err := lk.NewLicense(privateKey, docBytes)
	if err != nil {
		return "", err

	}

	// encode the new license to b64, this is what you give to your customer.
	licenseString, err := license.ToB64String()
	if err != nil {
		return "", err

	}
	return licenseString, nil
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func generatePublicKey(privateKeyString string) (string, error) {
	privateKey, err := lk.PrivateKeyFromB64String(privateKeyString)
	if err != nil {
		return "", err

	}

	// get the public key. The public key should be hardcoded in your app
	// to check licences. Do not distribute the private key!
	publicKey := privateKey.GetPublicKey()

	publicKeyString := publicKey.ToB64String()

	return publicKeyString, nil

}
