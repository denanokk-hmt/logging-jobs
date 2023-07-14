/*
======================
共通処理のファイル
========================
*/
package common

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	CONFIG "bwing.app/src/config"
)

///////////////////////////////////////////////////
/* ===========================================
日付差分
ARGS
	startDate: 開始日
	endDate: 終了日
	delimiter: 日付のデリミタ
RETURN
	日数
=========================================== */
func DateDiffCalculator(startDate, endDate, delimiter string) int {

	reg := regexp.MustCompile(`[-|/|.]`)

	startDate = reg.ReplaceAllString(startDate, delimiter)
	sArr := strings.Split(startDate, delimiter)
	sy, _ := strconv.Atoi(sArr[0])
	sm, _ := strconv.Atoi(sArr[1])
	sd, _ := strconv.Atoi(sArr[2])

	endDate = reg.ReplaceAllString(endDate, delimiter)
	eArr := strings.Split(endDate, delimiter)
	ey, _ := strconv.Atoi(eArr[0])
	em, _ := strconv.Atoi(eArr[1])
	ed, _ := strconv.Atoi(eArr[2])

	start := time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, time.Local)
	end := time.Date(ey, time.Month(em), ed, 0, 0, 0, 0, time.Local)

	//Diff
	diffDays := end.Sub(start).Hours() / 24

	return int(diffDays)
}

///////////////////////////////////////////////////
/* ===========================================
文字列の日付をUTCのTime型に変換
ARGS
	sDateTime: 日時文字列 Foamat:"2023-02-24T11:42:04"
	dtDelimiter: 日付と時間のデリミタ
	dDelimiter: 日付のデリミタ
	tDelimiter: 時間のデリミタ
RETURN
	Time型
=========================================== */
func ConvertStringDateToUtcTime(sDateTime, dtDelimiter, dDelimiter, tDelimiter string, utc bool) time.Time {

	//日付部分を取得
	dtArray := strings.Split(sDateTime, dtDelimiter)

	//日付を分解
	dt := strings.Split(dtArray[0], dDelimiter)
	dy, _ := strconv.Atoi(dt[0])
	dm, _ := strconv.Atoi(dt[1])
	dd, _ := strconv.Atoi(dt[2])

	//時間を分解
	ts := strings.Split(dtArray[1], tDelimiter)
	tHour, _ := strconv.Atoi(ts[0])
	tMin, _ := strconv.Atoi(ts[1])
	tSec, _ := strconv.Atoi(ts[2])
	tMsec := 0

	//Locationを考慮して、時間型へ変換
	var tm time.Time
	if utc {
		tm = time.Date(dy, time.Month(dm), dd, tHour, tMin, tSec, tMsec, time.UTC)
	} else {
		tm = time.Date(dy, time.Month(dm), dd, tHour, tMin, tSec, tMsec, time.Local)
	}

	return tm
}

///////////////////////////////////////////////////
/* ===========================================
日付追加
ARGS
	startDate: 開始日付
	delimiter: 日付のデリミタ
	addD: 追加日数
RETURN
	開始日付から日数分の日付を配列で戻す
=========================================== */
func DateAddCalculator(srcDate, delimiter string, addD int) []string {

	reg := regexp.MustCompile(`[-|/|.]`)

	srcDate = reg.ReplaceAllString(srcDate, delimiter)
	sArr := strings.Split(srcDate, delimiter)
	sy, _ := strconv.Atoi(sArr[0])
	sm, _ := strconv.Atoi(sArr[1])
	sd, _ := strconv.Atoi(sArr[2])

	var dArr []string
	for i := 0; i <= addD; i++ {
		t := time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, time.Local)
		t = t.AddDate(0, 0, i)
		ta := strings.Split(t.String(), " ")[0]
		ts := reg.ReplaceAllString(ta, delimiter)
		dArr = append(dArr, ts)
	}

	return dArr
}

///////////////////////////////////////////////////
/* ===========================================
時間追加
ARGS
	dateTime: 日時
	delimiter: 日付のデリミタ
	日付と時間の区切り文字
	addH: 追加時間
		exp:("2023-01-01T00:00:00", "T", "-", 1, "2006-01-02T15:04:05"
RETURN
	String::2006-01-02 15:04:05
=========================================== */
func HourAddCalculator(dateTime, dtSepalater string, dtDelimiter string, addH int, layout string) string {

	//Dateを導く
	dt := strings.Split(dateTime, dtSepalater)
	dArr := strings.Split(dt[0], dtDelimiter)
	dy, _ := strconv.Atoi(dArr[0])
	dm, _ := strconv.Atoi(dArr[1])
	dd, _ := strconv.Atoi(dArr[2])

	//Timeを導く
	t := dt[1]
	tArr := strings.Split(t, ":")
	th, _ := strconv.Atoi(tArr[0])
	tm, _ := strconv.Atoi(tArr[1])
	ts, _ := strconv.Atoi(tArr[2])
	var tms int = 0
	if len(tArr) > 3 {
		tms, _ = strconv.Atoi(dArr[3])
	}

	if layout == "" {
		layout = CONFIG.LOG_FILE_DATETIME_LAYOUT
	}

	//日時型に変換し時間を加減算した後、フォーマットし文字列で返却
	ta := time.Date(dy, time.Month(dm), dd, th, tm, ts, tms, time.Local)
	ta = ta.Add(time.Duration(addH) * time.Hour)
	rt := ta.Format(layout)
	return rt
}

///////////////////////////////////////////////////
/* ===========================================
次の日付を取得
ARGS
	frequency	[string]:
		"daily"-->次の日,
		"weekly"-->翌週日曜日の日付,
		"monthly"-->翌月初日,
	startDate	[string]:	起点とする日付文字列
	delimiter	[string]:	日付のデリミタ
	startOrEnd [string]: "s" or "e"
RETURN
	日付	[string]
=========================================== */
func GetNextFrequencyDates(frequency, startDate, delimiter string) []string {
	nextSE := make([]string, 2)
	switch frequency {
	case CONFIG.DEFAULT_STRING_VALUE_DAILY:
		nextSE[0] = startDate                                     //実行日
		nextSE[1] = DateAddCalculator(startDate, delimiter, 1)[1] //起点日に1日追加
	case CONFIG.DEFAULT_STRING_VALUE_WEEKLY:
		//起点日の曜日を調べて、次の日曜日を検索(日付追加し、最後の配列の要素)
		sArr := strings.Split(startDate, delimiter)
		sy, _ := strconv.Atoi(sArr[0])
		sm, _ := strconv.Atoi(sArr[1])
		sd, _ := strconv.Atoi(sArr[2])
		t := time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, time.Local)
		nds := DateAddCalculator(startDate, delimiter, 7-int(t.Weekday())) //次の日曜日までの日付
		nextSE[0] = nds[len(nds)-1]
		nde := DateAddCalculator(nextSE[0], delimiter, 6) //次の土曜日までの日付
		nextSE[1] = nde[len(nde)-1]
	case CONFIG.DEFAULT_STRING_VALUE_MONTHLY:
		sArr := strings.Split(startDate, delimiter)
		sy, _ := strconv.Atoi(sArr[0])
		sm, _ := strconv.Atoi(sArr[1])
		var t time.Time
		t = time.Date(sy, time.Month(sm+1), 1, 0, 0, 0, 0, time.Local) //翌月の1日
		nextSE[0] = strings.Split(t.String(), " ")[0]
		t = time.Date(sy, time.Month(sm+2), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1) //翌々月の初日から1dを引く
		nextSE[1] = strings.Split(t.String(), " ")[0]
	}
	return nextSE
}
