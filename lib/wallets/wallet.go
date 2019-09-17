package wallets

import (
	"math/rand"
	"sync"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
)

type Amount int

type IWallet interface {
	ID() string
	Owner() users.IUser //=owner of the wallet, who may send from wallet
	Name() string
	Balance() Amount
	MinBalance() Amount //default to 0, negative for account that may go negative
	DepositReference() string

	//todo: need strict access to this, may be implement outside wallet e.g. in ledger
	//and required logged in session to indicate who does it
	//and apply access control to operations
	Debit(amount Amount)
	Credit(amount Amount)
}

type IWalletFactory interface {
	New(u users.IUser, name string, minBalance Amount) (IWallet, error)
}

func Register(f IWalletFactory) {
	walletsMutex.Lock()
	defer walletsMutex.Unlock()
	if factory != nil {
		panic(log.Wrapf(nil, "Multiple wallet factories registered. First was %T", factory))
	}
	factory = f
}

var (
	walletsMutex   sync.Mutex
	factory        IWalletFactory
	walletByID     = make(map[string]IWallet)
	walletByUserID = make(map[string]map[string]IWallet)
)

func New(userID string, walletName string, minBalance Amount) (IWallet, error) {
	if len(walletName) < 1 {
		return nil, log.Wrapf(nil, "missing wallet name")
	}

	u := users.GetByID(userID)
	if u == nil {
		return nil, log.Wrapf(nil, "unknown user id")
	}

	walletsMutex.Lock()
	defer walletsMutex.Unlock()
	if _, ok := walletByUserID[userID]; ok {
		if _, ok := walletByUserID[userID][walletName]; ok {
			return nil, log.Wrapf(nil, "user already has a wallet named %s", walletName)
		}
	}
	if factory == nil {
		return nil, log.Wrapf(nil, "no wallet factory registered")
	}

	//make short unique deposit reference
	newWallet, err := factory.New(u, walletName, minBalance)
	if err != nil {
		return nil, log.Wrapf(err, "failed to create wallet")
	}
	if _, ok := walletByID[newWallet.ID()]; ok {
		return nil, log.Wrapf(err, "duplicate wallet.id=%s created by %T", newWallet.ID(), factory)
	}
	if newWallet.Owner().ID() != userID {
		return nil, log.Wrapf(err, "new wallet create with wrong user id by %T", factory)
	}
	walletByID[newWallet.ID()] = newWallet

	if _, ok := walletByUserID[newWallet.Owner().ID()]; !ok {
		walletByUserID[newWallet.Owner().ID()] = make(map[string]IWallet)
	}
	walletByUserID[newWallet.Owner().ID()][walletName] = newWallet
	return newWallet, nil
}

var (
	depRefMutex    sync.Mutex
	walletByDepRef = make(map[string]IWallet)
)

//wallet id is considered secret, and its long and tedious to type
//this function allocates a unique deposit reference for a wallet
//that is easier to type
//to make it simple, the range is limited, and it should only be
//used when user can make EFT payments into the wallet
func NewDepRef(w IWallet) (string, error) {
	depRefMutex.Lock()
	defer depRefMutex.Unlock()
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
		if _, ok := walletByDepRef[ref]; !ok {
			//found unique dep ref id
			walletByDepRef[ref] = w
			return ref, nil
		} //if found
	} //for each attempt
	return "", log.Wrapf(nil, "Unable to generate deposit reference")
}

func GetByDepRef(ref string) IWallet {
	depRefMutex.Lock()
	defer depRefMutex.Unlock()
	if w, ok := walletByDepRef[ref]; ok {
		return w
	}
	return nil
}

func All() map[string]IWallet {
	return walletByID
}
