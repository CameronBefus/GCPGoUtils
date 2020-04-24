package storage

import (
	"bitbucket.org/mjlogs/utilprg/utilities/util"
	"bufio"
	"cloud.google.com/go/storage"
	"encoding/csv"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	util.InitializeViper(`storage`)
}

func Test_FileExists(t *testing.T) {
	if !FileExists(`images/whoops.bmp`) {
		t.Error(`cloud storage verification function failed`)
	}
}

func Test_getFiles(t *testing.T) {
	fl := GetFiles("admin_update_csv/")
	if len(fl) == 0 {
		t.Error("GCP Storage: failed to find any files")
	}
}

func Test_GetFilesBF(t *testing.T) {
	fcnt := 0

	var pf = func(oa *storage.ObjectAttrs) bool {
		fcnt++
		return true
	}

	GetFilesBF(GetDefaultBucket(), "admin_update_csv/", pf)
	if fcnt == 0 {
		t.Error("GetFilesBF: failed to find any files")
	}
}

func Test_copy(t *testing.T) {
	content := `this is sample content \n for the file`
	fn := `testing/cpy.txt`
	if !WriteFile(fn, content) {
		t.Error(fn)
	}
	fn2 := `testing2/cpy2.txt`
	if !CopyFile(fn, fn2) {
		t.Error("copy failed")
	}

	ok := DeleteCloudFile(fn)
	if !ok {
		t.Error("Failed to delete1")
	}
	ok2 := DeleteCloudFile(fn2)
	if !ok2 {
		t.Error("Failed to delete2")
	}
}

func Test_wred(t *testing.T) {
	content := `this is sample content \n for the file`
	fn := `testing/wrd.txt`

	if !WriteFile(fn, content) {
		t.Error(fn)
	}

	cr, _ := GetFileReader(fn)
	defer cr.Close()
	c2, e2 := ioutil.ReadAll(cr)
	if e2 != nil {
		t.Error(e2)
	}

	if !FileExists(fn) {
		t.Error(`Expected file to exist`)
	}

	if content != string(c2) {
		t.Errorf(`Expected "%s" got "%s"`, content, c2)
	}

	ok := DeleteCloudFile(fn)
	if !ok {
		t.Error("Failed to delete")
	}

	if FileExists(fn) {
		t.Error(`Expected file NOT to exist`)
	}

}

func Test_getFileReader(t *testing.T) {
	cr, sz := GetFileReader(`admin_update_csv/test_wellData.csv`)
	defer cr.Close()

	if sz == 0 {
		t.Error("failed to get file size")
	}

	r := csv.NewReader(bufio.NewReader(cr))
	r.Comment = '#'
	r.FieldsPerRecord = 9
	r.ReuseRecord = true
	var rcnt = 0
	for {
		_, e2 := r.Read()
		if e2 == io.EOF {
			break
		} else if e2 != nil {
			log.Fatal(e2)
		}
		// fmt.Println(line)
		rcnt++
	}
	if rcnt < 26 {
		t.Error("failed to read 26 lines")

	}
}

func Test_DeleteOldFiles(t *testing.T) {
	e := writeTestFiles()
	if e != nil {
		t.Error(e)
	}
	i1 := DeleteOldFiles(`testing/`, 1)
	if i1 != 0 {
		t.Errorf("Expected 0, got %d", i1)
	}
	i2 := DeleteOldFiles(`testing/`, -1)
	if i2 != 2 {
		t.Errorf("Expected 2, got %d", i2)
	}
	r := GetFileInfo(`testing/`)
	if len(r) != 0 {
		t.Error(`Expected 0 files to be left`)
	}
}

const tf1 = `testing/a1.txt`
const tf2 = `testing/a2.json`

func writeTestFiles() error {
	content := `this is sample content \n for the file`
	if !WriteFile(tf1, content) {
		return errors.New("unable to write to " + tf1)
	}
	if !WriteCloudFile(tf2, []byte(content), `application/json`) {
		return errors.New("unable to write to " + tf2)
	}
	return nil
}

func deleteTestFiles() {
	ok := DeleteCloudFile(tf1)
	if !ok {
		panic("Failed to delete " + tf1)
	}
	ok2 := DeleteCloudFile(tf2)
	if !ok2 {
		panic("Failed to delete " + tf2)
	}
}

func Test_DownloadFiles(t *testing.T) {
	e := writeTestFiles()
	if e != nil {
		t.Error(e)
	}
	const localDir = `./`
	flist := GetFiles(`testing/`)
	err := DownloadFiles(flist, localDir)
	if err != nil {
		t.Error(`DownloadFiles: `, err)
	}
	for _, i2 := range flist {
		fb := filepath.Base(i2)
		if !util.FileExists(localDir + fb) {
			t.Error(`DownloadFiles - failed to find local file`)
		}
		_ = os.Remove(localDir + fb)
	}
	deleteTestFiles()
}

func Test_getFilesWithSuffix(t *testing.T) {
	e := writeTestFiles()
	if e != nil {
		t.Error(e)
	}
	defer deleteTestFiles()

	f1 := GetFiles(`testing/`)
	f2 := GetFilesWithSuffix(`testing/`, `.txt`)

	if (len(f1) != 2) || (len(f2) != 1) {
		t.Errorf("Expected 2 & 1, got %d & %d", len(f1), len(f2))
	}
}

func Test_CreateDownloadURL(t *testing.T) {
	e := writeTestFiles()
	if e != nil {
		t.Error(e)
	}
	r := GetFiles(`testing`)
	defer deleteTestFiles()
	//b := GetBucket(`mj-wls-dev-lidb`)
	//r := GetFilesB(b, `testkey1`)
	if len(r) == 0 {
		t.Error(`No files found in bucket`)
	} else {

		u, e1 := CreateDownloadURL(defaultBucket, 5, r[1])
		if e1 != nil {
			t.Errorf("Got error %v", e1)
		} else {
			_ = ioutil.WriteFile("./surl_test.txt", []byte(u), 0644)
		}
	}
}

func Test_GetFileInfo(t *testing.T) {
	e := writeTestFiles()
	if e != nil {
		t.Error(e)
	}
	defer deleteTestFiles()
	r := GetFileInfo(`testing/`)
	if len(r) != 2 {
		t.Errorf(`Expected 2 files found for GetFileInfo, got %d `, len(r))
	}
	if r[1].ContentType != `application/json` {
		t.Errorf(`Unexpected content type - %s`, r[1].ContentType)
	}
	if r[1].Size != 38 {
		t.Errorf(`Unexpected filesize - %d`, r[1].Size)
	}

}
