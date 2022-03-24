package rd

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/table"
)

// OrderInfo Параматры займа
type OrderInfo struct {
	ID                    int
	Num                   string
	Date                  time.Time
	DateText              string
	Days                  int
	Period                int
	PeriodsCount          int
	BasePayNeed           bool
	XXLimit               float64
	XXLimitSumm           float64
	XXAgainFlag           bool
	LgotPeriod            int
	LgotPeriodEnd         bool
	LgotPeriodPayFirstDay int
	Sum                   float64
	PercentRate           float64
	PercentFine1          float64
	PercentFine2          float64
	BackSum               float64
	Psk                   float64
	IsFineBegin           bool
	Annuity               bool
	QuasyAnnuity          bool
	BillingNumber         int
	IsNewOrder            bool
	IsOldOrder2016        bool
	IsOldOrder2017        bool
	IsAnnuity2018         bool
	Pdn                   float64
	Cps                   OrderCheckpoints
	CalcCorrectionFlag    int
	UUID                  string
	AdagCount             int
}

// NewOrderInfo создать структуру OrderInfo
func NewOrderInfo() OrderInfo {
	return OrderInfo{}
}

// GetOrder получить договор из БД
func (oInfo *OrderInfo) GetOrder(db *sql.DB, oNum string) error {
	trx, _ := db.Begin()
	defer trx.Rollback()

	q := sqlSetOrderNum + "'" + oNum + "'"
	rows, err := trx.Query(q)
	if err == nil {
		q = sqlGetOrderID
		rows, err = trx.Query(q)
		if err == nil {
			q = sqlOrderInfo
			rows, err = trx.Query(q)
			if err == nil {
				if rows.Next() {
					var tmpbnp int
					var tmpDate string
					var tmplpend int
					rows.Scan(&oInfo.ID, &oInfo.Num, &tmpDate, &oInfo.Days, &oInfo.Period, &oInfo.PeriodsCount, &tmpbnp,
						&oInfo.XXLimit, &oInfo.LgotPeriod, &oInfo.Sum, &oInfo.PercentRate, &oInfo.PercentFine1, &oInfo.PercentFine2,
						&oInfo.BackSum, &oInfo.Psk, &tmplpend, &oInfo.BillingNumber, &oInfo.Pdn, &oInfo.CalcCorrectionFlag, &oInfo.UUID, &oInfo.AdagCount)
					oInfo.Date = Date(tmpDate)
					oInfo.BasePayNeed = !I2B(tmpbnp)
					oInfo.QuasyAnnuity = I2B(tmpbnp)
					oInfo.LgotPeriodEnd = I2B(tmplpend)


                    // .................
                    // .................
                    // .................
                    // .................
                    // .................
                    // .................
                    

					rows.Close()
				} else {
					err = errors.New("Нет такого займа")
				}
			}
		}
	}
	return err
}

// OrderCheckpoint структура описывающая состояние займа за период
type OrderCheckpoint struct {
	// Дата
	Date time.Time

	// Оплаты, Заморозки, Простипени, Кредитные каникулы, Пролонгации
	IsPay, IsHold, IsPP, IsKK, IsAdag, IsStopPercent2016, IsSudCorr bool

	IsAdagDays int

	PaySum                                        float64
	PayBaseRd, PayPerscentRd, PayPeniRd, PayLawRd float64
	IsKKPP                                        float64
	IsHoldSum                                     OrderMoneyKit
	LawSum                                        float64

	// Новые поля рассчитанные алгоритмом
	BaseAll                                                    float64
	CalcProfit                                                 OrderMoneyKit
	CalcProfitDays                                             OrderDaysKit
	PayBaseGo, PayPerscentGo, PayPeni1Go, PayPeni2Go, PayLawGo float64
	PercentRate                                                float64
	PercentFine1                                               float64
	PercentFine2                                               float64
	IsXXX                                                      bool
	IsScheduleRecalc                                           bool
	BeginSchedule                                              OrderPaymentScheduleArray
	EndSchedule                                                OrderPaymentScheduleArray
	ActiveClientSchedule                                       OrderPaymentScheduleArray
	ActiveClientScheduleDate                                   time.Time

	IsRealKK     bool
	IsRealKKPP   float64
	IsRealKKDays int
}

