package memory

import (
	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/goods"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
)

//memoryGoods implements IGoods
type memoryGoods struct {
	id   string
	user users.IUser
	name string
	cost wallets.Amount
}

func (s memoryGoods) ID() string {
	return s.id
}

func (s memoryGoods) Owner() users.IUser {
	return s.user
}

func (s memoryGoods) Name() string {
	return s.name
}

func (s memoryGoods) Cost() wallets.Amount {
	return s.cost
}

type factory struct{}

func (f factory) New(u users.IUser, name string, cost wallets.Amount) (goods.IGoods, error) {
	if u == nil {
		return nil, log.Wrapf(nil, "user not specified")
	}
	if len(name) == 0 {
		return nil, log.Wrapf(nil, "missing name")
	}
	if cost <= 0 {
		return nil, log.Wrapf(nil, "cost=%d must be >0", cost)
	}

	g := &memoryGoods{
		id:   uuid.NewV1().String(),
		user: u,
		name: name,
		cost: cost,
	}
	return g, nil
}

func init() {
	goods.Register(factory{})
}
