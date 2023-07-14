/*
	=================================

サーバーのConfigを設定する
* =================================
*/
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	LOG "bwing.app/src/log"
)

var (
	SYSTEM_COMPONENT_VERSION = "1.1.0_lj"

	RESULT_SUCCESS = "Success"
	RESULT_FAILURE = "Failure"

	ENV_LOCAL_OR_REMOTE = "LocalOrRemote"
	ENV_GCP_PROJECT_ID  = "GcpProjectId"
	ENV_SERVER_CODE     = "ServerCode"
	ENV_APPLI_NAME      = "AppliName"
	ENV_ENV             = "Env"
	ENV_FORCE_INDEX     = "force_index"

	//デフォルト
	DEFAULT_STRING_VALUE_HOURLY  = "hourly"
	DEFAULT_STRING_VALUE_DAILY   = "daily"
	DEFAULT_STRING_VALUE_MONTHLY = "monthly"
	DEFAULT_STRING_VALUE_WEEKLY  = "weekly"
	DEFAULT_STRING_VALUE_TRUE    = "true"
	DEFAULT_STRING_VALUE_FALSE   = "false"
	DEFAULT_STRING_VALUE_15_MIN  = "15"
	DEFAULT_STRING_VALUE_30_MIN  = "30"
	DEFAULT_STRING_VALUE_60_MIN  = "60"

	//Jobsのコマンドライン引数に利用
	//コマンドラインの引数指定方法は、key:value 形式で行う
	ARGS_CONTENTS           = "contents"           //実行したいcontent名::複数を指定場合、contents:content1,content2,,,のようにvalueをカンマつなぎで指定する
	ARGS_FORCE_INDEX        = "force_index"        //強制的に実行する位置を指定
	ARGS_FREQUENCY          = "frequency"          //daily or monthly
	ARGS_TERM               = "term"               //ログファイルの時間間隔(15, 30, 60)
	ARGS_BUCKET_NAME        = "bucket_name"        //抽出したいログのGCSバケット名を指定
	ARGS_BUCKET_MIDDLE_PATH = "bucket_middle_path" //抽出したいログのGCSバケットのPATHを指定
	ARGS_BUCKET_PREFIX      = "bucket_prefix"      //バケット名にプレフィックスをつけたい場合に指定
	ARGS_BUCKET_SUFFIX      = "bucket_suffix"      //バケット名にサフィックスをつけたい場合に指定
	ARGS_BUCKET_TIME_JST    = "bucket_time_jst"    //バケットを取得する際にJST基準で取得した場合にtrue
	ARGS_EXTRACT_START_DATE = "extract_start_date" //抽出したいログの開始日を指定
	ARGS_EXTRACT_END_DATE   = "extract_end_date"   //抽出したいログの終了日を指定

	//GCS UPLOAD FILE関連
	_, b, _, _                     = runtime.Caller(0)
	root                           = filepath.Join(filepath.Dir(b), "../../")
	LOG_STORAGE_DIR_ABSOLUTE_PATH  = root + "/cmd/logStorage/events/" //Rootディレクトリを取得して、絶対パスを指定
	LOG_STORAGE_VOLUME_DIR_PATH    = "/mnt/gyroscope/events/"         //NFSマウントしたパス
	LOG_FILE_DATETIME_LAYOUT       = "2006-01-02T15:04:05"
	EVENTS_LOG_FILE_DIVIDE         = true //EVENTS_LOG_FILE_SIZE_THRESHOLDを超えるサイズだった場合、UPLOAD_FILE_MAX_ROWSで分割
	EVENTS_LOG_FILE_SIZE_THRESHOLD = int64(102400000)
	LOG_DIR_PATH                   = ""
	UPLOAD_FILE_MAX_ROWS           = 100000
	LOG_UPLOADED_DIR_PATH          = "uploaded/"
	LOGGING_GY_POD_DEFAULT_QTY     = 9 //Suffix適用する数:開始は0から行う→マイナス1、利用しない場合(k8s.Deplomentの利用)は0を指定

	/////【【【コンテント追加時に必須追加】処理名】】/////
	CONTENT_NAME_CREATE_FILE_GYROSCOPE = "create_file_gyroscope"
	CONTENT_NAME_DIVIDE_FILE_GYROSCOPE = "divide_file_gyroscope"
	CONTENT_NAME_UPLOAD_FILE_GYROSCOPE = "upload_file_gyroscope"
	CONTENT_NAME_DELETE_FILE_GYROSCOPE = "delete_file_gyroscope"

	/////【【【コンテント追加時に必須追加】実行引数】上記のARGS項目の値に対して、実行時に引き当てたい場合】/////
	//fligth-logbookでは問題の無い処理が本Jobでは、引数が入って来ない現象？？その対応
	//上記の問題の暫定対策として、環境変数にFORCE_INDEXを追加ここで、1API=1サーバー単位にて、INDEXを強制しながら構築させる
	ARGS_VALUE_CONTENTS           = "contents: create_file_gyroscope, divide_file_gyroscope, upload_file_gyroscope, delete_file_gyroscope" //必須設定事項
	ARGS_VALUE_FORCE_INDEX        = "force_index:"                                                                                         //デフォルトvalue指定なしでOK)
	ARGS_VALUE_FREQUENCY          = "frequency:"                                                                                           //デフォルトvalue指定なしでOK)
	ARGS_VALUE_TERM               = "term:"                                                                                                //デフォルトvalue指定なしでOK)
	ARGS_VALUE_BUCKET_NAME        = "bucket_name:"                                                                                         //デフォルトvalue指定なしでOK)
	ARGS_VALUE_BUCKET_MIDDLE_PATH = "bucket_middle_path:"                                                                                  //デフォルトvalue指定なしでOK)
	ARGS_VALUE_BUCKET_PREFIX      = "bucket_prefix:"                                                                                       //デフォルトvalue指定なしでOK)
	ARGS_VALUE_BUCKET_SUFFIX      = "bucket_suffix:"                                                                                       //デフォルトvalue指定なしでOK)
	ARGS_VALUE_BUCKET_TIME_JST    = "bucket_time_jst:"                                                                                     //デフォルトvalue指定なしでOK)
	ARGS_VALUE_EXTRACT_START_DATE = "extract_start_date:"                                                                                  //デフォルトvalue指定なしでOK)
	ARGS_VALUE_EXTRACT_END_DATE   = "extract_end_date:"                                                                                    //デフォルトvalue指定なしでOK)
)

