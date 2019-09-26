package goods

import (
	"sync"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
)

type IGoods interface {
	ID() string
	Owner() users.IUser //=owner of the goods, who gets the money when bought
	Name() string       //name of goods
	Cost() wallets.Amount
}

type IGoodsFactory interface {
	New(u users.IUser, name string, cost wallets.Amount) (IGoods, error)
}

func Register(f IGoodsFactory) {
	goodsMutex.Lock()
	defer goodsMutex.Unlock()
	if factory != nil {
		panic(log.Wrapf(nil, "Multiple goods factories registered. First was %T", factory))
	}
	factory = f
}

var (
	goodsMutex    sync.Mutex
	factory       IGoodsFactory
	goodsByID     = make(map[string]IGoods)
	goodsByUserID = make(map[string]map[string]IGoods)
)

func New(userID string, goodsName string, cost wallets.Amount) (IGoods, error) {
	if len(goodsName) < 1 {
		return nil, log.Wrapf(nil, "goods.name is required")
	}
	if cost <= 0 {
		return nil, log.Wrapf(nil, "goods.cost=%d must be >0", cost)
	}

	u := users.GetByID(userID)
	if u == nil {
		return nil, log.Wrapf(nil, "unknown user id")
	}

	goodsMutex.Lock()
	defer goodsMutex.Unlock()
	if _, ok := goodsByUserID[userID]; ok {
		if _, ok := goodsByUserID[userID][goodsName]; ok {
			return nil, log.Wrapf(nil, "user already has goods.name=%s", goodsName)
		}
	}
	if factory == nil {
		return nil, log.Wrapf(nil, "no goods factory registered")
	}

	newGoods, err := factory.New(u, goodsName, cost)
	if err != nil {
		return nil, log.Wrapf(err, "failed to create goods")
	}
	if _, ok := goodsByID[newGoods.ID()]; ok {
		return nil, log.Wrapf(err, "duplicate goods.id=%s created by %T", newGoods.ID(), factory)
	}
	if newGoods.Owner().ID() != userID {
		return nil, log.Wrapf(err, "new goods created with wrong user.id by %T", factory)
	}

	if _, ok := goodsByUserID[newGoods.Owner().ID()]; !ok {
		goodsByUserID[newGoods.Owner().ID()] = make(map[string]IGoods)
	}
	goodsByUserID[newGoods.Owner().ID()][goodsName] = newGoods
	goodsByID[newGoods.ID()] = newGoods
	return newGoods, nil
}

func UserGoods(userID string) (map[string]IGoods, bool) {
	goodsMutex.Lock()
	defer goodsMutex.Unlock()
	ug, ok := goodsByUserID[userID]
	return ug, ok
} //UserGoods()

func Get(goodsID string) IGoods {
	goodsMutex.Lock()
	defer goodsMutex.Unlock()
	return goodsByID[goodsID]
} //Get()

func Del(goodsID string) {
	goodsMutex.Lock()
	defer goodsMutex.Unlock()

	if g, ok := goodsByID[goodsID]; ok {
		if ug, ok := goodsByUserID[g.Owner().ID()]; ok {
			if _, ok := ug[g.Name()]; ok {
				delete(goodsByUserID[g.Owner().ID()], g.Name())
				log.Debugf("Deleted goods.name=%s for user.id=%s", g.Name(), g.Owner().ID())
			} else {
				log.Debugf("Delete goods.id=%s user.%s[name=%s] not found", goodsID, g.Owner().ID(), g.Name())
			}
		} else {
			log.Debugf("Delete goods.id=%s user.id=%s list not found", goodsID, g.Owner().ID())
		}
		delete(goodsByID, goodsID)
		log.Debugf("Deleted goods.id=%s", goodsID)
	} else {
		log.Debugf("Delete goods.id=%s not found", goodsID)
	}
} //Del()
