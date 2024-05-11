package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type PgmE struct {
	Error      []Error    `xml:"error"`
	Version    string     `xml:"version"`
	JobInfo    JobInfo    `xml:"jobinfo"`
	JobLogScan JobLogScan `xml:"joblogscan"`
	JobLog     string     `xml:"joblog"`
}

type Error struct {
	Errnoxml  string `xml:"errnoxml"`
	Xmlerrmsg string `xml:"xmlerrmsg"`
	Xmlhint   string `xml:"xmlhint"`
	Text      string `xml:",chardata"`
}

type JobInfo struct {
	Jobipc      string `xml:"jobipc"`
	Jobipcskey  string `xml:"jobipcskey"`
	JobName     string `xml:"jobname"`
	JobUser     string `xml:"jobuser"`
	JobNumber   string `xml:"jobnbr"`
	Jobsts      string `xml:"jobsts"`
	CurrentUser string `xml:"curuser"`
	Ccsid       string `xml:"ccsid"`
	Dftccsid    string `xml:"dftccsid"`
	Paseccsid   string `xml:"paseccsid"`
	Langid      string `xml:"langid"`
	Cntryid     string `xml:"cntryid"`
	Sbsname     string `xml:"sbsname"`
	Sbslib      string `xml:"sbslib"`
	CurrentLib  string `xml:"curlib"`
	Syslibl     string `xml:"syslibl"`
	Usrlibl     string `xml:"usrlibl"`
	Jobcpffind  string `xml:"jobcpffind"`
}

type JobLogScan struct {
	JobLogRec []JobLogRec `xml:"joblogrec"`
}

type JobLogRec struct {
	Jobcpf  string `xml:"jobcpf"`
	Jobtime string `xml:"jobtime"`
	Jobtext string `xml:"jobtext"`
}

type XmlServiceE struct {
	Pgm PgmE `xml:"pgm"`
}

func main() {

	var result XmlServiceE
	if err := xml.Unmarshal([]byte(xmlDataError), &result); err != nil {
		fmt.Printf("Error decoding XML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Version: %s\n", result.Pgm.Version)
	fmt.Printf("Error Messages:\n")
	for _, errMsg := range result.Pgm.Error {
		fmt.Printf("  - %s\n", errMsg)
	}

}