// NewOrderCheckpoint создать структуру OrderInfo
func NewOrderCheckpoint(d time.Time) OrderCheckpoint {
	tmpCp := OrderCheckpoint{}
	tmpCp.Date = d
	return tmpCp
}

// OrderCheckpoints массив структур описывающих состояние займа за периоды. Возможна сортировка по дате
type OrderCheckpoints []OrderCheckpoint

// Len Forward request for length
func (p OrderCheckpoints) Len() int {
	return len(p)
}

// Less Define compare
func (p OrderCheckpoints) Less(i, j int) bool {
	return p[i].Date.Before(p[j].Date)
}

// Swap Define swap over an array
func (p OrderCheckpoints) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// GetOrderCheckpoints получить все контрольные точки договора
func (oInfo *OrderInfo) GetOrderCheckpoints(db *sql.DB, qDate time.Time, shadowPayOn bool, everyMonthOn bool) (OrderCheckpoints, error) {
	var err error
	var orderC OrderCheckpoints

	orderC = make(OrderCheckpoints, 0)

	if oInfo.ID != 0 {
		trx, _ := db.Begin()
		defer trx.Rollback()

		q := sqlSetQueryDate + qDate.Format(ISO) + "'"
		rows, err := trx.Query(q)
		if err == nil {
			tmpShadowPayOnStr := "0"
			if shadowPayOn {
				tmpShadowPayOnStr = "1"
			}
			q := sqlSetShadowPayOn + tmpShadowPayOnStr + ""
			rows, err = trx.Query(q)
			if err == nil {
				q := sqlSetOrderID + strconv.Itoa(oInfo.ID)
				rows, err = trx.Query(q)
				if err == nil {
					if everyMonthOn {
						q = sqlSetMonthDateAsOrderDate
					} else {
						q = sqlSetMonthDateAsFakeDate
					}
					rows, err = trx.Query(q)
					if err == nil {
						q = sqlCheckPoints
						rows, err = trx.Query(q)
						if err == nil {

							var (
								date                                          string
								isPay, isHold, isPP, isKK                     int
								isAdagDays, isStopPercent2016                 int
								paySum                                        float64
								payBaseRd, payPerscentRd, payPeniRd, payLawRd float64
								isAdag                                        bool
								isKKPP                                        float64
								isHoldSum                                     OrderMoneyKit
								lawSum                                        float64
								sudCorr                                       int
								isRealKK                                      int
								isRealKKPP                                    float64
								isrealKKDays                                  int
							)

							for rows.Next() {
								rows.Scan(&date, &isPay, &isHold, &isHoldSum.Base, &isHoldSum.Perc, &isHoldSum.FOne, &isHoldSum.FTwo, &isHoldSum.Duty,
									&isPP, &isKK, &isKKPP, &isRealKK, &isRealKKPP, &isrealKKDays, &isAdagDays, &isStopPercent2016, &lawSum, &sudCorr,
									&paySum, &payBaseRd, &payPerscentRd,
									&payPeniRd, &payLawRd)
                                
                                // .................
                                // .................
                                // .................
                                // .................
                                // .................
                                // .................
                                // .................
                                
							}
							rows.Close()
						}
					}
				}
			}
		}
	}
	return orderC, err
}

