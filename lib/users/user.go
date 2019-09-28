package users

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/jansemmelink/log"
)

type IUser interface {
	ID() string
	Msisdn() string
	Name() string
	Auth(password string) bool
	SetPassword(oldPassword, newPassword string) error
	//Profile() image
}

type IUsers interface {
	New(msisdn, name, password string) (IUser, error)
	GetMsisdn(msisdn string) IUser
	GetID(id string) IUser
}

// func Register(f IUserFactory) {
// 	if factory != nil {
// 		panic(log.Wrapf(nil, "Multiple user factories registered. First was %T", factory))
// 	}
// 	factory = f
// }

var (
	msisdnPattern = regexp.MustCompile(`^27[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]$`)
	namePattern   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9@_\.-]*[a-zA-Z0-9]$`)
	// factory IUserFactory
)

func ValidatePassword(password string) (string, error) {
	p := strings.Trim(password, " ")
	if len(p) < 4 {
		return "", log.Wrapf(nil, "password is shorter than 4 characters")
	}
	for _, c := range p {
		if !unicode.IsPrint(c) {
			return "", log.Wrapf(nil, "password contains invalid character")
		}
	}
	return p, nil
}

func ValidateMsisdn(msisdn string) (string, error) {
	m := strings.Trim(msisdn, " ")
	if !msisdnPattern.MatchString(m) {
		return "", log.Wrapf(nil, "invalid user.msisdn=\"%s\" must be 27+9digits", m)
	}
	return m, nil
}

func ValidateName(name string) (string, error) {
	n := strings.Trim(name, " ")
	if !namePattern.MatchString(n) {
		return "", log.Wrapf(nil, "invalid user.name=\"%s\"", n)
	}
	return n, nil
}
