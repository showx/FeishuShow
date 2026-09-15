package models

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OpenID        string    `gorm:"uniqueIndex;size:128" json:"openId"`
	Name          string    `json:"name"`
	Avatar        string    `json:"avatar"`
	Email         string    `json:"email"`
	TenantKey     string    `json:"tenantKey"`
	AccessToken   string    `json:"-"`
	RefreshToken  string    `json:"-"`
	TokenExpiry   time.Time `json:"-"`
	RefreshExpiry time.Time `json:"-"`
	Demo          bool      `json:"demo"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type MinuteJob struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	EventID   string    `json:"eventId"`
	Title     string    `json:"title"`
	Markdown  string    `gorm:"type:text" json:"markdown"`
	DocURL    string    `json:"docUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type PublicUser struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Email  string `json:"email"`
	Demo   bool   `json:"demo"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID:     u.ID,
		Name:   u.Name,
		Avatar: u.Avatar,
		Email:  u.Email,
		Demo:   u.Demo,
	}
}
