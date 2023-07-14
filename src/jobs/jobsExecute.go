/*
======================
ETL Jobの処理のまとめやく
========================
*/
package jobs

import (
	"fmt"

	CONFIG "bwing.app/src/config"
	CONTENT "bwing.app/src/content"
	ARGS "bwing.app/src/jobs/args"
	LOG "bwing.app/src/log"
	"github.com/pkg/errors"
)

///////////////////////////////////////////////////
/* ===========================================
//Contentで指定されたログのETLを行う
=========================================== */
func ExecuteJobs(content string) ([]string, error) {

	//コマンドライン引数で指定がなかった項目に、contentに応じたDefault値を格納
	err := ARGS.AddDefault2Args(content)
	if err != nil {
		return nil, err
	}

	//コマンドライン引数のバリデーション
	err = ARGS.ValidationReqArgs(content)
	if err != nil {
		return nil, err
	}

	//コマンドライン引数の値を出力
	args := CONFIG.GetConfigArgsAllString()
	fmt.Println(LOG.SetLogEntry(LOG.INFO, "Args", fmt.Sprintf("%+v", args)))

	//コンテントで実行内容を変更
	switch content {
	/*=============================================================
	Create File Gyroscope
	*/
	case CONFIG.CONTENT_NAME_CREATE_FILE_GYROSCOPE:

		//Termに従って、先に書き込み用のログファイルを生成しておく
		var c CONTENT.CreateFileGyroscope
		ufs, err := c.CreateEventsLogFilesByTerm()
		if err != nil {
			return nil, err
		}

		return ufs, nil

	/*=============================================================
	Divide File Gyroscope
	*/
	case CONFIG.CONTENT_NAME_DIVIDE_FILE_GYROSCOPE:

		//ファイルリミットに応じてをファイル分割して、GSCへアップロード
		var c CONTENT.DivideFileGyroscope
		ufs, err := c.DivideFile()
		if err != nil {
			return nil, err
		}

		return ufs, nil

	/*=============================================================
	Upload File Gyroscope
	*/
	case CONFIG.CONTENT_NAME_UPLOAD_FILE_GYROSCOPE:

		//Termに従って、先に書き込み用のログファイルを生成しておく
		var c CONTENT.UploadFileGyroscope
		ufs, err := c.UploadFile()
		if err != nil {
			return nil, err
		}

		return ufs, nil

		/*=============================================================
		Delete File Gyroscope
		*/
	case CONFIG.CONTENT_NAME_DELETE_FILE_GYROSCOPE:

		//Termに従って、先に書き込み用のログファイルを生成しておく
		var c CONTENT.DeleteFileGyroscope
		dfs, err := c.DeleteFile()
		if err != nil {
			return nil, err
		}

		return dfs, nil

	/*=============================================================
	例外処理 */
	default:
		err := errors.New("【Error】content is nothing.")
		if err != nil {
			return nil, err
		}
	}

	//形骸リターン処理
	return nil, nil
}
