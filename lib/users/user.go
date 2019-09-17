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
	Name() string
	Auth(password string) bool
	SetPassword(oldPassword, newPassword string) error
	//Profile() image
}

type IUserFactory interface {
	New(username, password string) (IUser, error)
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
	namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9@_\.-]*[a-zA-Z0-9]$`)

	usersMutex sync.Mutex
	factory    IUserFactory
	userByName = make(map[string]IUser)
	userByID   = make(map[string]IUser)
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

func GetByName(n string) IUser {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	if u, ok := userByName[n]; ok {
		return u
	}
	return nil
}

func New(username, password string) (IUser, error) {
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
	if _, ok := userByName[n]; ok {
		return nil, log.Wrapf(nil, "user.name=\"%s\" already exists.", n)
	}
	if factory == nil {
		return nil, log.Wrapf(nil, "no user factory registered")
	}
	newUser, err := factory.New(n, password)
	if err != nil {
		return nil, log.Wrapf(nil, "failed to create new user")
	}
	if _, ok := userByID[newUser.ID()]; ok {
		return nil, log.Wrapf(nil, "duplicate user.id=\"%s\"", newUser.ID())
	}

	userByID[newUser.ID()] = newUser
	userByName[newUser.Name()] = newUser
	return newUser, nil
} //New()