// Print Print checkpoint table
func (p OrderCheckpoints) Print(debug bool, w http.ResponseWriter, begin time.Time) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"День", "Дата", "Платеж", "Заморозка", "Простипени", "Каникулы", "Пролонгация", "ГП", "Сумма", "Тело", "Процент", "Пени", "Пошлины"})

	totalSum := 0.0
	totalBase := 0.0
	totalPerscent := 0.0
	totalPeni := 0.0
	totalLaw := 0.0
	if debug {
		for _, cp := range p {
			dd := DiffDays(begin, cp.Date)
			begin = cp.Date
			tmpKK := "0"
			if cp.IsKK {
				tmpKK = fmt.Sprintf("%f", cp.IsKKPP)
			}

			tmpHold := B2I(cp.IsHold)
			if cp.IsStopPercent2016 {
				tmpHold = 2
			}

			t.AppendRow([]interface{}{dd, cp.Date.Format(ISO), B2I(cp.IsPay), tmpHold, B2I(cp.IsPP), tmpKK, B2I(cp.IsAdag), cp.LawSum, cp.PaySum, cp.PayBaseRd, cp.PayPerscentRd, cp.PayPeniRd, cp.PayLawRd})
			//fmt.Println(id, name, surname)
			totalSum += cp.PaySum
			totalBase += cp.PayBaseRd
			totalPerscent += cp.PayPerscentRd
			totalPeni += cp.PayPeniRd
			totalLaw += cp.PayLawRd
		}
		t.AppendFooter(table.Row{"", "", "", "", "", "", "", "Итого:", Round(totalSum, 2), Round(totalBase, 2), Round(totalPerscent, 2), Round(totalPeni, 2), Round(totalLaw, 2)})

		if w != nil {
			t.SetOutputMirror(w)
		}

		t.Render()
	}
}

// BasePay расчета Аннуитетного платежа. Работает и для дифференцированного и для квази-аннуитета.
// Параметры: Сумма займа, процентная ставка, дни на сколько выдан, дней в периоде платежа
func BasePay(summa float64, percent float64, days int, period int, baseNeedPayed bool) float64 {
	var tmpPaySum float64
	if period == 0 {
		period = days
	}
	if period != days && days > period {
		if baseNeedPayed {
			periodsCnt := days / period                             // Количество целых периодов
			periodsOst := days % period                             // Количество дней в последнем периоде
			tmpSum := summa                                         // Начальная сумма для расчета графика (чтоб цикл сработал)
			tmpPaySumLo := 0.0                                      // Нижний  порог суммы для метода половинного деления
			tmpPaySumHi := tmpSum + summa*percent/100*float64(days) // Верхний порог суммы для метода половинного деления
			for tmpSum < 0 || tmpSum > 0.001 {                      // Цикл расчета суммы аннуитетного платежа с точностью 0.001
				tmpSum = summa                              // Начальная сумма для расчета графика равна сумме займа
				tmpPaySum = (tmpPaySumLo + tmpPaySumHi) / 2 // Методом половинного деления берем среднее значение суммы платежа в период
				for i := 0; i < periodsCnt; i++ {           // Цикл по целым периодам. Вычисляем по формуле СуммаОстаток+Проценты-СуммаПлатежа
					tmpSum = tmpSum + tmpSum*percent/100*float64(period) - tmpPaySum // Из того что получилось вычитаем СуммуПлатежа за неполный период
				}
				tmpSum = tmpSum + tmpSum*percent/100*float64(periodsOst) - (tmpPaySum / float64(period) * float64(periodsOst))
				if tmpSum < 0 { // Если разница меньше нуля значит СуммаПлатежа больше чем нужно. Сдвигаем верхний порог
					tmpPaySumHi = tmpPaySum
				} else { // Если разница больше нуля значит СуммаПлатежа меньше чем нужно. Сдвигаем нижний порог
					tmpPaySumLo = tmpPaySum
				}
			}
		} else {
			tmpPaySum = summa * percent / 100 * float64(period)
		}
	} else {
		tmpPaySum = summa + summa*percent/100*float64(days)
	}
	return tmpPaySum
}

