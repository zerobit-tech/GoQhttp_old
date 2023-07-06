package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

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
		log.Panicf(fmt.Sprintf("Please make sure ODBC driver '%s' for IBM I are installed", driverName))
	default:
		log.Printf("Can not validate the ODBC drivers. Please make sure ODBC driver '%s' for IBM I are installed \n", driverName)
	}
}

func checkWindowsOdbcDriver(driverName string) int {

	return 0

}

func checkLinuxOdbcDriver(driverName string) int {
	cmd := exec.Command("bash", "-c", "odbcinst -q -d")
	out, err := cmd.Output()
	if err != nil {
		// if there was any error, print it here
		fmt.Println("could not run command: ", err)
		return 2 // can not validate driver
	}
	// otherwise, print the output from running the command
	fmt.Println("Output: ", string(out))

	for _, d := range strings.Split(string(out), "\n") {
		if d == driverName || d == fmt.Sprintf("[%s]", driverName) {
			return 0 // all good
		}
	}

	return 1 // driver not found
}