// /////////////////////////////////////////////////
var configMapEnv map[string]string  //環境変数の箱
var configMapArgs map[string]string //CMDライン引数の箱
var uuv4Tokens []string             //サーバー認証のためのTokenの箱

// /////////////////////////////////////////////////
// 起動時にGCP ProjectID、NS, Kindを登録する
func init() {

	//Set environ values
	NewConfigEnv()

	//output message finish config settings.
	initString := fmt.Sprintf("[Project:%s][ServerCode:%s][Appli:%s][Env:%s][FroceIndex:%s]",
		configMapEnv[ENV_GCP_PROJECT_ID],
		configMapEnv[ENV_SERVER_CODE],
		configMapEnv[ENV_APPLI_NAME],
		configMapEnv[ENV_ENV],
		configMapEnv[ENV_FORCE_INDEX])
	fmt.Println(LOG.SetLogEntry(LOG.INFO, "LOG-jobs INIT", initString))

	//Verion print
	initString = fmt.Sprintf("LOG-jobs component version :%s", SYSTEM_COMPONENT_VERSION)
	fmt.Println(LOG.SetLogEntry(LOG.INFO, "LOG-jobs STARTING", initString))

	//Set CMD Args values
	NewConfigArgs()
}

///////////////////////////////////////////////////
/* =================================
	環境変数の格納
		$PORT
		$GCP_PROJECT_ID
		$SERVER_CODE
		$APPLI_NAME
		$ENV
		$SERVICE_OR_JOBS
* ================================= */
func NewConfigEnv() {

	//環境変数をMapping
	configMapEnv = make(map[string]string)
	configMapEnv[ENV_LOCAL_OR_REMOTE] = os.Getenv("LOCAL_OR_REMOTE")
	configMapEnv[ENV_GCP_PROJECT_ID] = os.Getenv("GCP_PROJECT_ID")
	configMapEnv[ENV_SERVER_CODE] = os.Getenv("SERVER_CODE")
	configMapEnv[ENV_APPLI_NAME] = os.Getenv("APPLI_NAME")
	configMapEnv[ENV_ENV] = os.Getenv("ENV")
	configMapEnv[ENV_FORCE_INDEX] = os.Getenv("FORCE_INDEX")

	//mount先のNFSパス
	if configMapEnv[ENV_LOCAL_OR_REMOTE] == "local" {
		configMapEnv[LOG_DIR_PATH] = LOG_STORAGE_DIR_ABSOLUTE_PATH
	} else {
		configMapEnv[LOG_DIR_PATH] = LOG_STORAGE_VOLUME_DIR_PATH
	}

}

