package rd

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"
)

func ApplyLgotPeriods(percentPeriod int, allDay int, oInfo *OrderInfo) int {
	if oInfo.LgotPeriod > 0 {
		if oInfo.LgotPeriodEnd {
			lp := allDay - oInfo.Days + oInfo.LgotPeriod
			if lp > 0 {
				if lp > oInfo.LgotPeriod {
					lp = oInfo.LgotPeriod
				}
				p := 0
				percentPeriod, lp, p = MinusI(percentPeriod, lp)
				oInfo.LgotPeriod -= p
			}
		} else if allDay > 1 {
			percentPeriod, oInfo.LgotPeriod, _ = MinusI(percentPeriod, oInfo.LgotPeriod)
		}
	}
	if oInfo.LgotPeriodPayFirstDay > 0 && allDay > 0 {
		percentPeriod, oInfo.LgotPeriodPayFirstDay, _ = MinusI(percentPeriod, oInfo.LgotPeriodPayFirstDay)
	}
	return percentPeriod
}

func PayRecursive(oS OrderPaymentSchedule, cps *OrderCheckpoints, cpsIndex int, debug bool, w http.ResponseWriter) (OrderPaymentSchedule, int) {

	fmtPrintln(debug, w, "")
	fmtPrintln(debug, w, ">>>>>>>>>>>>>>>>>>>>>")
	fmtPrintln(debug, w, "")

	s := oS.Copy()

	// ..............
    // ..............
    // ..............
    // ..............
    // ..............
    // ..............

		var flag int
		s, flag = PayRecursive(s, cps, cpsIndex+1, debug, w)

		if flag == 1 {
			if !IStartPP {
				(*cps)[cpsIndex].IsPP = false
				return oS, 1 // Исходное состояние возвращаем
			} else {
				(*cps)[cpsIndex].IsPP = false
				s.IsPP = false
				s, flag = PayRecursive(s, cps, cpsIndex+1, debug, w)
			}
		}

		if flag == 2 {
			if !IStartHold {
				(*cps)[cpsIndex].IsHold = false
				return oS, 2 // Исходное состояние возвращаем
			} else {
				(*cps)[cpsIndex].IsHold = false
				s.IsHold = false
				s, flag = PayRecursive(s, cps, cpsIndex+1, debug, w)
			}
		}

	}
	return s, 0
}

