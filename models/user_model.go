package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UserType string

const (
	System  UserType = "system"
	Manager UserType = "manager"
	Agent   UserType = "agent"
)

func (UserType) GormDataType() string {
	return "user_type"
}

func (UserType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	if db.Dialector.Name() == "postgres" {
		return "user_type"
	}
	return "varchar(10)"
}

type User struct {
	gorm.Model
	Name     string   `gorm:"size:100;not null;index"`
	Account  string   `gorm:"size:100;unique;not null"`
	Password string   `gorm:"size:255;not null"`
	Status   bool     `gorm:"default:true;index"`
	Type     UserType `gorm:"type:user_type;not null;default:'agent';index"`
	TeamID   *uint    `gorm:"index"`
	Team     *Team    `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("Password") {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}

	if u.Type == System {
		if u.TeamID != nil {
			return errors.New("sistem kullanıcısının (system user) bir takımı olamaz (TeamID NULL olmalı)")
		}
	} else if u.Type == Manager || u.Type == Agent {
		if u.TeamID == nil {
			return errors.New("yönetici (manager) veya temsilci (agent) kullanıcısının bir takımı olmalı (TeamID boş olamaz)")
		}
	} else {
		return errors.New("geçersiz kullanıcı tipi (UserType)")
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return errors.New("şifre boş olamaz")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func (u *User) IsManager() bool {
	return u.Type == Manager
}
