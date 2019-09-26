package memory

import (
	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/satori/uuid"
)

//memoryUser implements IUser
type memoryUser struct {
	id       string
	msisdn   string
	username string
	password string
}

func (u memoryUser) ID() string {
	return u.id
}

func (u memoryUser) Msisdn() string {
	return u.msisdn
}

func (u memoryUser) Name() string {
	return u.username
}

func (u memoryUser) Auth(password string) bool {
	if u.password == password {
		return true
	}
	return false
}

func (u *memoryUser) SetPassword(oldPassword, newPassword string) error {
	if u.password != oldPassword {
		return log.Wrapf(nil, "Incorrect old password")
	}
	if err := users.ValidatePassword(newPassword); err != nil {
		return log.Wrapf(nil, "Cannot set invalid password")
	}
	u.password = newPassword
	return nil
}

type factory struct{}

func (f factory) New(msisdn, name, password string) (users.IUser, error) {
	u := &memoryUser{
		id:       uuid.NewV1().String(),
		msisdn:   msisdn,
		username: name,
		password: password,
	}
	return u, nil
}

func init() {
	users.Register(factory{})
}
