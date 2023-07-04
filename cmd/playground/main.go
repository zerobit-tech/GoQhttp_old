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
	for i := 0; i <= 10000; i++ {
		wg.Add(1)
		go main2(&wg)
	}

	wg.Wait()
}
func main2(wg *sync.WaitGroup) {
	defer concurrent.Recoverer("main2")
	defer wg.Done()

	//time.Sleep(1 * time.Second)

	url := "https://0.0.0.0:4081/api/spparm"
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

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "M2MxODIyNmE3MmNmYjg2NmVlOWE4Y2RkMDdkMzZiYWFjNjMzMWRlNjIyODhhN2ViYjdkZXY3MjBAZXhhbXBsZS5jb20=994df748b305ea066d8ae4cd9bb2e22ad612667cdc5673fddd")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	fmt.Println("res", res.StatusCode)
	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(body))
}
