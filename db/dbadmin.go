package db

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	sqlDriver "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MySQL - global reference
var MySQL *sql.DB
var connectionSuccessful = false
var psLock sync.RWMutex
var preparedStmts = map[string]*sql.Stmt{}

// IsConnected - returns true if we have a valid connection
func IsConnected() bool {
	return connectionSuccessful
}

// GetPreparedStatement - returns a previously prepared statement in a thread safe manner
func GetPreparedStatement(key string) *sql.Stmt {
	psLock.RLock()
	result := preparedStmts[key]
	psLock.RUnlock()
	return result
}

// SavePreparedStatement - adds a new prepared statement to the 'saved' list
func SavePreparedStatement(key string, s *sql.Stmt) *sql.Stmt {
	psLock.Lock()
	if s == nil {
		delete(preparedStmts, key)
	} else {
		preparedStmts[key] = s
	}
	psLock.Unlock()
	return s
}

// Setup - primary method for initializing Google Cloud Mysql database connection
// production flag indicates if the connection is coming from within the cloud or externally
func Setup(local bool) bool {
	if !IsConnected() {
		connectionSuccessful = connect(local)
	}
	return connectionSuccessful
}

func configureTLS() {
	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(viper.GetString("TLS_SERVER_CA"))
	if err != nil {
		log.Fatal(err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to append PEM.")
	}
	clientCert := make([]tls.Certificate, 0, 1)
	certs, err := tls.LoadX509KeyPair(viper.GetString("TLS_CLIENT_CERT"), viper.GetString("TLS_CLIENT_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	clientCert = append(clientCert, certs)
	e1 := sqlDriver.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:            rootCertPool,
		Certificates:       clientCert,
		InsecureSkipVerify: true,
	})
	if e1 != nil {
		log.Panic(e1)
	}
}

// connect - returns false if unable to connect successfully
func connect(local bool) bool {
	var cs string
	if local {
		log.Info("Configuring TLS...")
		cs = viper.GetString(`CLOUDSQL_LOCAL`)
		configureTLS()
	} else {
		cs = viper.GetString(`CLOUDSQL`)
	}
	// sql.Open appears to do nothing, succeeds even if db unavailable
	msql, err := sql.Open("mysql", cs)
	if err != nil {
		log.Error(err, ` open=> `, cleanConnectionString(cs))
		return false
	}

	err = msql.Ping()
	if err != nil {
		log.Error(err, ` ping=> `, cleanConnectionString(cs))
		return false
	}

	MySQL = msql
	return true
}

// cleanConnectionString - strip userid/password from logfiles
func cleanConnectionString(cs string) string {
	x := strings.Index(cs, `@`)
	return cs[x:]
}

// ClosePreparedStatements - only used if you wish to clear prepared statements without killing your database connection
func ClosePreparedStatements() {
	for k, v := range preparedStmts {
		_ = v.Close()
		SavePreparedStatement(k, nil)
	}
}

// Close - shut down the current database connection
func Close() {

	if MySQL != nil {
		_ = MySQL.Close()
		MySQL = nil
	}
	connectionSuccessful = false
}

// use this format to prepare time stamps to be inserted into the database
func formatDateTime(ts time.Time) string {
	return ts.Format("2006-01-02 15:04:05")
}

// GetStats -
// MaxOpenConnections int // Maximum number of open connections to the database.
// // Pool Status
// OpenConnections int // The number of established connections both in use and idle.
// InUse           int // The number of connections currently in use.
// Idle            int // The number of idle connections.
// // Counters
// WaitCount         int64         // The total number of connections waited for.
// WaitDuration      time.Duration // The total time blocked waiting for a new connection.
// MaxIdleClosed     int64         // The total number of connections closed due to SetMaxIdleConns.
// MaxLifetimeClosed int64         // The total number of connections closed due to SetConnMaxLifetime.
func GetStats() sql.DBStats {
	return MySQL.Stats()
}

// GetCount - execute a sql query that returns a single integer value
func GetCount(q string, args ...interface{}) int {
	var result = 0
	err := MySQL.QueryRow(q, args...).Scan(&result)
	if err != nil {
		log.Error(err)
	}
	return result
}

// IsDuplicate - returns true if err is mysql duplicate entry notification
func IsDuplicate(err error) bool {
	me, ok := err.(*sqlDriver.MySQLError)
	return ok && me.Number == 1062

}

func IsForeignKeyConstraint(err error) bool {
	me, ok := err.(*sqlDriver.MySQLError)
	return ok && me.Number == 1452
}
