package memory

import (
	"sync"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/sessions"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/satori/uuid"
)

//new memory pool of sessions
func New(users users.IUsers) (sessions.ISessions, error) {
	return &factory{
		users: users,
		byID:  make(map[string]sessions.ISession),
	}, nil
} //New()

type factory struct {
	users users.IUsers
	mutex sync.Mutex
	byID  map[string]sessions.ISession
}

func (f *factory) New(userID string, password string) (sessions.ISession, error) {
	if f == nil {
		panic(log.Wrapf(nil, "nil.New()"))
	}

	log.Debugf("Creating session for %s %s", userID, password)
	user := f.users.GetID(userID)
	if user == nil {
		return nil, log.Wrapf(nil, "unknown user")
	}
	if !user.Auth(password) {
		return nil, log.Wrapf(nil, "incorrect password")
	}

	newSession := &memorySession{
		id:    uuid.NewV1().String(),
		start: time.Now(),
		user:  user,
		data:  make(map[string]interface{}),
	}
	newSession.Extend()

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.byID[newSession.id]; ok {
		panic(log.Wrapf(nil, "session factory created duplicate session id=\"%s\"", newSession.ID()))
	}

	//end and remove other sessions for the same user
	for id, s := range f.byID {
		if s.User().ID() == user.ID() {
			delete(f.byID, id)
			log.Debugf("  ENDING OTHER session.id=%s for user.id=%s started at %v", id, user.ID(), s.Start())
			continue
		}
		if s.Expire().Before(time.Now()) {
			delete(f.byID, id)
			log.Debugf("  ENDING EXPIRED session.id=%s for user.id=%s started at %v", id, user.ID(), s.Start())
			continue
		}
	}

	//cleared old entries, now add this one
	f.byID[newSession.id] = newSession

	{
		s := newSession
		log.Debugf("SESSION START: {id:%s, start:%s, dur:%v, user:%s}", s.ID(), s.Start(), time.Now().Sub(s.Start()), s.User().ID())
		f.logSessionList("started session.id=" + s.ID())
	}

	return newSession, nil
} //f.New()

func (f *factory) GetID(id string) sessions.ISession {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	s, ok := f.byID[id]
	if !ok {
		f.logSessionList("session not found with id=" + id)
		return nil
	}
	if s.Expire().Before(time.Now()) {
		log.Debugf("Session.id=%s expired at %v", id, s.Expire())
		delete(f.byID, id)
		return nil
	}
	//automatically extend the session while being used
	s.Extend()
	return s
} //GetID()

func (f *factory) End(id string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if s, ok := f.byID[id]; ok {
		log.Debugf("SESSION ENDED: {id:%s, start:%s, dur:%v, user:%s}", s.ID(), s.Start(), time.Now().Sub(s.Start()), s.User().ID())
		delete(f.byID, id)
		f.logSessionList("ended session.id=" + s.ID())
	}
}

func (f *factory) IsValid(s sessions.ISession) bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if s != nil {
		if existing, ok := f.byID[s.ID()]; ok {
			if existing.ID() == s.ID() {
				if existing.Expire().After(time.Now()) {
					return true
				}
			}
		}
	}
	return false
} //factory.IsValid()

func (f *factory) logSessionList(title string) {
	log.Debugf("===== %s =====", title)
	for id, s := range f.byID {
		log.Debugf("  %s: user.name=%s exp=%s", id, s.User().Name(), s.Expire())
	}
	log.Debugf("===== end =====")
}

//memorySession implements ISession
type memorySession struct {
	id     string
	start  time.Time
	expire time.Time
	user   users.IUser
	data   map[string]interface{}
}

func (s memorySession) ID() string {
	return s.id
}

func (s memorySession) Start() time.Time {
	return s.start
}

func (s memorySession) Expire() time.Time {
	return s.expire
}

func (s memorySession) User() users.IUser {
	return s.user
}

func (s *memorySession) Set(n string, v interface{}) {
	if s != nil {
		s.data[n] = v
		s.Extend()
	}
}

func (s *memorySession) Get(n string) (interface{}, bool) {
	if s != nil {
		v, ok := s.data[n]
		s.Extend()
		return v, ok
	}
	return nil, false
}

func (s *memorySession) Extend() {
	s.expire = time.Now().Add(time.Minute * 5)
	log.Debugf("+ %s: user.name=%s exp=%s", s.id, s.User().Name(), s.Expire())
}
