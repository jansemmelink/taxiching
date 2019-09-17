package memory

import (
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
)

type memoryWallet struct {
	id         string
	owner      users.IUser
	name       string
	depRef     string
	balance    wallets.Amount
	minBalance wallets.Amount
}

func (w memoryWallet) ID() string {
	return w.id
}

func (w memoryWallet) Owner() users.IUser {
	return w.owner
}

func (w memoryWallet) Name() string {
	return w.name
}

func (w memoryWallet) Balance() wallets.Amount {
	return w.balance
}

func (w memoryWallet) MinBalance() wallets.Amount {
	return w.minBalance
}

func (w *memoryWallet) Debit(amount wallets.Amount) {
	w.balance -= amount
}
func (w *memoryWallet) Credit(amount wallets.Amount) {
	w.balance += amount
}

func (w *memoryWallet) DepositReference() string {
	if w.depRef == "" {
		//not yet assigned, try to get one
		ref, err := wallets.NewDepRef(w)
		if err != nil || ref == "" {
			return ""
		}
		w.depRef = ref
	}
	return w.depRef
}

type factory struct{}

func (f factory) New(u users.IUser, name string, minBalance wallets.Amount) (wallets.IWallet, error) {
	w := &memoryWallet{
		id:         uuid.NewV1().String(),
		owner:      u,
		name:       name,
		depRef:     "",
		balance:    0,
		minBalance: minBalance,
	}
	return w, nil
}

func init() {
	wallets.Register(factory{})
}
