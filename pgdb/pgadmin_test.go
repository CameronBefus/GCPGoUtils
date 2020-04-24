package pgdb

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

	val, err := Execute(`update log_type set log_group = 'x' where log_type_id = $1`, -1)
	if val != 0 && err != nil {
		t.Error(`Failed to execute simple query`)
	}

	cnt, err := Execute(`select count(*) from log_type where log_type_id < $1`, 21)
	if cnt != 20 && err != nil {
		t.Error(`Failed to get count`, cnt, err)
	}
}

func Test_ParallelCalls(t *testing.T) {

	if !Setup(true) {
		t.Error(`Failed to connect`)
		return
	}

	e := Optimize()
	if e != nil {
		t.Error(e)
	}

	q := `select mjw_id from well limit 10`
	rows, err := Query(q)
	if err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			_, e1 := GetCount(`select count(*) from logs where mjw_id = 33346613`)
			if e1 != nil {
				t.Error(e1)
			}
		}
		rows.Close()
	}

	mc := DBCon.Stat().MaxConns()
	if mc != maxConnections {
		t.Error(`unexpected maxConnections: `, mc)
	}

}
