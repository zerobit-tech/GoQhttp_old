package responseprocessor

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type PgmE struct {
	Error      []Error    `xml:"error"`
	Version    string     `xml:"version"`
	JobInfo    JobInfo    `xml:"jobinfo"`
	JobLogScan JobLogScan `xml:"joblogscan"`
	JobLog     string     `xml:"joblog" json:"-" `
	JobLogs    []string
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

// ------------------------------------------------------------------
//
// ------------------------------------------------------------------
func ProcessErrorXML(xmlData string) (any, string, error) {

	var result XmlServiceE
	err := xml.Unmarshal([]byte(xmlData), &result)
	if err != nil {
		fmt.Printf("Error unmarshalling XML: %v\n", err)
		return nil, "", err
	}

	errorMessage := "Error"
	s1 := result.Pgm.JobLogScan.JobLogRec
	if len(s1) > 1 {
		z1 := s1[len(s1)-1]
		errorMessage = z1.Jobcpf
	}

	result.Pgm.JobLogs = strings.Split(result.Pgm.JobLog, "\n")

	return result.Pgm, errorMessage, nil

}
