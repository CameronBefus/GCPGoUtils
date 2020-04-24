package pgdb

import (
	"context"
	"crypto/tls"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"

	//"github.com/jackc/pgx/log/logrusadapter"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MyDB - global reference
//var MyDB *sql.DB
var DBCon *pgxpool.Pool
var connectionSuccessful = false
var CTxt = context.Background()

const maxConnections = 20

// IsConnected - returns true if we have a valid connection
func IsConnected() bool {
	return connectionSuccessful
}

// Setup - primary method for initializing Google Cloud Mysql database connection
// production flag indicates if the connection is coming from within the cloud or externally
func Setup(local bool) bool {
	if !IsConnected() {
		connectionSuccessful = connect(local)
	}
	return connectionSuccessful
}

func configureTLS() *tls.Config {
	cert, err := tls.LoadX509KeyPair(viper.GetString("TLS_CLIENT_CERT"), viper.GetString("TLS_CLIENT_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}
}

// connect - returns false if unable to connect successfully
func connect(local bool) bool {

	var tlsc *tls.Config
	var cs string
	if local {
		log.Info("Configuring TLS...")
		cs = viper.GetString(`CLOUDSQL_LOCAL`)
		tlsc = configureTLS()
		tlsc.InsecureSkipVerify = true
	} else {
		cs = viper.GetString(`CLOUDSQL`)
	}
	cfg, err := pgxpool.ParseConfig(cs)
	if err != nil {
		log.Errorf("Unable to parse connection parameters: %v", err)
		return false
	}
	cfg.ConnConfig.TLSConfig = tlsc
	cfg.MaxConns = maxConnections

	//cfg.ConnConfig.Logger = logrusadapter.NewLogger( log.StandardLogger())
	//cfg.ConnConfig.LogLevel = pgx.LogLevelDebug

	DBCon, err = pgxpool.ConnectConfig(CTxt, cfg)
	if err != nil {
		log.Errorf("Unable to establish connection: %v", err)
		return false
	}
	return true
}

// Close - shut down the current database connection
func Close() {

	if DBCon != nil {
		DBCon.Close()
	}
	connectionSuccessful = false
}

// use this format to prepare time stamps to be inserted into the database
//func formatDateTime(ts time.Time) string {
//	return ts.Format("2006-01-02 15:04:05")
//}

// GetCount - execute a sql query that returns a single integer value
func GetCount(q string, args ...interface{}) (int, error) {
	count := 0
	rows, err := DBCon.Query(CTxt, q, args...)
	if err != nil {
		return 0, err
	} else {
		rows.Next()
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		rows.Close()
	}
	return count, nil
}

func Query(q string, args ...interface{}) (pgx.Rows, error) {
	return DBCon.Query(CTxt, q, args...)
}

func Execute(q string, args ...interface{}) (int, error) {
	commandTag, err := DBCon.Exec(CTxt, q, args...)
	if err == nil {
		return int(commandTag.RowsAffected()), nil
	}
	return 0, err
}

func BToI(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IToB(i int) bool {
	return i != 0
}

//IsDuplicate - returns true if err is mysql duplicate entry notification
func IsDuplicate(err error) bool {
	me, ok := err.(*pgconn.PgError)
	return ok && me.Code == pgerrcode.UniqueViolation
}

func IsForeignKeyConstraint(err error) bool {
	me, ok := err.(*pgconn.PgError)
	return ok && me.Code == pgerrcode.ForeignKeyViolation
}

func Optimize() error {
	_, err := Execute(`Vacuum Analyze`)
	return err
}
