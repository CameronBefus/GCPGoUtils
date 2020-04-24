package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"io/ioutil"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var storageClient *storage.Client
var defaultBucket *storage.BucketHandle

type BucketHandlePtr *storage.BucketHandle

const envBucketName = `CLOUD_STORAGE_BUCKET`
const storageCred = `STORAGE_CRED`

var credentialsFile string

func getStorageClient() *storage.Client {
	if storageClient == nil {
		ctx := context.Background()
		if getCredentialsFile() != `` {
			var err error
			storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile(getCredentialsFile()))
			if err != nil {
				log.Error(fmt.Errorf(`error initializing storage client with credentials file %s, Error: %s`, credentialsFile, err.Error()))
			}
		} else {
			var err error
			storageClient, err = storage.NewClient(ctx)
			if err != nil {
				log.Error(fmt.Errorf(`error initializing storage client : %s`, err.Error()))
			}
		}

	}
	return storageClient
}

// GetDefaultBucket - usually called internally
func GetDefaultBucket() BucketHandlePtr {
	if defaultBucket == nil {
		bucketName := viper.GetString(envBucketName)
		if len(bucketName) == 0 {
			log.Warning(`failed to get default Storage Bucket name`)
		}
		defaultBucket = getStorageClient().Bucket(bucketName)
	}
	return defaultBucket
}

func getCredentialsFile() string {
	if len(credentialsFile) == 0 {
		credentialsFile = viper.GetString(storageCred)
	}
	return credentialsFile
}

// GetBucket - return a bucket handle
func GetBucket(bn string) BucketHandlePtr {
	return getStorageClient().Bucket(bn)
}

// GetFilesB - return list of files within specified bucket / path
func GetFilesB(b *storage.BucketHandle, path string) []string {
	var result []string
	var q = storage.Query{Prefix: path}
	it := b.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Warning(err)
			break
		}
		result = append(result, objAttrs.Name)
		// fmt.Printf("objects in bucket: %+v", objAttrs)
	}
	return result
}

// GetFilesBF - walk through the objects within specified bucket / path, passing each to a function
// continues as long as result is true
func GetFilesBF(b *storage.BucketHandle, path string, pf func(oa *storage.ObjectAttrs) bool) {
	var q = storage.Query{Prefix: path}
	it := b.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Warning(err)
			break
		}
		if !pf(objAttrs) {
			break
		}
	}
	return
}

// DeleteOldFiles - delete all files within the specified path that are older than age in hours
func DeleteOldFiles(path string, ageh int) int {
	return DeleteOldFilesB(GetDefaultBucket(), path, ageh)
}

// DeleteOldFilesB - delete all files within the specified path that are older than age in hours
func DeleteOldFilesB(b *storage.BucketHandle, path string, ageh int) int {
	timeLimit := time.Now().Add(time.Duration(ageh*-1) * time.Hour)
	result := 0
	oa := GetFileInfoB(b, path)
	for _, attrs := range oa {
		if attrs.Created.Before(timeLimit) {
			DeleteCloudFileB(b, attrs.Name)
			result++
		}
	}
	return result
}

// GetFiles - returns slice of files within the default bucket
func GetFiles(path string) []string {
	return GetFilesB(GetDefaultBucket(), path)
}

// GetFileInfo - returns slice of files within the bucket
func GetFileInfo(path string) []storage.ObjectAttrs {
	return GetFileInfoB(GetDefaultBucket(), path)
}

// GetFileInfoB -
func GetFileInfoB(b *storage.BucketHandle, path string) []storage.ObjectAttrs {
	var result []storage.ObjectAttrs
	var q = storage.Query{Prefix: path}
	it := b.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Warning(err)
			break
		}
		result = append(result, *objAttrs)
	}
	return result
}

// GetFilesWithSuffix - returns list of files from the the specified path where the file name ends with suffix
func GetFilesWithSuffix(path string, suffix string) []string {
	return GetFilesWithSuffixB(GetDefaultBucket(), path, suffix)
}

// GetFilesWithSuffixB - returns list of files from the the specified path where
// the file name ends with suffix
func GetFilesWithSuffixB(b *storage.BucketHandle, path string, suffix string) []string {
	var result []string
	var q = storage.Query{Prefix: path}
	it := b.Objects(context.Background(), &q)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Warning(err)
			break
		}
		if strings.HasSuffix(objAttrs.Name, suffix) {
			result = append(result, objAttrs.Name)
		}
	}
	return result
}

// GetFileReader - remember to close the Reader after use. returns nil if file not found
// second parameter is file size in bytes, if found
func GetFileReader(fn string) (*storage.Reader, int64) {
	return GetFileReaderB(GetDefaultBucket(), fn)
}

