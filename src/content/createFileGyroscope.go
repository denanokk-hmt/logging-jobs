/*
======================
Create File 処理を実行する
========================
*/
package content

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	COMMON "bwing.app/src/common"
	CONFIG "bwing.app/src/config"
	FILE "bwing.app/src/file"
)

type CreateFileGyroscope struct{}

///////////////////////////////////////////////////
/* ===========================================
指定期間内でTermに従ってイベントログファイルを生成
=========================================== */
func (d CreateFileGyroscope) CreateEventsLogFilesByTerm() ([]string, error) {

	term, _ := strconv.Atoi(CONFIG.GetConfigArgs(CONFIG.ARGS_TERM))
	startDate := CONFIG.GetConfigArgs(CONFIG.ARGS_EXTRACT_START_DATE)
	endDate := CONFIG.GetConfigArgs(CONFIG.ARGS_EXTRACT_END_DATE)

	//EventsLog files名を生成
	fileNames, err := getTermEventsLogFilesNames(term, startDate, endDate, "-", "txt")
	if err != nil {
		return nil, err
	}

	//イベントログの生成
	f := &FILE.EventsLog{FileNames: fileNames}
	err = f.CreateEventsLogFilesByTerm()
	if err != nil {
		return nil, err
	}

	return f.FileNames, nil
}

///////////////////////////////////////////////////
/*===========================================
指定期間におけるTerm毎ログファイル名を取得
	startDate, endDate: "2023-01-01"
	delimiter: "-"
	extention: "txt"
	term: 15 or 30 or 60
	例：Terｍが15、2023/1/1の00時の場合に欲しい、ファイル名一覧
		2023-01-01T00:00:00_00:14:59.txt
		2023-01-01T00:15:00_00:29:59.txt
		2023-01-01T00:30:00_00:44:59.txt
		2023-01-01T00:45:00_00:59:59.txt
===========================================*/
func getTermEventsLogFilesNames(term int, startDate, endDate, dateDelimiter, extension string) ([]string, error) {

	//不要な文字列を削除する正規表現
	reg := regexp.MustCompile(`[-|/|:| |　|T]`)

	//現在日時を設定
	nt := time.Now().UTC()
	dt := nt.Format(CONFIG.LOG_FILE_DATETIME_LAYOUT)
	dt = reg.ReplaceAllString(dt, "")
	iDt, _ := strconv.Atoi(dt)

	//指定期間の日数
	diff := COMMON.DateDiffCalculator(startDate, endDate, dateDelimiter)

	//日数をもとに期間の年月日を配列で取得
	sDates := COMMON.DateAddCalculator(startDate, "-", diff)

	//Suffix番号を設定
	var suffixNos []string
	for i := 0; i < CONFIG.LOGGING_GY_POD_DEFAULT_QTY; i++ {
		suffixNos = append(suffixNos, strconv.Itoa(i))
	}

	var fileNames []string

	//日付期間中において、termで分割された24時間分のファイル名を配列で取得
	for _, s := range sDates {
		//24時間分回す
		for h := 0; h < 24; h++ {

			//term別の分割回数(=Loop回数)
			var termLoop int
			switch term {
			case 15:
				termLoop = 4
			case 30:
				termLoop = 2
			case 60:
				termLoop = 1
			}

			//ファイル名を生成
			var termBuff int
			var fileName string
			for i := 0; i < termLoop; i++ {
				//ファイル名を整形
				minS := fmt.Sprintf("%02d", termBuff)                 //termバッファ=開始時刻:分
				termBuff = termBuff + term                            //termをバッファ(累積)
				minE := fmt.Sprintf("%02d", termBuff-1)               //termバッファから1を引く=終了時刻:分
				startT := fmt.Sprintf("%02d", h) + ":" + minS + ":00" //開始時刻を形成
				endT := fmt.Sprintf("%02d", h) + ":" + minE + ":59"   //終了時刻を形成
				fileName = s + "T" + startT + "_" + endT              //日付と時刻をつなぎ合わせる

				//現在時刻とファイル名(日付&現在時刻)を照らし合わせて、同じまたは古い場合は対象外
				checkDt := s + "T" + startT //確認用
				checkDt = reg.ReplaceAllString(checkDt, "")
				iCheckDt, _ := strconv.Atoi(checkDt)
				if iCheckDt > iDt {
					if len(suffixNos) == 0 {
						fileNames = append(fileNames, fileName+"."+extension) //ファイル名を格納
					} else {
						//Statefulsetを利用する場合、PodとSuffixが一致するファイルを作成(Pod台数=termログファイル数)
						for _, s := range suffixNos {
							fileNames = append(fileNames, fileName+"_"+s+"."+extension) //ファイル名を格納
						}
					}
				}
			}
		}
	}

	//ファイル名を確認し出力(アップロード前の作成済みのログファイル)
	alreadyFiles, err := FILE.GetEventsLogFileNames("")
	if err != nil {
		return nil, err
	}

	//未作成のファイル名のみにする
	var notYetFiles []string
	for _, fn := range fileNames {
		if !COMMON.StringSliceSearch(alreadyFiles, fn) {
			notYetFiles = append(notYetFiles, fn)
		}
	}

	return notYetFiles, nil
}
