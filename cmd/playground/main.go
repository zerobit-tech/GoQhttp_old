package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

func main() {

	var wg sync.WaitGroup
	for i := 0; i <= 20; i++ {
		wg.Add(1)
		go main2(&wg)
	}

	wg.Wait()
}
func main2(wg *sync.WaitGroup) {
	defer concurrent.Recoverer("main2")
	defer wg.Done()

	//time.Sleep(1 * time.Second)

	url := "https://localhost:4081/api/spchar"
	method := "POST"

	payload := strings.NewReader(` 	

{
  "CLOBFIELD": "a",
  "IN_OUT_CHAR_FIELD": "b",
  "IOCLOBFIELD": "d",
  "IOVARCHARFIELD": "e",
  "VARCHARFIELD": "f",
  "XXCCXC": "g"
}`)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// t := &http.Transport{
	// 	Dial: (&net.Dialer{
	// 		Timeout:   60 * time.Second,
	// 		KeepAlive: 30 * time.Second,
	// 	}).Dial,
	// 	ResponseHeaderTimeout: time.Hour,
	// 	MaxConnsPerHost:       99999,
	// 	DisableKeepAlives:     true,
	// 	// We use ABSURDLY large keys, and should probably not.
	// 	TLSHandshakeTimeout: 60 * time.Second,
	// 	//InsecureSkipVerify:  true,
	// }
	// t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// c := &http.Client{
	// 	Transport: t,
	// }

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c := &http.Client{}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "OWVmYTI0NDcyZWNkNTc0NDdjNTkwMWVmN2ZmYTZkMDI4OWJhMTI3MmYxNGNkMjg1ODZhZG1pbjJAZXhhbXBsZS5jb20=ca7ba3a96b2cb49746879d025f9077c8b631eb02bdfa4bafc9")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	// fmt.Println("res", res.StatusCode)
	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(body))
}
