package memory

import (
	"math/rand"
	"sync"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
)

func New(users users.IUsers) (wallets.IWallets, error) {
	return &factory{
		users:    users,
		byID:     make(map[string]wallets.IWallet),
		byUserID: make(map[string]map[string]wallets.IWallet),
		byDepRef: make(map[string]wallets.IWallet),
	}, nil
}

type factory struct {
	mutex    sync.Mutex
	users    users.IUsers
	byID     map[string]wallets.IWallet
	byUserID map[string]map[string]wallets.IWallet

	depRefMutex sync.Mutex
	byDepRef    map[string]wallets.IWallet
}

func (f *factory) New(u users.IUser, walletName string, minBalance wallets.Amount) (wallets.IWallet, error) {
	if f == nil {
		return nil, log.Wrapf(nil, "no wallet factory registered")
	}
	if len(walletName) < 1 {
		return nil, log.Wrapf(nil, "missing wallet name")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.byUserID[u.ID()]; ok {
		if _, ok := f.byUserID[u.ID()][walletName]; ok {
			return nil, log.Wrapf(nil, "user already has a wallet named %s", walletName)
		}
	}

	//make short unique deposit reference
	w := &memoryWallet{
		id:         uuid.NewV1().String(),
		owner:      u,
		name:       walletName,
		balance:    0,
		minBalance: minBalance,
	}
	if _, ok := f.byID[w.id]; ok {
		return nil, log.Wrapf(nil, "duplicate wallet.id=%s created", w.id)
	}
	f.byID[w.id] = w
	if _, ok := f.byUserID[w.Owner().ID()]; !ok {
		f.byUserID[w.Owner().ID()] = make(map[string]wallets.IWallet)
	}
	f.byUserID[w.Owner().ID()][walletName] = w
	w.depRef, _ = f.NewDepRef(w)
	return w, nil
} //factory.New()

func (f *factory) UserWallet(userID string, walletName string) wallets.IWallet {
	if f == nil {
		return nil
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if userWallets, ok := f.byUserID[userID]; ok {
		if w, ok := userWallets[walletName]; ok {
			return w
		}
	}
	return nil
} //factory.UserWallet()

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
	return w.depRef
}

//wallet id is considered secret, and its long and tedious to type
//this function allocates a unique deposit reference for a wallet
//that is easier to type
//to make it simple, the range is limited, and it should only be
//used when user can make EFT payments into the wallet
func (f *factory) NewDepRef(w wallets.IWallet) (string, error) {
	f.depRefMutex.Lock()
	defer f.depRefMutex.Unlock()
	for attempt := 0; attempt < 10; attempt++ {
		ref := "W-"
		ref += string('0' + rand.Intn(10))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += "-"
		ref += string('0' + rand.Intn(10))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		if _, ok := f.byDepRef[ref]; !ok {
			//found unique dep ref id
			f.byDepRef[ref] = w
			return ref, nil
		} //if found
	} //for each attempt
	return "", log.Wrapf(nil, "Unable to generate deposit reference")
} //factory.NewDepRef()

func (f *factory) GetByDepRef(ref string) wallets.IWallet {
	f.depRefMutex.Lock()
	defer f.depRefMutex.Unlock()
	if w, ok := f.byDepRef[ref]; ok {
		return w
	}
	return nil
}

func (f *factory) All() map[string]wallets.IWallet {
	return f.byID
}