func Ost(OrderNum string, QueryDateStr string, debug bool, w http.ResponseWriter, payCorrect bool, ifTodayProlong bool, fullJson bool, everyMonthOn bool) (OrderMoneyKit, OrderPaymentSchedule, map[string]interface{}, bool) {

	start := time.Now()

	var QueryDate time.Time

	if QueryDateStr != "" {
		QueryDate = Date(QueryDateStr)
	} else {
		Now := time.Now()
		QueryDate = Date(Now.Format(ISO))
	}

	if payCorrect {
		Now := time.Now()
		QueryDate = Date(Now.Format(ISO))
	}

	hostname, err := os.Hostname()
	_ = hostname
	//fmt.Println(hostname)

	var db *sql.DB
    db = InitDB("ro", "", "", "localhost")

	defer CloseDB(db)

	oInfo := NewOrderInfo()

	err = oInfo.GetOrder(db, OrderNum)
	IsFail(err, "Query GetOrder fail")

	duration1 := time.Since(start)

	fmtPrintln(debug, w, oInfo.ID)
	fmtPrintln(debug, w, oInfo.Num)
	fmtPrintln(debug, w, "Аннуитет", oInfo.Annuity)
	fmtPrintln(debug, w, "Period", oInfo.Period)
	fmtPrintln(debug, w, "Date", oInfo.Date.Format(ISO))
	fmtPrintln(debug, w, "PercentRate", oInfo.PercentRate)
	fmtPrintln(debug, w, "Fine1Rate", oInfo.PercentFine1)
	fmtPrintln(debug, w, "Fine2Rate", oInfo.PercentFine2)
	fmtPrintln(debug, w, "BillingNumber", oInfo.BillingNumber)
	fmtPrintln(debug, w, "XXLimit", oInfo.XXLimit)
	fmtPrintln(debug, w, "XXLimitSumm", oInfo.XXLimitSumm)
	fmtPrintln(debug, w, "CalcCorrectionFlag", oInfo.CalcCorrectionFlag)

	//goland:noinspection GrazieInspection
	if QueryDate.Before(oInfo.Date) {
		QueryDate = oInfo.Date
	}

	QueryDatePrn := QueryDate.Format(ISO)
	_ = QueryDatePrn

	start = time.Now()
	cps, err := oInfo.GetOrderCheckpoints(db, QueryDate, true, everyMonthOn)
	IsFail(err, "Query GetCheckpoints fail")
	duration2 := time.Since(start)

	cps.Print(debug, w, oInfo.Date)

	if ifTodayProlong {
		for r := range cps {
			if cps[r].Date.Equal(QueryDate) {
				if !cps[r].IsAdag {
					cps[r].IsAdag = true
					cps[r].IsAdagDays = oInfo.Days
				}
			}
		}
		cps.Print(debug, w, oInfo.Date)
	}

	start = time.Now()
	oS, _ := GetSchedule(oInfo)
	duration3 := time.Since(start)
	oS.Print(debug, w, false)

	start = time.Now()
	var s OrderPaymentSchedule
	if payCorrect { // При корректировке платежей сначала надо очистить в cps все ненужные заморозки и простипени
		s, _ = PayRecursive(oS, &cps, 0, debug, w)

		cpsTwo, err := oInfo.GetOrderCheckpoints(db, QueryDate, false, false)
		IsFail(err, "Query GetCheckpoints fail")

		if len(cps) != len(cpsTwo) {
			IsFail(err, "Query GetCheckpoints LEN fail")
		}

		for i := range cps {
			if cps[i].Date != cpsTwo[i].Date {
				IsFail(err, "Query GetCheckpoints DATE fail")
			}
			cpsTwo[i].IsHold = cps[i].IsHold
			cpsTwo[i].IsPP = cps[i].IsPP
		}

		if QueryDateStr != "" {
			QueryDate = Date(QueryDateStr)

			var tmpCps OrderCheckpoints
			tmpEqual := false
			for _, v := range cpsTwo {
				tmpDate := v.Date
				tmpDatePrn := v.Date.Format(ISO)
				_ = tmpDatePrn
				if tmpDate.Before(QueryDate) {
					tmpCps = append(tmpCps, v)
				} else if tmpDate.Equal(QueryDate) {
					tmpCps = append(tmpCps, v)
					tmpEqual = true
				}
			}
			if len(tmpCps) > 0 && tmpEqual {
				s, _ = PayRecursive(oS, &tmpCps, 0, debug, w)
			} else {
				IsFail(err, "Query GetCheckpoints for Pay Correct fail")
			}
		}
	} else {
		s, _ = PayRecursive(oS, &cps, 0, debug, w)
	}
	duration4 := time.Since(start)

	fmtPrintln(debug, w, "Запрос условий займа", duration1)
	fmtPrintln(debug, w, "Запрос контрольных точек", duration2)
	fmtPrintln(debug, w, "Расчет графика", duration3)
	fmtPrintln(debug, w, "Расчет остатков", duration4)
	fmtPrintln(debug, w)
	fmtPrintln(debug, w)

	s.FDayHistoryCB.Print(debug, w)
	fmtPrintln(debug, w, s.ClientScheduleDate.Format(ISO))
	s.Print(debug, w, true)
	fmtPrintln(debug, w, oInfo.Num, s.AllCalcDay, QueryDate.Format(ISO), s.Schedule[0].BalanceBefore.Base, s.Schedule[0].BalanceBefore.Perc, Round(s.Schedule[0].BalanceBefore.Fine(), 2), s.Schedule[0].BalanceBefore.FDay)

	PlusFDay := 0
	PlusFlag := 0
	if s.FDayHistoryCB[0].Cnt > 0 {
		if s.AllCalcDay <= 365 {
			for _, v := range s.FDayHistoryCB {
				PlusFDay += v.Cnt
			}
			PlusFlag = 1
		} else {
			MinDay := s.AllCalcDay - 365
			for _, v := range s.FDayHistoryCB {
				if v.Cnt > 0 {
					if v.Day >= MinDay {
						if v.Day-v.Cnt >= MinDay {
							PlusFDay += v.Cnt
							if PlusFlag == 0 {
								PlusFlag = 2
							}
						} else {
							PlusFDay += v.Day - MinDay
							PlusFlag = 3
						}
					}
				}
			}

		}
	}

	s.FDayHistoryCB.Print(debug, w)

	fmtPrintln(debug, w, "Добавленные дни просрока для ЦБ", PlusFDay, "Флаг", PlusFlag)
	tD := Date("2021-01-01")

	for _, v := range cps {
		if v.Date.After(tD) {
			if v.IsAdag {
				if v.Date.After(s.XXXDate) {
					fmtPrintln(debug, w, "Ошибка! дата нач. пролонгации после иксов:", s.XXXDate.Format(ISO), "пролонгация от", v.Date.Format(ISO), "на", v.IsAdagDays, "дня", "до", AddDays(v.Date, v.IsAdagDays).Format(ISO))
					break
				} else if AddDays(v.Date, v.IsAdagDays).After(s.XXXDate) {
					fmtPrintln(debug, w, "Ошибка! окончание пролонгации после иксов:", s.XXXDate.Format(ISO), "пролонгация от", v.Date.Format(ISO), "на", v.IsAdagDays, "дня", "до", AddDays(v.Date, v.IsAdagDays).Format(ISO))
					break
				}
			}
		}
	}

	cps.PrintEnd(debug, w, oInfo.Date, oInfo.XXLimitSumm)

	var ret map[string]interface{}

	XXXdatePrn := s.XXXDate.Format(ISO)
	if len(XXXdatePrn) > 10 {
		s.XXXDate = Date("3030-01-01")
	}

	if fullJson {
		ret = map[string]interface{}{"OInfo": oInfo, "BeginSchedule": oS.Schedule, "OrderPaymentSchedule": s, "OrderCheckpoints": cps}
	}

	return s.Schedule[0].BalanceBefore, s, ret, s.ErrorsPresent
}
