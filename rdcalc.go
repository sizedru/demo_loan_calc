package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sizedru/rdcalc/rd"
)


type RDOst struct {
	orderNum   string
	date       string
	toBase     float64
	toPerc     float64
	toFine     float64
	toFine1    float64
	toFine2    float64
	toFDay     int
	orderId    int
	pdn        float64
	prolongCnt int
}

type RDOstArray []RDOst

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func Order(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, "Error!")
		}
	}()

	query := r.URL.Query()
	order, op := query["order"]
	if !op || len(order) == 0 {
		fmt.Fprintln(w, "Orders")
	} else {
		var ostDate string
		date, dp := query["date"]
		if dp && len(order) > 0 {
			ostDate = date[0]
		}
		ost, _, _, _ := rd.Ost(order[0], ostDate, false, nil, false, false, false, false)

		fmt.Fprintln(w, ost)
	}

}

func OrderJson(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, "Error!")
		}
	}()

	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	order, op := query["order"]
	if !op || len(order) == 0 {
		fmt.Fprintln(w, "Orders")
	} else {
		var ostDate string
		date, dp := query["date"]
		if dp && len(date) > 0 {
			ostDate = date[0]
		}

		var payCorrect bool
		payCorrect = false
		correct, cp := query["correct"]
		if cp && len(correct) > 0 {
			if correct[0] == "1" {
				payCorrect = true
			}
		}

		ost, s, _, err := rd.Ost(order[0], ostDate, false, nil, payCorrect, false, false, false)

		B := fmt.Sprintf("%.2f", ost.Base)
		BN := fmt.Sprintf("%.2f", s.CalcProfit.Base)
		P := fmt.Sprintf("%.2f", ost.Perc)
		F1 := fmt.Sprintf("%.2f", ost.FOne)
		F2 := fmt.Sprintf("%.2f", ost.FTwo)
		D := fmt.Sprintf("%.2f", ost.Duty)
		FD := strconv.Itoa(ost.FDay)
		E := "0"
		if err {
			E = "1"
		}
		XXXR := "0"
		if s.XXLimitReach {
			XXXR = "1"
		}

		FDAbs := strconv.Itoa(s.FineDaysAbsolute)

		fmt.Println(ost)
		json.NewEncoder(w).Encode(map[string]string{"Base": B, "BaseNeed": BN, "Perc": P, "FOne": F1, "FTwo": F2, "Duty": D, "FDay": FD,
			"Err": E, "XXXDate": s.XXXDate.Format(rd.ISO), "XXX": XXXR, "FDAbs": FDAbs})
	}

}

func OrderFullJson(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, "Error!")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query()
	order, op := query["order"]
	if !op || len(order) == 0 {
		fmt.Fprintln(w, "Orders")
	} else {
		var ostDate string
		date, dp := query["date"]
		if dp && len(date) > 0 {
			ostDate = date[0]
		}

		var payCorrect bool
		payCorrect = false
		correct, cp := query["correct"]
		if cp && len(correct) > 0 {
			if correct[0] == "1" {
				payCorrect = true
			}
		}

		ost, s, all, err := rd.Ost(order[0], ostDate, false, nil, payCorrect, false, true, false)

		_ = err
		_ = s

		json.NewEncoder(w).Encode(all)
	}

}

func OrderLog(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(w, "Error!")
		}
	}()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<html>")
	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<pre>")

	query := r.URL.Query()
	order, op := query["order"]
	if !op || len(order) == 0 {
		fmt.Fprintln(w, "Orders")
	} else {
		var ostDate string
		date, dp := query["date"]
		if dp && len(date) > 0 {
			ostDate = date[0]
		}
		var everyMonthOn bool
		emo, dp := query["emo"]
		if dp && len(emo) > 0 {
			everyMonthOn = true
		}

		ost, _, _, err := rd.Ost(order[0], ostDate, true, w, false, false, false, everyMonthOn)

		fmt.Fprintln(w, ost)

		if err {
			fmt.Fprintln(w, "Есть ошибки распределения оплат!!!!")
		}
	}

	fmt.Fprintln(w, "</pre>")
	fmt.Fprintln(w, "</body>")
	fmt.Fprintln(w, "</html>")

}


func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/", Index)
	router.HandleFunc("/order", Order)
	router.HandleFunc("/orderlog", OrderLog)
	router.HandleFunc("/orderjson", OrderJson)
	router.HandleFunc("/orderfulljson", OrderFullJson)

	router.Use(loggingMiddleware)

	log.Fatal(http.ListenAndServe(":8008", router))

}
