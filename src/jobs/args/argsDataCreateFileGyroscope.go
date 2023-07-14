/*
======================
ログファイルの作成について
========================
*/
package args

import (
	"strings"
	"time"

	COMMON "bwing.app/src/common"
	CONFIG "bwing.app/src/config"
	BUCKET "bwing.app/src/gcs/bucket"
)

type ArgsCreateFileGyroscope struct{}

///////////////////////////////////////////////////
/* ===========================================
コマンドライン引数で指定がなかった項目に、contentに応じたDefault値を格納
=========================================== */
func (a *ArgsCreateFileGyroscope) AddDefault2Args(dvm *DVM) {

	/*-----------------------------------
	Content別にデフォルト値を準備
	-----------------------------------*/

	/////////////////////////
	//デフォルト値を設定(content別)

	//frequency
	dvm.DefaultValueMap[CONFIG.ARGS_FREQUENCY] = CONFIG.DEFAULT_STRING_VALUE_WEEKLY

	//term
	dvm.DefaultValueMap[CONFIG.ARGS_TERM] = CONFIG.DEFAULT_STRING_VALUE_15_MIN

	//bucket
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_NAME] = BUCKET.GCS_BUCKET_NAME_GYROSCOPE
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_MIDDLE_PATH] = BUCKET.GCS_MIDDLE_PATH_GYROSCOPE

	//開始日、終了日を設定
	//バッチではDilayで回して、実行日と次の日の期間でファイルを作成する
	//作成済みのファイルは除外するため、実行日で不足分を含め、未来1日分のファイルを先に生成していく
	startDate := strings.Split(time.Now().String(), " ")[0] //実行日
	startAndEndDate := COMMON.GetNextFrequencyDates(CONFIG.DEFAULT_STRING_VALUE_DAILY, startDate, "-")
	dvm.DefaultValueMap[CONFIG.ARGS_EXTRACT_START_DATE] = startAndEndDate[0]
	dvm.DefaultValueMap[CONFIG.ARGS_EXTRACT_END_DATE] = startAndEndDate[1]

}
