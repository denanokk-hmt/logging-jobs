/*
======================
GCSファイルアップロード
========================
*/
package gcs

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"

	LOG "bwing.app/src/log"
)

type GcsUpload struct {
	BucketName    string //GCSバケット名
	BucketDirPath string //GCSバケットディレクトリ
	DirPath       string //Uploadするファイルのディレクトリ
	FileName      string //Uploadするファイル名
}

///////////////////////////////////////////////////
/* ===========================================
ログファイルをGCSバケットにアップロード
=========================================== */
func (g GcsUpload) UploadFiletoGcsBucket() error {

	//ログを書き込まれたファイルを取得
	f, err := os.Open(g.DirPath + g.FileName)
	if err != nil {
		return err
	}
	defer f.Close()

	//Contextを設定
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	//GCS clientを生成
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	//middleパス+年月日+ファイル名
	objectPath := g.BucketDirPath + "/" + g.FileName

	LOG.JustLogging(fmt.Sprintf("BUCKET NAME:[%s], ObjDirPath:[%s]", g.BucketName, objectPath))

	//オブジェクトのWriterを作成し、GCSバケットにファイルをアップロード
	writer := client.Bucket(g.BucketName).Object(objectPath).NewWriter(ctx)
	if _, err := io.Copy(writer, f); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}
