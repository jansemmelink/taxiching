package memory

import (
	"time"

	"github.com/jansemmelink/taxiching/lib/sessions"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/satori/uuid"
)

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
}

type factory struct{}

func (f factory) New(u users.IUser) sessions.ISession {
	s := &memorySession{
		id:    uuid.NewV1().String(),
		start: time.Now(),
		user:  u,
		//expire: time.Now().Add(time.Minute * 5),
		data: make(map[string]interface{}),
	}
	s.Extend()
	return s
}

func init() {
	sessions.Register(factory{})
}
