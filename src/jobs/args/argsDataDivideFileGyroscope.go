/*
======================
ログファイル分割について
========================
*/
package args

import (
	"time"

	CONFIG "bwing.app/src/config"
	BUCKET "bwing.app/src/gcs/bucket"
)

type ArgsDivideFileGyroscope struct{}

///////////////////////////////////////////////////
/* ===========================================
コマンドライン引数で指定がなかった項目に、contentに応じたDefault値を格納
=========================================== */
func (a *ArgsDivideFileGyroscope) AddDefault2Args(dvm *DVM) {

	/*-----------------------------------
	Content別にデフォルト値を準備
	-----------------------------------*/

	/////////////////////////
	//デフォルト値を設定(content別)

	//frequency
	dvm.DefaultValueMap[CONFIG.ARGS_FREQUENCY] = "hourly"

	//bucket
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_NAME] = BUCKET.GCS_BUCKET_NAME_GYROSCOPE
	dvm.DefaultValueMap[CONFIG.ARGS_BUCKET_MIDDLE_PATH] = BUCKET.GCS_MIDDLE_PATH_GYROSCOPE

	//開始日、終了日に実行日
	timeNow := time.Now().UTC()
	dayNow := timeNow.Format("2006/01/02")
	dvm.DefaultValueMap[CONFIG.ARGS_EXTRACT_START_DATE] = dayNow //extract_start_date
	dvm.DefaultValueMap[CONFIG.ARGS_EXTRACT_END_DATE] = dayNow   //extract_end_date

}
