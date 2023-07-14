/*
======================
イベントログの書き込み処理
========================
*/
package file

import (
	"fmt"
	"strconv"
	"strings"

	COMMON "bwing.app/src/common"
	CONFIG "bwing.app/src/config"
	GCS "bwing.app/src/gcs"
	LOG "bwing.app/src/log"
)

type EventsLog struct {
	Sdate       string   //日時の文字列(書き込み:タイムスタンプ、バッチなどの毎時のUpload:実行日時)Foamat:"2023-02-24T11:42:04"
	EventsLog   string   //書き込むイベントログ文字列
	ForceUpload bool     //強制的にアップロードフラグ
	FileDir     string   //ファイル格納ディレクトリ
	FileNames   []string //アップロードを行うファイル名たち
}

///////////////////////////////////////////////////
/*===========================================
現在より古いファイル名だけを選出
fns [文字列Slice]：ログファイル名達
dH [int]：Hour 例:1を指定した場合、1時間以上古いログなら採用
===========================================*/
func (e *EventsLog) GetFileNameOldies(fns []string, dH int) []string {

	//アップロードするファイル名を格納
	var uploadFiles []string
	if e.ForceUpload {
		uploadFiles = fns //強制的にファイルを指定
	} else {
		//実行日時と比較して、日付やHourが古い場合のファイルを検索
		//実行日時を分解して、対象ファイルを導く
		//Foamat:"2023-02-24T11:42:04"
		sTs := COMMON.ConvertStringDateToUtcTime(e.Sdate, "T", "-", ":", true)

		//アップロードするファイル名を箱に詰める
		for _, fn := range fns {

			//Foamat:"2023-01-01T00:00:00_00:14:59.txt"
			ts := strings.Split(fn, "_")[0]                                                   //日付部分と前半時間部分
			tm := COMMON.HourAddCalculator(ts, "T", "-", dH, CONFIG.LOG_FILE_DATETIME_LAYOUT) //1Hourを足す
			fTs := COMMON.ConvertStringDateToUtcTime(tm, "T", "-", ":", true)

			//ファイル日時(+1hour)と実行日時を比較して、ファイル日時が古い場合、採用
			if fTs.Before(sTs) {
				uploadFiles = append(uploadFiles, fn)
			}
		}
	}

	return uploadFiles
}

///////////////////////////////////////////////////
/*===========================================
ログファイルを分割
===========================================*/
func (e *EventsLog) DivideFiles() error {

	var err error
	var uploadFiles []string

	//ファイルリミットに応じて分割
	if CONFIG.EVENTS_LOG_FILE_DIVIDE {
		uploadFiles, err = devideEventsLogFileByite(e.FileDir, e.FileNames)
		if err != nil {
			return err
		}
	}

	e.FileNames = uploadFiles

	return nil
}

///////////////////////////////////////////////////
/*===========================================
GSCへログファイルをアップロード
===========================================*/
func (e EventsLog) UploadEventsLog2Gcs() error {

	/*-----------------------------
	GCSバケット準備（バケット名、パス）
	----------------------------- */
	//Set bucket name
	bucketName := CONFIG.GetConfigArgs(CONFIG.ARGS_BUCKET_PREFIX) + CONFIG.GetConfigArgs(CONFIG.ARGS_BUCKET_NAME) + CONFIG.GetConfigArgs(CONFIG.ARGS_BUCKET_SUFFIX) //complete bucket name

	//bucket middle path
	bucketPath := CONFIG.GetConfigArgs(CONFIG.ARGS_BUCKET_MIDDLE_PATH)

	/*-----------------------------
	GCSへアップロードとローテート(移動)
	----------------------------- */
	for i, uf := range e.FileNames {
		LOG.JustLogging(fmt.Sprintf("GCSへアップロード中:[%s][%d/%d]\n", uf, i+1, len(e.FileNames)))

		//アップロードDirをファイル名から取得(Format:2006-1-12T00:00:00_00:59:59.txt)
		bDir := strings.ReplaceAll(strings.Split(uf, "T")[0], "-", "/")

		//GCSバケット情報をセット
		g := &GCS.GcsUpload{
			BucketName:    bucketName,
			BucketDirPath: bucketPath + "/" + bDir,
			DirPath:       e.FileDir,
			FileName:      uf,
		}

		//GCSバケットアップロード
		err := g.UploadFiletoGcsBucket()
		if err != nil {
			return err
		}
		LOG.JustLogging(fmt.Sprintf("GCSへアップロード完了:[%s][%d/%d]\n", uf, i+1, len(e.FileNames)))

		//アップロードしたファイルを移動
		f := &Files{DirPath: e.FileDir, FileName: uf}
		destinyFilePath := e.FileDir + CONFIG.LOG_UPLOADED_DIR_PATH + uf
		err = f.RenameFile(destinyFilePath)
		if err != nil {
			return err
		}

		LOG.JustLogging(fmt.Sprintf("GCSへアップロード完了:[%s][%d/%d]\n", uf, i+1, len(e.FileNames)))
	}

	return nil
}