///////////////////////////////////////////////////
/* =================================
	環境変数の返却
* ================================= */
func GetConfigEnv(name string) string {
	return configMapEnv[name]
}
func GetConfigEnvAll() map[string]string {
	return configMapEnv
}

///////////////////////////////////////////////////
/* =================================
	コマンドライン引数の格納
		force_index:強制的実行
			"0"or "1" or ...
		frequency:jobの周期
			"daily" or "monthly" などのFrequencyを設定する
		term:ログファイルの時間間隔=ファイル名
			15 or 30 or 60
		contents:jobの種類("content1,content2,,,")
			contentをカンマつなぎで指定
			これを配列に分割し、JOBS実行には、CLOUD_RUN_TASK_INDEXと一致させて並列実行を実現行わせる
		その他の項目について
			強制的に、0の要素に対する引数と利用させたい場合にのみ指定する
			用途：検証、バッチ処理(Daily, Monthly)以外のマイグレーションなど
				(*=指定がなければDefaultが指定される)
				bucket_middle_path:*
				bucket_name:*
				bucket_prefix:*
				bucket_suffix:*
				bucket_time_jst:*
				extract_start_date:*
				extract_end_date:*
* ================================= */
func NewConfigArgs() {

	//起動時の引数から取得
	flag.Parse()
	args := flag.Args()

	fmt.Println("ARGS:", args)

	//外側からコマンドライン引数が適用されない場合、configの値を利用する
	//fligth-logbookでは問題の無い処理が、本Jobでは、コンテナ引数が入って来ない現象？？その対応
	if len(args) == 0 {
		ForseSetConfigMapArgs() //コマンドライン引数を強制的にConfigの値(ARGS_VALUE_*)でMapping
	} else {
		err := SetConfigMapArgs(args) //コマンドライン引数をMapping
		if err != nil {
			log.Fatal(err)
		}
	}
}

///////////////////////////////////////////////////
/* =================================
コマンドライン引数を"key:value"に分割しMapping、格納する
* ================================= */
func SetConfigMapArgs(args []string) error {

	if len(configMapArgs) == 0 {
		configMapArgs = make(map[string]string)
	}

	for _, a := range args {
		if a == "" {
			continue
		}

		//コロンでスプリットをしkey-valueに変換
		keys := strings.Split(a, ":")[0]
		value := strings.Split(a, ":")[1]

		//さらにハイフンでスプリット(contentを導く)
		key := strings.Split(keys, "-")

		//Mapping
		switch key[0] {
		case ARGS_CONTENTS:
			configMapArgs[ARGS_CONTENTS] = strings.ReplaceAll(value, " ", "")
		case ARGS_FORCE_INDEX:
			configMapArgs[ARGS_FORCE_INDEX] = value
		case ARGS_FREQUENCY:
			configMapArgs[ARGS_FREQUENCY] = value
		case ARGS_TERM:
			configMapArgs[ARGS_TERM] = value
		case ARGS_BUCKET_NAME:
			configMapArgs[ARGS_BUCKET_NAME] = value
		case ARGS_BUCKET_MIDDLE_PATH:
			configMapArgs[ARGS_BUCKET_MIDDLE_PATH] = value
		case ARGS_BUCKET_PREFIX:
			if value != "" {
				value = value + "_"
			}
			configMapArgs[ARGS_BUCKET_PREFIX] = value
		case ARGS_BUCKET_SUFFIX:
			if value != "" {
				value = "_" + value
			}
			configMapArgs[ARGS_BUCKET_SUFFIX] = value
		case ARGS_BUCKET_TIME_JST:
			configMapArgs[ARGS_BUCKET_TIME_JST] = value
		case ARGS_EXTRACT_START_DATE:
			configMapArgs[ARGS_EXTRACT_START_DATE] = value
		case ARGS_EXTRACT_END_DATE:
			configMapArgs[ARGS_EXTRACT_END_DATE] = value
		default:
			return fmt.Errorf("【Error】[func:%s][Args:%s][%s]", "src/config/config.go/setConfigMapArgs()", "コマンドライン引数のキーが一致しません。", key)
		}
	}
	LOG.JustLogging(fmt.Sprintf("ConfigMaoArgs:[%v]", configMapArgs))
	return nil
}

