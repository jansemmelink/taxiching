package memory

import (
	"sync"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/goods"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
)

func New(users users.IUsers) (goods.IProducts, error) {
	return &factory{
		users:    users,
		byID:     make(map[string]goods.IProduct),
		byUserID: make(map[string]map[string]goods.IProduct),
	}, nil
}

type factory struct {
	mutex    sync.Mutex
	users    users.IUsers
	byID     map[string]goods.IProduct
	byUserID map[string]map[string]goods.IProduct
}

func (f *factory) New(userID string, goodsName string, cost wallets.Amount) (goods.IProduct, error) {
	if f == nil {
		panic("nil.New()")
	}

	if len(goodsName) < 1 {
		return nil, log.Wrapf(nil, "goods.name is required")
	}
	if cost <= 0 {
		return nil, log.Wrapf(nil, "goods.cost=%d must be >0", cost)
	}
	u := f.users.GetID(userID)
	if u == nil {
		return nil, log.Wrapf(nil, "unknown user id")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.byUserID[userID]; ok {
		if _, ok := f.byUserID[userID][goodsName]; ok {
			return nil, log.Wrapf(nil, "user already has goods.name=%s", goodsName)
		}
	}

	g := &memoryGoods{
		id:   uuid.NewV1().String(),
		user: u,
		name: goodsName,
		cost: cost,
	}

	if _, ok := f.byID[g.id]; ok {
		return nil, log.Wrapf(nil, "duplicate goods.id=%s created", g.id)
	}
	if _, ok := f.byUserID[g.id]; !ok {
		f.byUserID[g.id] = make(map[string]goods.IProduct)
	}
	f.byUserID[g.id][goodsName] = g
	f.byID[g.id] = g
	return g, nil
} //factory.New()

func (f *factory) UserGoods(userID string) (map[string]goods.IProduct, bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	ug, ok := f.byUserID[userID]
	return ug, ok
} //factory.UserGoods()

func (f *factory) GetID(goodsID string) goods.IProduct {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.byID[goodsID]
} //factory.GetID()

func (f *factory) DelID(goodsID string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if g, ok := f.byID[goodsID]; ok {
		if ug, ok := f.byUserID[g.Owner().ID()]; ok {
			if _, ok := ug[g.Name()]; ok {
				delete(f.byUserID[g.Owner().ID()], g.Name())
				log.Debugf("Deleted goods.name=%s for user.id=%s", g.Name(), g.Owner().ID())
			} else {
				log.Debugf("Delete goods.id=%s user.%s[name=%s] not found", goodsID, g.Owner().ID(), g.Name())
			}
		} else {
			log.Debugf("Delete goods.id=%s user.id=%s list not found", goodsID, g.Owner().ID())
		}
		delete(f.byID, goodsID)
		log.Debugf("Deleted goods.id=%s", goodsID)
	} else {
		log.Debugf("Delete goods.id=%s not found", goodsID)
	}
} //factory.DelID()

//memoryGoods implements IProduct
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
