package memory

import (
	"sync"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/satori/uuid"
)

func Users() (users.IUsers, error) {
	return &factory{
		mutex:    sync.Mutex{},
		byID:     make(map[string]users.IUser),
		byMsisdn: make(map[string]users.IUser),
	}, nil
}

//memoryUser implements IUser
type memoryUser struct {
	id       string
	msisdn   string
	name     string
	password string
}

func (u memoryUser) ID() string {
	return u.id
}

func (u memoryUser) Msisdn() string {
	return u.msisdn
}

func (u memoryUser) Name() string {
	return u.name
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
	p, err := users.ValidatePassword(newPassword)
	if err != nil {
		return log.Wrapf(nil, "Cannot set invalid password")
	}
	u.password = p
	return nil
}

type factory struct {
	mutex    sync.Mutex
	byID     map[string]users.IUser
	byMsisdn map[string]users.IUser
}

func (f *factory) New(msisdn, name, password string) (users.IUser, error) {
	if f == nil {
		return nil, log.Wrapf(nil, "<nil>.New()")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	m, err := users.ValidateMsisdn(msisdn)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid msisdn")
	}
	n, err := users.ValidateName(name)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid name")
	}
	p, err := users.ValidatePassword(password)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid password")
	}

	if _, ok := f.byMsisdn[m]; ok {
		return nil, log.Wrapf(nil, "user.msisdn=\"%s\" already exists.", m)
	}

	u := &memoryUser{
		id:       uuid.NewV1().String(),
		msisdn:   m,
		name:     n,
		password: p,
	}

	if _, ok := f.byID[u.id]; ok {
		return nil, log.Wrapf(nil, "duplicate user.id=\"%s\"", u.id)
	}

	f.byID[u.id] = u
	f.byMsisdn[u.msisdn] = u

	{
		log.Debugf("USER CREATED:{id:%s,msisdn:%s,name:%s}",
			u.id,
			u.msisdn,
			u.name)
		log.Debugf("Now %d users:", len(f.byID))
		for _, user := range f.byID {
			log.Debugf("{id:%s,msisdn:%s,name:%s}", user.ID(), user.Msisdn(), user.Name())
		}
	}
	return u, nil
} //factory.New()

func (f *factory) GetMsisdn(m string) users.IUser {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if u, ok := f.byMsisdn[m]; ok {
		return u
	}
	log.Debugf("User not found among %d users", len(f.byMsisdn))
	for m, u := range f.byMsisdn {
		log.Debugf("  %s: %v", m, u)
	}
	return nil
} //factory.GetMsisdn()

func (f *factory) GetID(id string) users.IUser {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if u, ok := f.byID[id]; ok {
		return u
	}
	return nil
} //factory.GetID()
