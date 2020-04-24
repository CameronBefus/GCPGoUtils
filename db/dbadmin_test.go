package db

import (
	"bitbucket.org/mjlogs/utilprg/utilities/util"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
	util.InitializeViper(`test_db`)
}

func Test_Setup(t *testing.T) {

	if !Setup(true) {
		t.Error(`Failed to connect`)
		return
	}

	val := GetCount(`select 39`)
	if val != 39 {
		t.Error(`Failed to execute simple query`)
	}
}
