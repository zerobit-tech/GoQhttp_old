package main

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	// Define the target APIs you want to load test
	targets := []string{
		"https://0.0.0.0:4081/uapi/v1/helloworld?NAME=a",
	}

	rate := vegeta.Rate{Freq: 20, Per: time.Second} // 20 requests per second
	duration := 30 * time.Minute                    // Duration of the load test

	var results vegeta.Results
	var wg sync.WaitGroup

	for _, target := range targets {
		target := target // Capture range variable for goroutine

		wg.Add(1)
		go func() {
			defer wg.Done()
			res := loadTest(target, rate, duration)
			for _, r := range res {
				results.Add(&r)
			}

		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Calculate the summary of the load test

	//metrics := vegeta.NewMetrics(results)
	var metrics vegeta.Metrics
	for _, res := range results {
		metrics.Add(&res)
	}

	metrics.Close()

	fmt.Printf("Mean Response Time: %.2f ms\n", metrics.Latencies.Mean.Seconds()*1000)
	fmt.Printf("Requests per second: %.2f\n", metrics.Rate)
	fmt.Printf("99th Percentile Response Time: %.2f ms\n", metrics.Latencies.P99.Seconds()*1000)
	fmt.Printf("Total Requests: %d\n", metrics.Requests)
	fmt.Printf("Total Successful Requests: %f\n", metrics.Success)
	fmt.Printf("Total Failed Requests: %s\n", metrics.Errors)

	// Save the results to a file
	// reportFile, err := os.Create("load_test_results.bin")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer reportFile.Close()

	// enc := vegeta.NewEncoder(reportFile)
	// for _, result := range results {
	// 	if err := enc.Encode(&result); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
}

func loadTest(target string, rate vegeta.Rate, duration time.Duration) vegeta.Results {
	fmt.Println("Running...", target)
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    target,
	})

	tls := vegeta.TLSConfig(&tls.Config{InsecureSkipVerify: true})

	attacker := vegeta.NewAttacker(tls)

	var results vegeta.Results
	for res := range attacker.Attack(targeter, rate, duration, "Load Test") {
		results.Add(res)
	}

	return results
}
