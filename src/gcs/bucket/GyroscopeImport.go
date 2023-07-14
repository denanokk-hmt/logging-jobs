/*
	=================================

GCS
bucket::flight_record
バケットの構造をここで指定する
バケットファイルをパースする
* =================================
*/
package bucket

import (
	"time"
)

///////////////////////////////////////////////////
/* ===========================================
Table::
* =========================================== */

var (
	GCS_BUCKET_NAME_GYROSCOPE = "gyroscope"
	GCS_MIDDLE_PATH_GYROSCOPE = "events"
)

// ログ格納
type LogGyroscope struct {
	Log  string
	Path string
}

// データロード結果
type LoadGyroscopeImportResults struct {
	Result  int
	Client  string
	Cdt     time.Time
	LogNo   int
	LogPath string
	LogDate string
	TTL     int
}

// データロード情報
type LoadGyroscopeImport struct {
	TimeStamp time.Time
}