///////////////////////////////////////////////////
/* =================================
コンフィグで指定したARGS_VALUE_*を"key:value"に分割しMapping、格納する
* ================================= */
func ForseSetConfigMapArgs() {
	var args []string
	args = append(args, ARGS_VALUE_CONTENTS)
	args = append(args, ARGS_VALUE_FREQUENCY)
	args = append(args, ARGS_VALUE_TERM)
	args = append(args, ARGS_VALUE_BUCKET_NAME)
	args = append(args, ARGS_VALUE_BUCKET_MIDDLE_PATH)
	args = append(args, ARGS_VALUE_BUCKET_SUFFIX)
	args = append(args, ARGS_VALUE_BUCKET_TIME_JST)
	args = append(args, ARGS_VALUE_EXTRACT_START_DATE)
	args = append(args, ARGS_VALUE_EXTRACT_END_DATE)

	//logging-jobsのClouRunJobs環境でコマンド引数が取得されない事象から
	//運用を考慮し、本意ではないが下記ARGSを固定化させる
	//force_index, bucket_prefix
	//force_index:
	force_index := GetConfigEnv(ENV_FORCE_INDEX)
	args = append(args, ARGS_VALUE_FORCE_INDEX+force_index)

	//bucket_prefix:
	env := GetConfigEnv(ENV_ENV)
	if env == "prd" {
		//"bucket_prefix:"
		args = append(args, ARGS_VALUE_BUCKET_PREFIX)
	} else {
		//"bucket_prefix:[env]"
		args = append(args, ARGS_VALUE_BUCKET_PREFIX+env)
	}

	SetConfigMapArgs(args)
}

///////////////////////////////////////////////////
/* =================================
	コマンドライン引数の返却
* ================================= */
func GetConfigArgs(name string) string {
	return configMapArgs[name]
}
func GetConfigArgsAll() map[string]string {
	return configMapArgs
}
func GetConfigArgsAllString() string {
	var s string
	for k, v := range configMapArgs {
		s += k + ":" + v + ", "
	}
	return s
}
func GetConfigArgsAllKeyValue() (key []string, val []string) {
	var retK []string
	var retV []string
	for k, v := range configMapArgs {
		retK = append(retK, k)
		retV = append(retV, v)
	}
	return retK, retV
}

///////////////////////////////////////////////////
/* =================================
	//Configの返却
* ================================= */
func GetConfig(name string) string {
	return configMapEnv[name]
}
func GetConfigAll() map[string]string {
	return configMapEnv
}

///////////////////////////////////////////////////
/* =================================
	サーバー間認証に用いるUUV4トークンをJSONファイルから取得しておく
* ================================= */
func NewUuv4Tokens() {

	//箱を準備
	type Uuv4TokenJson struct {
		Uuv4tokens []string `json:"uuv4tokens"`
	}

	//Rootディレクトリを取得して、tokensのJSONファイルの絶対パスを指定
	var (
		_, b, _, _ = runtime.Caller(0)
		root       = filepath.Join(filepath.Dir(b), "../../")
	)
	path := root + "/cmd/authorization/uuv4tokens.json"

	// JSONファイル読み込み
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// JSONデコード
	var tokens Uuv4TokenJson
	if err := json.Unmarshal(bytes, &tokens); err != nil {
		log.Fatal(err)
	}
	// デコードしたデータを表示
	for _, t := range tokens.Uuv4tokens {
		uuv4Tokens = append(uuv4Tokens, t)
	}
}
func GetUuv4Tokens() []string {
	return uuv4Tokens
}
