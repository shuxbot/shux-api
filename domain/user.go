package domain

import "time"

type (
	UserRepository interface {
		Get(UserId string, ServerId string) (User, error)
		Delete(UserId string, ServerId string) error
		Update(u *User, ServerId string) error
		Replace(u *User, ServerId string) error
		Create(u *User, ServerId string) error
	}
	Warning struct {
		Reason string    `json:"reason"`
		Date   time.Time `json:"date"`
	}
	User struct {
		UserId         string    `json:"id,omitempty"`
		Description    string    `json:"description,omitempty"`
		Points         float64   `json:"points,omitempty"`
		Warnings       int64     `json:"warnings,omitempty"`
		WarningsRecord []Warning `json:"warnings_record,omitempty"`
		Tickets        int64     `json:"tickets,omitempty"`
		Flags          int64     `json:"flags,omitempty"`
	}
)
