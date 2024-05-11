package main

import (
	"fmt"

	"github.com/zerobit-tech/GoQhttp/utils/templateutil"
)

func main() {
	tCache, err := templateutil.NewTemplateCache()

	if err != nil {
		fmt.Println("ERROR::", err)
		return
	}

	for k, v := range tCache {
		fmt.Println(">>", k, v.Name())
	}

	m := map[string]any{"MESSAGE": "hellloo"}
	s, err := templateutil.TemplateToString(tCache, "test.html", map[string]any{"Data": m})
	if err != nil {
		fmt.Println("ERROR2::", err)
		return
	}

	fmt.Println("Data::", s)

	// s, err := templateutil.TemplateToString(tCache, "base_test.html", map[string]any{"name": "sumit"})
	// if err != nil {
	// 	fmt.Println("ERROR2::", err)
	// 	return
	// }

	// fmt.Println("Data::", s)

}
