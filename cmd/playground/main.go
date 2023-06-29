package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

func main() {

	var wg sync.WaitGroup
	for i := 0; i <= 100; i++ {
		wg.Add(1)
		go main2(&wg)
	}

	wg.Wait()
}
func main2(wg *sync.WaitGroup) {
	defer wg.Done()

	//time.Sleep(1 * time.Second)

	url := "https://0.0.0.0:4081/api/testchar"
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
	req.Header.Add("Authorization", "MGFjNjRmZWMzOTdmMTVkM2I5OGVmZTcxOTY2ZTFmYzg4YWNjODM4ZjhkMTY1NWE3YTdhZG1pbjJAZXhhbXBsZS5jb20=9969d9b5a54b5ba442c58c772c88eaf07efbbe648f850685ec")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
