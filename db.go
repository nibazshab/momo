package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

type User struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Password string `json:"password" gorm:"size:32;not null"`
}

type Session struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	UserId    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	User      User
}

type Msg struct {
	ID       uint      `gorm:"primaryKey;auto_increment"`
	ConvId   int       `gorm:"not null"`
	UserId   int       `json:"user_id"`
	UserName string    `json:"user_name" gorm:"not null"`
	Time     time.Time `gorm:"autoCreateTime"`
	FmtTime  string    `json:"time" gorm:"not null"`
	Text     string    `json:"text" gorm:"not null"`
	Type     int       `json:"type" gorm:"not null"`
	User     User
}

type Group struct {
	ID      int    `json:"id" gorm:"primaryKey"`
	OwnerId int    `json:"owner_id"`
	Name    string `json:"name" gorm:"not null"`
	Owner   User   `gorm:"foreignKey:OwnerId"`
	User    []User `gorm:"many2many:group_members;constraint:OnDelete:CASCADE;"`
}

type GroupMember struct {
	GroupId int `json:"group_id"`
	UserId  int `json:"user_id"`
}

type File struct {
	UUID         string `json:"uuid" gorm:"primaryKey"`
	OriginalName string `json:"original_name" gorm:"not null"`
	Type         int    `json:"type" gorm:"not null"`
	Size         int64  `json:"size" gorm:"not null"`
	UserId       int
	User         User
}

const (
	dbUser = "root"
	dbPass = "haosql"
	dbHost = "127.0.0.1"
	dbPort = 3306
	dbName = "test" // "momo"
)

func initDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		TranslateError: true,
	})
	if err != nil {
		panic("[err]" + err.Error())
	}

	err = db.AutoMigrate(
		&User{},
		&Session{},
		&Msg{},
		&Group{},
		&File{},
	)
	if err != nil {
		panic("[err]" + err.Error())
	}
}
