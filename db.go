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

const (
	dbUser = "root"
	dbPass = "haosql"
	dbHost = "127.0.0.1"
	dbPort = 3306
	dbName = "test" // "momo"
)

var db *gorm.DB

type User struct {
	Id       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name,omitempty" gorm:"not null"`
	Password string `json:"password,omitempty" gorm:"size:64;not null"`
}

type Secret struct {
	Key string `gorm:"not null;size:16"`
}

type Msg struct {
	Id       uint      `gorm:"primaryKey;auto_increment"`
	ConvId   int       `gorm:"not null"`
	UserId   int       `json:"user_id"`
	UserName string    `json:"user_name" gorm:"not null"`
	Time     time.Time `gorm:"autoCreateTime"`
	FmtTime  string    `json:"time" gorm:"not null"`
	Text     string    `json:"text" gorm:"not null"`
	Type     int       `json:"type" gorm:"not null"`
	User     *User     `gorm:"constraint:OnDelete:CASCADE;"`
}

type Group struct {
	Id      int    `json:"id" gorm:"primaryKey"`
	OwnerId int    `json:"owner_id,omitempty"`
	Name    string `json:"name,omitempty" gorm:"not null"`
	Owner   *User  `json:",omitempty" gorm:"foreignKey:OwnerId"`
	User    []User `json:",omitempty" gorm:"many2many:group_members;constraint:OnDelete:CASCADE;"`
}

type GroupMember struct {
	GroupId int `json:"group_id,omitempty"`
	UserId  int `json:"user_id,omitempty"`
}

type File struct {
	Uuid         string `json:"uuid" gorm:"primaryKey;size:36"`
	OriginalName string `json:"original_name" gorm:"not null"`
	Type         int    `json:"type" gorm:"not null"`
	Size         int64  `json:"size" gorm:"not null"`
	UserId       int    `json:"user_id"`
	User         *User  `gorm:"constraint:OnDelete:CASCADE;"`
}

func initDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		//NamingStrategy: schema.NamingStrategy{
		//	TablePrefix: "momo_",
		//},
		// Logger: logger.Default.LogMode(logger.Silent),
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		),
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("[err]", err)
	}

	err = db.AutoMigrate(
		&User{},
		&Secret{},
		&Msg{},
		&Group{},
		&File{},
	)
	if err != nil {
		log.Fatal("[err]", err)
	}
}