// GetFileReaderB - remember to close the Reader after use. returns nil if file not found
// second parameter is file size in bytes, if found
func GetFileReaderB(b *storage.BucketHandle, fn string) (*storage.Reader, int64) {
	it := b.Object(fn)
	var fsize int64
	ita, e1 := it.Attrs(context.Background())
	if e1 != nil {
		log.Error(e1, fn)
	} else {
		fsize = ita.Size
	}
	r, err := it.NewReader(context.Background())
	if err != nil {
		if err != storage.ErrObjectNotExist {
			log.Error(err, fn)
		}
		return nil, 0
	}
	return r, fsize
}

// DeleteCloudFile -
func DeleteCloudFile(fn string) bool {
	return DeleteCloudFileB(GetDefaultBucket(), fn)
}

// DeleteCloudFileB -
func DeleteCloudFileB(b *storage.BucketHandle, fn string) bool {
	it := b.Object(fn)
	err := it.Delete(context.Background())
	if err == nil {
		return true
	}
	log.Error(err)
	return false
}

// FileExists -
func FileExists(fn string) bool {
	return FileExistsB(GetDefaultBucket(), fn)
}

// FileExistsB -
func FileExistsB(b *storage.BucketHandle, fn string) bool {
	it := b.Object(fn)
	_, err := it.Attrs(context.Background())
	return err == nil
}

// WriteFile creates a file in Google Cloud Storage.
func WriteFile(fn string, content string) bool {
	return WriteFileB(GetDefaultBucket(), fn, content)
}

// WriteFileB creates a file in Google Cloud Storage.
func WriteFileB(b *storage.BucketHandle, fn string, content string) bool {
	return WriteCloudFileB(b, fn, []byte(content), "text/plain")
}

// WriteCloudFile - write data to a file in the google cloud
// fn is the dest filename/path
// content - what to write
// ftype is the Mime contentType
func WriteCloudFile(fn string, content []byte, ftype string) bool {
	return WriteCloudFileB(GetDefaultBucket(), fn, content, ftype)
}

// WriteCloudFileB - write data to a file in the google cloud
// fn is the dest filename/path
// content - what to write
// ftype is the Mime contentType
func WriteCloudFileB(b *storage.BucketHandle, fn string, content []byte, ftype string) bool {
	// ita, _ := b.Attrs(context.Background())
	wc := b.Object(fn).NewWriter(context.Background())
	wc.ContentType = ftype

	if _, err := wc.Write(content); err != nil {
		log.Errorf("unable to write data to file %s: %v", fn, err)
		return false
	}

	if err := wc.Close(); err != nil {
		log.Errorf("unable to close file %s: %v", fn, err)
		return false
	}
	return true
}

// CopyFile - from locations within the cloud
func CopyFile(fn string, dest string) bool {
	return CopyFileB(GetDefaultBucket(), GetDefaultBucket(), fn, dest)
}

// CopyFileB - from locations within the cloud
func CopyFileB(b *storage.BucketHandle, t *storage.BucketHandle, fn string, dest string) bool {
	s := b.Object(fn)
	d := t.Object(dest)

	_, err := d.CopierFrom(s).Run(context.Background())
	if err != nil {
		log.Errorf("failed to copy from %s/%s to %s/%s: %v", s.BucketName(), fn, d.BucketName(), dest, err)
		return false
	}
	return true
}

// CreateDownloadURL - create a signed, time limited url to access the specified file
// gsutil signurl -d 1d  c:\Users\camer\.ssh\GoogleStorageCredentials.json gs://proven-impact-212716.appspot.com/images/whoops.bmp
// https://cloud.google.com/storage/docs/access-control/signing-urls-manually
// https://cloud.google.com/storage/docs/authentication/canonical-requests
// specifying the content type when requesting a download url seems to cause issues
func CreateDownloadURL(b *storage.BucketHandle, minutes int, path string) (string, error) {
	// get credentials file name
	f := b.Object(path)
	jsonKey, err := ioutil.ReadFile(getCredentialsFile())
	if err != nil {
		dir, _ := os.Getwd()
		log.Errorf("cannot read the JSON key file '%s/%s', err: %v", dir, getCredentialsFile(), err)
		return "", err
	}

	conf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		log.Errorf("google.JWTConfigFromJSON: %v", err)
		return "", err
	}

	opts := &storage.SignedURLOptions{
		Method:         "GET",
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        time.Now().Add(time.Duration(minutes) * time.Minute),
	}

	u, err := storage.SignedURL(f.BucketName(), path, opts)
	if err != nil {
		log.Errorf("Unable to generate a signed URL: %v", err)
		return "", err
	}
	return u, nil
}

// DownloadFiles - assumes list of files contains folder names
// dest is local file path
func DownloadFiles(files []string, dest string) error {
	for _, fn := range files {
		cr, _ := GetFileReader(fn)
		if cr != nil {
			c2, e1 := ioutil.ReadAll(cr)
			if e1 != nil {
				return fmt.Errorf("Failed to read file %s : %s ", fn, e1)
			}
			cr.Close()

			fb := filepath.Base(fn)
			e2 := ioutil.WriteFile(dest+fb, c2, 0644)
			if e2 != nil {
				return fmt.Errorf("Failed to write file to %s : %s ", dest+fb, e2)
			}
		}
	}
	return nil
}
