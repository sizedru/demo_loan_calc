package rd

import (
	"fmt"
	"log"
	"math"
	"net/http"
)

// IsFail check error
func IsFail(err error, s string) {
	if err != nil {
		log.Panic(s, " ", err)
		//panic(err)
	}
}

func Prn(a ...interface{}) {
	fmt.Println(a...)
}

func fmtPrintln(debug bool, w http.ResponseWriter, a ...interface{}) {
	if debug {
		if w == nil {
			Prn(a...)
		} else {
			fmt.Fprintln(w, a...)
		}
	}
}

// B2I bool to int
func B2I(b bool) int {
	if b {
		return 1
	}
	return 0
}

// I2B bool to int
func I2B(i int) bool {
	if i == 0 {
		return false
	}
	return true
}

// Round правильное округление
func Round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}
	return rounder / pow
}

func MinusF(a, b float64) (aa, bb, cc float64) {
	cc = 0
	if a >= b {
		cc = b
		aa = a - b
		bb = 0
	} else {
		cc = a
		aa = 0
		bb = b - a
	}
	return
}

func MinusI(a, b int) (aa, bb, cc int) {
	cc = 0
	if a >= b {
		cc = b
		aa = a - b
		bb = 0
	} else {
		cc = a
		aa = 0
		bb = b - a
	}
	return
}
