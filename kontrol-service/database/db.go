package database

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Run scripts/genereate_db_snapshot.sh when the cloud backend deployment results in a schema change so the cloud backend package seeds the DB with recent data.

const (
	defaultLockTimeout     = 10 * time.Second
	lockObjectName         = "db-migrate"
	unlockRetryAttempts    = 10
	unlockRetryWaitSeconds = 5
)

type Db struct {
	db *gorm.DB
}

func NewDb(
	dbConnInfo *DatabaseConnectionInfo,
) (*Db, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", dbConnInfo.host, dbConnInfo.username, dbConnInfo.password, dbConnInfo.databaseName, dbConnInfo.port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening the connection to the database with dsn %s", dsn)
	}

	return &Db{
		db: db,
	}, nil
}

func (db *Db) AutoMigrate(dst ...interface{}) error {
	err := db.db.AutoMigrate(dst...)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred auto migrating the tables")
	}

	return nil
}

func (db *Db) LockAndReturnUnlock(lockId string) (func(), error) {
	err := db.Lock(lockId)
	if err != nil {
		return func() {}, stacktrace.NewError("could not instantiate a new instance for locking id %s. This can be due to another process is holding the distributed lock. Try again later.", lockId)
	}
	return func() {
		_, err := retry(
			unlockRetryAttempts,
			unlockRetryWaitSeconds,
			func() (bool, error) {
				err := db.Unlock(lockId)
				if err != nil {
					logrus.Errorf("Could not release lock for locking id: %s with following error given: %v", lockId, err)
					return false, err
				}
				return true, err
			},
		)
		if err != nil {
			logrus.Errorf("Could not release lock for locking id after retries: %s with following error given: %v", lockId, err)
		}
	}, nil
}

func (db *Db) Lock(lockId string) error {
	if len(lockId) < 1 {
		return stacktrace.NewError("database lock id must not be empty")
	}
	hashedId := hash(lockId)
	var lock bool
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultLockTimeout)
	defer cancel()
	result := db.db.WithContext(ctxWithTimeout).Raw(fmt.Sprintf("SELECT pg_try_advisory_lock(%d)", hashedId)).Scan(&lock)
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An error occurred trying to retrieve the advisory lock from postgres")
	}
	if !lock {
		return stacktrace.NewError("database lock is already taken for id %s (hashed as %d)", lockId, hashedId)
	}
	logrus.Infof("Obtained database lock for id %s (hashed as %d)", lockId, hashedId)
	return nil
}

func (db *Db) Unlock(lockId string) error {
	if len(lockId) < 1 {
		return stacktrace.NewError("database lock id must not be empty")
	}
	hashedId := hash(lockId)
	var lock bool
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultLockTimeout)
	defer cancel()
	result := db.db.WithContext(ctxWithTimeout).Raw(fmt.Sprintf("SELECT pg_advisory_unlock(%d)", hashedId)).Scan(&lock)
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An error occurred retrieving the advisory lock from postgres")
	}
	if !lock {
		return stacktrace.NewError("unable to release database lock")
	}
	logrus.Infof("Released database lock for id %s (hashed as %d)", lockId, hashedId)
	return nil
}

func (db *Db) Migrate() error {
	unlockFunc, err := db.LockAndReturnUnlock(lockObjectName)
	if err != nil {
		logrus.Infof("DB migration already running")
		return nil
	}
	defer unlockFunc()

	err = db.AutoMigrate(&User{})
	if err != nil {
		return stacktrace.Propagate(err, "An internal error has occurred migrating the tables")
	}

	return nil
}

type DatabaseConnectionInfo struct {
	username     string
	password     string
	host         string
	port         uint16
	databaseName string
}

func NewDatabaseConnectionInfo(
	username string,
	password string,
	host string,
	port uint16,
	databaseName string,
) (*DatabaseConnectionInfo, error) {
	return &DatabaseConnectionInfo{
		username:     username,
		password:     password,
		host:         host,
		port:         port,
		databaseName: databaseName,
	}, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// TODO: Rather than sleeping, replace with timer ticks and background processing (channel) to avoid blocking the thread
func retry[T any](attempts int, sleep int, f func() (T, error)) (result T, err error) {
	hasRetried := false
	for i := 0; i < attempts; i++ {
		if i > 0 {
			logrus.Infof("Retry attempt # %d", i)
		}
		result, err = f()
		if err == nil {
			if hasRetried {
				logrus.Infof("Succeeded call after %d attempts", i+1)
			}
			return result, nil
		}
		logrus.Warnf("command failed with error '%s'", err)
		time.Sleep(time.Duration(sleep) * time.Second)
		// sleep *= 2 // enable for exponential backoff
		hasRetried = true
	}
	return result, fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}