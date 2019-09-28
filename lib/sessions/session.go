package sessions

import (
	"time"

	"github.com/jansemmelink/taxiching/lib/users"
)

type ISessions interface {
	New(userID string, password string) (ISession, error)
	GetID(id string) ISession
	IsValid(s ISession) bool
	End(id string)
}

type ISession interface {
	ID() string
	Start() time.Time
	Expire() time.Time
	User() users.IUser
	Set(n string, v interface{})
	Get(n string) (interface{}, bool)
	Extend()
}
