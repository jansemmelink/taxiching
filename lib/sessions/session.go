package sessions

import (
	"sync"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
)

type ISession interface {
	ID() string
	Start() time.Time
	Expire() time.Time
	User() users.IUser
	Set(n string, v interface{})
	Get(n string) (interface{}, bool)
	Extend()
}

type ISessionFactory interface {
	New(u users.IUser) ISession
}

func Register(f ISessionFactory) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	if factory != nil {
		panic(log.Wrapf(nil, "Multiple session factories registered. First was %T", factory))
	}
	factory = f
}

var (
	sessionsMutex sync.Mutex
	factory       ISessionFactory
	sessionByID   = make(map[string]ISession)
)

func New(u string, p string) (ISession, error) {
	user := users.GetByName(u)
	if user == nil {
		return nil, log.Wrapf(nil, "unknown user")
	}
	if !user.Auth(p) {
		return nil, log.Wrapf(nil, "incorrect password")
	}

	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	if factory == nil {
		panic(log.Wrapf(nil, "no session factory registered"))
	}

	newSession := factory.New(user)
	if _, ok := sessionByID[newSession.ID()]; ok {
		panic(log.Wrapf(nil, "session factory created duplicate session id=\"%s\"", newSession.ID()))
	}
	if newSession.User() != user {
		panic(log.Wrapf(nil, "session factory created session with unexpected user"))
	}
	sessionByID[newSession.ID()] = newSession
	return newSession, nil
} //New()

func IsValid(s ISession) bool {
	if s != nil {
		if existing, ok := sessionByID[s.ID()]; ok {
			if existing.ID() == s.ID() {
				if existing.Expire().After(time.Now()) {
					return true
				}
			}
		}
	}
	return false
}
