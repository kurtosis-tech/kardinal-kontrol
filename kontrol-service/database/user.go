package database

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserId string `gorm:"uniqueIndex"`
	Active bool
}

func (db *Db) CreateUser(
	userId string,
	active bool,
) (*User, error) {
	user := &User{
		UserId: userId,
		Active: active,
	}
	result := db.db.Create(user)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred creating the user '%v'", userId)
	}
	logrus.Infof("Success! Stored user %s in database", userId)
	return user, nil
}

func (db *Db) SaveUser(user *User) error {
	result := db.db.Save(user)
	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "An internal error has occurred updating the user '%v'", user.UserId)
	}
	return nil
}

func (db *Db) GetUserByUserID(
	userId string,
) (*User, error) {
	var user User
	result := db.db.Where("user_id = ?", userId).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (db *Db) GetAllUsers() (*[]User, error) {
	var users []User
	result := db.db.Find(&users)
	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "An internal error has occurred getting the list of all users")
	}
	return &users, nil
}