///////////////////////////////////////////////////
/*===========================================
event log ファイル名を全部戻す
===========================================*/
func devideEventsLogFileByite(FileDir string, uploadFiles []string) ([]string, error) {

	var ufs []string //分割ファイル名を入れる箱

	//ファイルリミットに応じて分割
	for _, uf := range uploadFiles {

		//ソースFilesの準備(ファイル格納ディレクトリ)
		f := &Files{DirPath: FileDir, FileName: uf}

		//ファイルサイズを取得
		fSize, err := f.GetClosedFileSize()
		if err != nil {
			return nil, err
		}

		//書き込みを行ったファイルサイズが設定しきい値を越えた場合、アップロード対象とする
		if fSize <= CONFIG.EVENTS_LOG_FILE_SIZE_THRESHOLD {
			ufs = uploadFiles
		} else {

			//元ファイルを読み込み
			ss, err := f.ReadAllByScanner()
			if err != nil {
				return nil, err
			}

			//チャンク数を計算
			chunks := COMMON.ChunkCalculator2(len(ss), CONFIG.UPLOAD_FILE_MAX_ROWS)

			//データのチャンク数に応じて分割ファイルを生成する
			for i, c := range chunks.Positions {

				//新しいファイル名を設定(拡張子の前に連番_S[i]を挿入)
				suffix := "_S" + strconv.Itoa(i)
				nf := strings.Split(uf, ".")[0] + suffix + ".txt"

				LOG.JustLogging(fmt.Sprintf("アップロードファイルを分割中:[%s][%d/%d]\n", nf, i+1, len(chunks.Positions)))

				//ファイルを新規作成 & 追記モード
				f := &Files{DirPath: FileDir, FileName: nf}
				newFile, err := f.OpenFileMode_RDWR_CREATE_APPEND_SYNC()
				if err != nil {
					return nil, err
				}
				defer newFile.Close()

				//新規作成ファイルに行単位で書き込む
				for i := c.Start; i < c.End; i++ {
					fmt.Fprintln(newFile, ss[i]) //ログを追記
				}

				//新規作成ファイル名をつめる
				ufs = append(ufs, nf)

				LOG.JustLogging(fmt.Sprintf("ファイル分割完了:[%s][%d/%d]\n", nf, i+1, len(chunks.Positions)))
			}

			//分割した元ファイルを移動
			f := &Files{DirPath: FileDir, FileName: uf}
			destinyFilePath := FileDir + CONFIG.LOG_UPLOADED_DIR_PATH + uf
			err = f.RenameFile(destinyFilePath)
			if err != nil {
				return nil, err
			}
		}
	}

	return ufs, nil
}

///////////////////////////////////////////////////
/*===========================================
event log ファイル名を全部戻す
===========================================*/
func GetEventsLogFileNames(addDir string) ([]string, error) {

	//Filesの準備(ファイル格納ディレクトリ)
	dir := CONFIG.GetConfig(CONFIG.LOG_DIR_PATH)
	if addDir != "" {
		dir += addDir //例uploaded削除バッチ時に指定する
	}
	f := &Files{DirPath: dir}

	//ファイル名を降順で取得
	f.Asc = false
	fileNames, err := f.GetFileNamesAndSortSimple()
	if err != nil {
		return nil, err
	}

	//NFS対応(os.Readの検索で"lost+found"がファイル名に入ってくる)、余計なファイル名を削除
	fileNames = COMMON.RemoveSliceValue(fileNames, "lost+found")

	return fileNames, nil
}

///////////////////////////////////////////////////
/*===========================================
event log を削除
===========================================*/
func (e EventsLog) DeleteEventsLog(addDir string) error {

	//Filesの準備(ファイル格納ディレクトリ)
	dir := CONFIG.GetConfig(CONFIG.LOG_DIR_PATH)
	if addDir != "" {
		dir += addDir //例uploaded削除バッチ時に指定する
	}
	f := &Files{DirPath: dir}

	//すべてのログファイルを削除
	for _, df := range e.FileNames {
		f = &Files{DirPath: dir, FileName: df}
		err := f.DeleteFile()
		if err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////
/*===========================================
Tremに従いevent logファイルを生成
===========================================*/
func (e *EventsLog) CreateEventsLogFilesByTerm() error {

	//空のログファイルを生成する
	var createFiles []string
	for i, fn := range e.FileNames {

		//Filesの準備(ファイル格納ディレクトリ, ファイル名をセット)
		f := &Files{DirPath: CONFIG.GetConfig(CONFIG.LOG_DIR_PATH), FileName: fn}

		//新規作成または、追記モードでファイルをOpen
		file, err := f.OpenFileMode_RDWR_CREATE_APPEND_SYNC()
		LOG.JustLogging(fmt.Sprintf("create log file:[%s][%d/%d]\n", fn, i+1, len(e.FileNames)))
		if err != nil {
			return err
		}
		defer file.Close()
		createFiles = append(createFiles, fn)
	}

	e.FileNames = createFiles

	return nil
}
