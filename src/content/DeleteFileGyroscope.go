/*
======================
Delete File 処理を実行する
========================
*/
package content

import (
	"time"

	CONFIG "bwing.app/src/config"
	FILE "bwing.app/src/file"
)

type DeleteFileGyroscope struct{}

///////////////////////////////////////////////////
/* ===========================================
イベントログファイルをリミットに応じてファイルを分割し、GCSへアップロード
=========================================== */
func (u DeleteFileGyroscope) DeleteFile() ([]string, error) {

	/*-----------------------------
	準備(日時、ファイル名)
	----------------------------- */

	//現在日時を設定
	nt := time.Now().UTC()
	dt := nt.Format(CONFIG.LOG_FILE_DATETIME_LAYOUT)

	//日付をセット
	f := &FILE.EventsLog{Sdate: dt}

	//ファイル名を確認し出力(書き込み用のログファイル)
	fileNames, err := FILE.GetEventsLogFileNames(CONFIG.LOG_UPLOADED_DIR_PATH)
	if err != nil {
		return nil, err
	}

	//ファイル名を選別:現在時(Hour)より古いものに限定
	fileNames = f.GetFileNameOldies(fileNames, 3*24)

	//Gard
	if len(fileNames) == 0 {
		fileNames = append(fileNames, "not found Delete file.")
		return fileNames, nil
	}

	/*-----------------------------
	DELETE
	----------------------------- */

	//Filesの準備(ファイル格納ディレクトリ)
	f.FileDir = CONFIG.GetConfig(CONFIG.LOG_DIR_PATH)

	//ファイル名をセット
	f.FileNames = fileNames

	//ファイルを削除
	err = f.DeleteEventsLog(CONFIG.LOG_UPLOADED_DIR_PATH)
	if err != nil {
		return nil, err
	}

	return f.FileNames, nil
}
