package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func askToContinue() {
	cont := "Y"
	fmt.Print("Do you want to continue....(Y/N): ")
	fmt.Scanln(&cont)
	cont = strings.ToUpper(cont)

	if cont == "N" || cont == "NO" {
		os.Exit(2)
	}
}

// -----------------------------------------------------------------------------
//
// -----------------------------------------------------------------------------
func validateSetup() {
	driverName := "IBM i Access ODBC Driver"

	odbcStatus := 0
	if runtime.GOOS == "windows" {
		odbcStatus = checkWindowsOdbcDriver(driverName)
	} else {
		odbcStatus = checkLinuxOdbcDriver(driverName)
	}

	switch odbcStatus {
	case 0:
		log.Println("Valid ODBC drivers found")
	case 1:
		log.Printf("Please make sure '%s' is  installed \n", driverName)
		log.Println("For more details please check: https://www.ibm.com/support/pages/odbc-driver-ibm-i-access-client-solutions")

		askToContinue()
	default:
		log.Printf("Can not validate the ODBC drivers. Please make sure '%s' is installed \n", driverName)
		log.Println("For more details please check: https://www.ibm.com/support/pages/odbc-driver-ibm-i-access-client-solutions")
		askToContinue()

	}
}

// -----------------------------------------------------------------------------
//
// -----------------------------------------------------------------------------
func checkWindowsOdbcDriver(driverName string) int {
	log.Println("Validating Windows setup.")

	cmd := exec.Command("powershell", "-NoProfile", "Get-OdbcDriver | Format-Table name, platform -AutoSize")
	out, err := cmd.Output()
	if err != nil {
		// if there was any error, print it here
		fmt.Println("could not run command: ", err)
		return 2 // can not validate driver
	}
	// otherwise, print the output from running the command

	for _, d := range strings.Split(string(out), "\n") {
		//fmt.Println(":::>>", d)
		d = strings.TrimSpace(d)
		if d == driverName || d == fmt.Sprintf("[%s]", driverName) {
			return 0 // all good
		}
		if strings.HasPrefix(d, driverName) && strings.HasSuffix(d, "64-bit") {
			return 0
		}
	}

	return 1 // driver not found

}

// -----------------------------------------------------------------------------
//
// -----------------------------------------------------------------------------
func checkLinuxOdbcDriver(driverName string) int {

	log.Println("Validating Linux setup.")
	cmd := exec.Command("bash", "-c", "odbcinst -q -d")
	out, err := cmd.Output()
	if err != nil {
		// if there was any error, print it here
		fmt.Println("could not run command: ", err)
		return 2 // can not validate driver
	}
	// otherwise, print the output from running the command

	for _, d := range strings.Split(string(out), "\n") {
		d = strings.TrimSpace(d)

		if d == driverName || d == fmt.Sprintf("[%s]", driverName) {
			return 0 // all good
		}

		if strings.HasPrefix(d, driverName) && strings.HasSuffix(d, "64-bit") {
			return 0
		}
	}

	return 1 // driver not found
}