// CalcPSK Функция: расчета ПСК.
// Параметры: Сумма займа, процентная ставка, дни на сколько выдан, дней в периоде платежа
func CalcPSK(summa float64, percent float64, days int, period int, baseNeedPayed bool, oInfo OrderInfo) float64 {
	type pskScheduleElement struct {
		qk float64
		ek float64
		ps float64
	}

	Annuity := false
	if period > 0 && baseNeedPayed {
		Annuity = true
	}

	if period == 0 {
		period = days
	}
	if !baseNeedPayed { // Вот как правильно считать для квазианнуитетов
		period = days
	}

	tmpI := 0.0

	// Формула аннуитета работает и для деффиренцированного займа и для квазианнуитета. Условие сдалано для оптимизации
	if Annuity {
		schedule, _ := GetScheduleArray(summa, percent, days, period, baseNeedPayed, oInfo)
		pskSchedule := make([]pskScheduleElement, len(schedule))

		tmpPeriod := 0
		for k := range schedule {
			tmpPeriod += schedule[k].Period
			pskSchedule[k].qk = float64(tmpPeriod / period)
			pskSchedule[k].ek = float64((tmpPeriod % period) / period)
			pskSchedule[k].ps = Round(schedule[k].Schedule.Perc+schedule[k].Schedule.Base, 2)
		}
		tmpIlo := 0.0 // Нижний  порог для метода половинного деления
		tmpIhi := 1.0 // Верхний порог для метода половинного деления
		tmpSum := -summa
		tmpCnt := 1
		tmpI = 0.0
		for tmpSum < 0 || tmpSum > 0.001 {
			tmpSum = -summa
			tmpI = (tmpIlo + tmpIhi) / 2
			for k := range schedule {
				tmpSum += pskSchedule[k].ps / ((1.0 + tmpI*pskSchedule[k].ek) * math.Pow(1.0+tmpI, pskSchedule[k].qk))
			}
			if tmpSum < 0 {
				tmpIhi = tmpI
			} else {
				tmpIlo = tmpI
			}
			tmpCnt++
		}
	} else {
		tmpI = (summa+summa*percent/100*float64(days))/summa - 1
		period = days
	}
	return Round(tmpI*365/float64(period)*100, 3)
}

// GetScheduleArray возвращает график платежей
func GetScheduleArray(sumMa float64, percent float64, days int, period int, baseNeedPayed bool, oInfo OrderInfo) (OrderPaymentScheduleArray, float64) {
	var (
		i       int
		paySum  float64
		tmpPrc  float64
		workSum float64
	)

	if days <= 0 {
		days = 1
	}

	if period == 0 {
		period = days
	}

	workSum = sumMa
	paySum = BasePay(workSum, percent, days, period, baseNeedPayed)
	if baseNeedPayed {
		// до 2017-01-23 сумма платежа не округлялась (была с копейками).
		if oInfo.BillingNumber == 3 || oInfo.Date.After(Date("2017-01-23")) {
			paySum = math.Ceil(paySum)
		} else {
			paySum = Round(paySum, 2)
		}
	} else {
		// У квазианнуитетов сумма платежа не округляется, остается с копейками.
		paySum = Round(paySum, 2)
	}

	periodsCnt := days / period // Количество целых периодов
	periodsOst := days % period // Количество дней в последнем периоде

	if periodsOst == 0 {
		periodsOst = period
		periodsCnt--
	}

	Schedule := make(OrderPaymentScheduleArray, periodsCnt+1)
	for i = 0; i < periodsCnt; i++ {
		tmpPrc = Round(workSum*percent/100*float64(period), 2)
		Schedule[i].Fill(period, workSum, Round(paySum-tmpPrc, 2), tmpPrc)
		Schedule[i].FineOn = true
		Schedule[i].Index = i + 1
		workSum = Round(workSum+tmpPrc-paySum, 2)
	}

	tmpPrc = Round(workSum*percent/100*float64(periodsOst), 2)
	paySum = Round(workSum+tmpPrc, 2)
	Schedule[i].Fill(periodsOst, workSum, Round(paySum-tmpPrc, 2), tmpPrc)
	Schedule[i].FineOn = true
	Schedule[i].Index = i + 1
	workSum = Round(workSum+tmpPrc-paySum, 2)

	Schedule[0].FineOn = false

	return Schedule, workSum
}
