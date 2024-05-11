package main

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/zerobit-tech/GoQhttp/utils/typeutils"
// )

// // ------------------------------------------------------------------
// //
// // ------------------------------------------------------------------
// type XmlService struct {
// 	Pgm Pgm `xml:"pgm"`
// }

// // ------------------------------------------------------------------
// // <pgm error="off" lib="sumitg1" name="QHTTPTEST2" var="QHTTPTEST2">
// //
// // ------------------------------------------------------------------
// type Pgm struct {
// 	Error   string `xml:"error,attr"`
// 	Lib     string `xml:"lib,attr"`
// 	Name    string `xml:"name,attr"`
// 	Var     string `xml:"var,attr"`
// 	Parms   []Parm `xml:"parm"`
// 	Success string `xml:"success"`
// }

// func (pgm *Pgm) ToJson() any {
// 	outMap := make(map[string]any)

// 	//dataMapList := make([]any, 0)
// 	for _, prm := range pgm.Parms {

// 		prm.ToJson(outMap)
// 		//dataMapList = append(dataMapList, dataMap)
// 	}

// 	return outMap

// }

// //------------------------------------------------------------------
// //
// //------------------------------------------------------------------

// // <parm io="both" var="in2">
// type Parm struct {
// 	Io   string `xml:"io,attr"`
// 	Var  string `xml:"var,attr"`
// 	Data []Data `xml:"data"` // value will be in data.Text
// 	Ds   []Ds   `xml:"ds"`
// }

// func (prm *Parm) ToJson(outMap map[string]any) {

// 	for _, data := range prm.Data {
// 		//key := fmt.Sprintf("%s", prm.Var)
// 		data.ToJson(outMap)

// 	}

// 	for _, nestedDS := range prm.Ds {

// 		outMap2 := make(map[string]any)
// 		key := fmt.Sprintf("%s", prm.Var)
// 		key = strings.ToUpper(key)
// 		nestedDS.ToJson(key, outMap2)

// 		val, found := outMap[key]

// 		if found {
// 			if typeutils.IsList(val) {
// 				valList, ok := val.([]any)
// 				if ok {
// 					valList = append(valList, outMap2)
// 				}
// 				outMap[key] = valList

// 			} else {
// 				valList := []any{val, outMap2}
// 				outMap[key] = valList

// 			}
// 		} else {
// 			outMap[key] = outMap2

// 		}

// 	}

// 	//var dataMapList2 any = dataMapList

// }

// // ------------------------------------------------------------------
// //
// // ------------------------------------------------------------------
// type Ds struct {
// 	Var  string `xml:"var,attr"`
// 	Data []Data `xml:"data"`
// 	Ds   []Ds   `xml:"ds"`
// }

// func (ds *Ds) ToJson(baseName string, outMap map[string]any) {

// 	//dataMapList := make([]any, 0)
// 	for _, data := range ds.Data {

// 		data.ToJson(outMap)

// 	}

// 	for _, nestedDS := range ds.Ds {
// 		outMap2 := make(map[string]any)
// 		key := fmt.Sprintf("%s", ds.Var)
// 		key = strings.ToUpper(key)
// 		nestedDS.ToJson(key, outMap2)

// 		val, found := outMap[key]

// 		if found {
// 			if typeutils.IsList(val) {
// 				valList, ok := val.([]any)
// 				if ok {
// 					valList = append(valList, outMap2)
// 				}
// 				outMap[key] = valList

// 			} else {
// 				valList := []any{val, outMap2}
// 				outMap[key] = valList

// 			}
// 		} else {
// 			outMap[key] = outMap2

// 		}
// 	}

// }

// // ------------------------------------------------------------------
// //
// // ------------------------------------------------------------------
// type Data struct {
// 	Type string `xml:"type,attr"`
// 	Var  string `xml:"var,attr"`
// 	Text string `xml:",chardata"`
// }

// func (data *Data) ToJson(outMap map[string]any) {
// 	key := fmt.Sprintf("%s", data.Var)

// 	key = strings.ToUpper(key)
// 	val, found := outMap[key]

// 	if found {
// 		if typeutils.IsList(val) {
// 			valList, ok := val.([]any)
// 			if ok {
// 				valList = append(valList, data.Text)
// 			}
// 			outMap[key] = valList

// 		} else {
// 			valList := []any{val, data.Text}
// 			outMap[key] = valList

// 		}
// 	} else {
// 		outMap[key] = data.Text

// 	}
// }
