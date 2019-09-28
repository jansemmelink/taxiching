package goods

import (
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
)

type IProducts interface {
	New(userID string, name string, cost wallets.Amount) (IProduct, error)
	DelID(id string)
	GetID(goodsID string) IProduct
	UserGoods(userID string) (map[string]IProduct, bool)
}

type IProduct interface {
	ID() string
	Owner() users.IUser //=owner of the goods, who gets the money when bought
	Name() string       //name of goods
	Cost() wallets.Amount
}
