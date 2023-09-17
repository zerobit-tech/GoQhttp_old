package main

import (
	"fmt"

	"github.com/onlysumitg/GoQhttp/internal/rpg"
)

func main() {

	p1 := &rpg.Param{
		Seq:             1,
		Name:            "P1",
		DataType:        "ZONED",
		Length:          5,
		DecimalPostions: 0,
		IsDs:            false,
		DsDim:           0,
		IsVarying:       false,
	 
	}
	p2 := &rpg.Param{
		Seq:             2,
		Name:            "P2",
		DataType:        "ZONED",
		Length:          5,
		DecimalPostions: 0,
		IsDs:            false,
		DsDim:           0,
		IsVarying:       false,
	 
	}

	p3 := &rpg.Param{
		Seq:             2,
		Name:            "P3",
		DataType:        "ZONED",
		Length:          5,
		DecimalPostions: 0,
		IsDs:            false,
		DsDim:           0,
		IsVarying:       false,
	 
	}

	pgm := rpg.Program{

		Name:       "QHTTPTEST1",
		Lib:        "SUMITG1",
		Parameters: []*rpg.Param{p1, p2, p3},
	}

	//fmt.Println(pgm.ToXML())
}
