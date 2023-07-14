/*
======================
Cloud-Run Jobsからの
設定値を取得
設定値についての処理
========================
*/
package args

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	VALID "bwing.app/src/common"
	CONFIG "bwing.app/src/config"
)

type DVM struct {
	DefaultValueMap map[string]string
}

///////////////////////////////////////////////////
/* ===========================================
コマンドライン引数で指定がなかった項目に、contentに応じたDefault値を格納
=========================================== */
func AddDefault2Args(content string) error {

	/*-----------------------------------
	Content別にデフォルト値を準備
	-----------------------------------*/

	//デフォルト値の箱を準備(Mapper)
	var dvm DVM
	dvm.DefaultValueMap = make(map[string]string)

	/////////////////////////
	//デフォルト値を設定(content別)
	/////【【【コンテント追加時に必須追加】】】/////

	switch content {

	//divide file gyroscopeの場合
	case CONFIG.CONTENT_NAME_CREATE_FILE_GYROSCOPE:
		var args ArgsCreateFileGyroscope
		args.AddDefault2Args(&dvm)

	//divide file gyroscopeの場合
	case CONFIG.CONTENT_NAME_DIVIDE_FILE_GYROSCOPE:
		var args ArgsDivideFileGyroscope
		args.AddDefault2Args(&dvm)

	//upload file gyroscopeの場合
	case CONFIG.CONTENT_NAME_UPLOAD_FILE_GYROSCOPE:
		var args ArgsUploadFileGyroscope
		args.AddDefault2Args(&dvm)

	//delete file gyroscopeの場合
	case CONFIG.CONTENT_NAME_DELETE_FILE_GYROSCOPE:
		var args ArgsDeleteFileGyroscope
		args.AddDefault2Args(&dvm)

	//エラー
	default:
		return fmt.Errorf("【Error】[func:%s][Args:%s][msg:%s]", "/src/jobs/argsData.go/AddDefault2Args()", content, "コマンドライン引数のキーが一致しません。")
	}

	/////////////////////////
	//デフォルト値を設定(共通項目)
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_PREFIX] = ""
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_SUFFIX] = ""
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_TIME_JST] = "false"

	/*-----------------------------------
	未登録のリクエストデータにデフォルト値または、引数指定値をあてる
	-----------------------------------*/

	var args []string //CONFIG上書き用の箱
	for k, v := range dvm.DefaultValueMap {
		switch k {
		case CONFIG.ARGS_EXTRACT_END_DATE:
			//extract_end_dateのコマンドライン引数の指定があった場合(ローカルデバッグ時を想定)、
			//指定されたextract_end_dateに対して1日加算を行う(条件；bucket_time_jstがtrue以外)
			endD := CONFIG.GetConfigArgs(k)
			if endD != "" {
				jst := CONFIG.GetConfigArgs(CONFIG.ARGS_BUCKET_TIME_JST) //bucket_time_jstを取得
				if jst == "true" {
					//指定引数をtime.Time型に変換
					t := strings.ReplaceAll(endD, "/", "-") + "T00:00:00+00:00"
					parsedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", t)
					//1日加算する
					time1dAfter := parsedTime.AddDate(0, 0, 1)
					//フォーマット
					day1after := time1dAfter.Format("2006/01/02")
					//CONFIG上書き用に確保
					args = append(args, k+":"+day1after)
				}
			} else {
				//extract_end_dateの引数指定がなければ、デフォルトをあてる
				args = append(args, k+":"+v)
			}
		default:
			//引数指定がなければ、デフォルトをあてる
			vv := CONFIG.GetConfigArgs(k)
			if vv == "" {
				args = append(args, k+":"+v)
			}
		}
	}

	/*-----------------------------------
	ConfigのArgsMapperを上書き
	-----------------------------------*/

	//Again Set CMD Args values
	if len(args) != 0 {
		err := CONFIG.SetConfigMapArgs(args)
		if err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////
/* ===========================================
contentに応じたValidation
	frequency::値の有無
	contents::検査しない→switchのdefaultでエラー
	force_index::検査なし
	bucket_name::値の有無
	bucket_middle_path::値の有無
	bucket_prefix::検査なし
	bucket_suffix::検査なし
	bucket_time_jst::検査なし
	extract_start_date::日付の確認、※1
	extract_end_date::日付の確認、※1:extract_start_date <= extract_end_dateの関係性であること
=========================================== */
func ValidationReqArgs(content string) error {

	//抽出日検証用マップの準備
	extractDate := make(map[string]int)

	//引数コマンドラインの値を取得
	configs := CONFIG.GetConfigArgsAll()

	//引数コマンドラインごとに検証を実施
	for key, val := range configs {
		switch key {
		case
			CONFIG.ARGS_FREQUENCY:
			switch val {
			case CONFIG.DEFAULT_STRING_VALUE_HOURLY:
			case CONFIG.DEFAULT_STRING_VALUE_DAILY:
			case CONFIG.DEFAULT_STRING_VALUE_WEEKLY:
			case CONFIG.DEFAULT_STRING_VALUE_MONTHLY:
			default:
				return fmt.Errorf("valid error frequency is not daily or weekly or monthly:[%s]", val)
			}
		case
			CONFIG.ARGS_BUCKET_NAME,
			CONFIG.ARGS_BUCKET_MIDDLE_PATH:
			//値の有無を確認
			if val == "" {
				return fmt.Errorf("【Valid Error】[Args:%s][msg:%s]", key, "値が設定されていません。")
			}
		case
			CONFIG.ARGS_EXTRACT_START_DATE,
			CONFIG.ARGS_EXTRACT_END_DATE:
			//日付確認
			if !VALID.DateFormatChecker(val) {
				return fmt.Errorf("【Valid Error】[Args:%s][val:%s][msg:%s]", key, val, "日付が設定されていません。")
			}
			//日付期間確認
			reg := regexp.MustCompile(`[-|/]`)
			str := reg.ReplaceAllString(val, "")
			num, err := strconv.Atoi(str) //数値に変換
			if err != nil {
				return fmt.Errorf("【Error】[func:%s][Args:%s][msg:%s]", "/src/jobs/argsData.go/ValidationReqArgs()", key, err)
			}
			extractDate[key] = num
			start := extractDate[CONFIG.ARGS_EXTRACT_START_DATE]
			end := extractDate[CONFIG.ARGS_EXTRACT_END_DATE]
			if len(extractDate) == 2 {
				if start > end {
					return fmt.Errorf("【Valid Error】[Args:%s][startDate:%d][startEnd:%d][msg:%s]", key, start, end, "開始日と終了日の関係性が異常です。")
				}
			}
		}
	}

	/////【【【コンテント追加時に必須追加】】】/////
	//content個別バリデーション
	switch content {
	case CONFIG.CONTENT_NAME_CREATE_FILE_GYROSCOPE:
		for key, val := range configs {
			switch key {
			case CONFIG.ARGS_TERM:
				switch val {
				case CONFIG.DEFAULT_STRING_VALUE_15_MIN:
				case CONFIG.DEFAULT_STRING_VALUE_30_MIN:
				case CONFIG.DEFAULT_STRING_VALUE_60_MIN:
				default:
					return fmt.Errorf("valid error term is not 15 or 30 or 60:[%s]", val)
				}
			}
		}
	case
		CONFIG.CONTENT_NAME_DIVIDE_FILE_GYROSCOPE:
	case
		CONFIG.CONTENT_NAME_UPLOAD_FILE_GYROSCOPE:
	case
		CONFIG.CONTENT_NAME_DELETE_FILE_GYROSCOPE:
	default:
		return fmt.Errorf("【Error】[func:%s][Args:%s][msg:%s]", "/src/jobs/argsData.go/ValidationReqArgs()", content, "コマンドライン引数のキーが一致しません。")
	}
	return nil
}
