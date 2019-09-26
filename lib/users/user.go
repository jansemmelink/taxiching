package users

import (
	"regexp"
	"strings"
	"sync"
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

type IUserFactory interface {
	New(msisdn, name, password string) (IUser, error)
}

func Register(f IUserFactory) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	if factory != nil {
		panic(log.Wrapf(nil, "Multiple user factories registered. First was %T", factory))
	}
	factory = f
}

var (
	msisdnPattern = regexp.MustCompile(`^27[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]$`)
	namePattern   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9@_\.-]*[a-zA-Z0-9]$`)

	usersMutex   sync.Mutex
	factory      IUserFactory
	userByID     = make(map[string]IUser)
	userByMsisdn = make(map[string]IUser)
)

func ValidatePassword(s string) error {
	if len(s) < 4 {
		return log.Wrapf(nil, "password is shorter than 4 characters")
	}
	for _, c := range s {
		if !unicode.IsPrint(c) {
			return log.Wrapf(nil, "password contains invalid character")
		}
	}
	return nil
}

func GetByID(id string) IUser {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	if u, ok := userByID[id]; ok {
		return u
	}
	return nil
}

func GetByMsisdn(m string) IUser {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	if u, ok := userByMsisdn[m]; ok {
		return u
	}
	log.Debugf("User not found among %d users", len(userByMsisdn))
	for m, u := range userByMsisdn {
		log.Debugf("  %s: %v", m, u)
	}
	return nil
}

func New(msisdn, username, password string) (IUser, error) {
	m := strings.Trim(msisdn, " ")
	if !msisdnPattern.MatchString(m) {
		return nil, log.Wrapf(nil, "invalid user.msisdn=\"%s\" must be 27+9digits", msisdn)
	}
	//name must be defined and valid
	n := strings.Trim(username, " ")
	if !namePattern.MatchString(n) {
		return nil, log.Wrapf(nil, "invalid user.name=\"%s\" must be alpha-numeric characters only", username)
	}
	//password must be defined and valid
	if err := ValidatePassword(password); err != nil {
		return nil, log.Wrapf(err, "invalid user password")
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()
	if _, ok := userByMsisdn[m]; ok {
		return nil, log.Wrapf(nil, "user.msisdn=\"%s\" already exists.", m)
	}
	if factory == nil {
		return nil, log.Wrapf(nil, "no user factory registered")
	}
	newUser, err := factory.New(m, n, password)
	if err != nil {
		return nil, log.Wrapf(nil, "failed to create new user")
	}
	if _, ok := userByID[newUser.ID()]; ok {
		return nil, log.Wrapf(nil, "duplicate user.id=\"%s\"", newUser.ID())
	}

	userByID[newUser.ID()] = newUser
	userByMsisdn[newUser.Msisdn()] = newUser

	{
		log.Debugf("USER CREATED:{id:%s,msisdn:%s,name:%s}",
			newUser.ID(),
			newUser.Msisdn(),
			newUser.Name())
		log.Debugf("Now %d users:", len(userByID))
		for _, u := range userByID {
			log.Debugf("{id:%s,msisdn:%s,name:%s}", u.ID(), u.Msisdn(), u.Name())
		}
	}
	return newUser, nil
} //New()
