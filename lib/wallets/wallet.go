package wallets

import (
	"github.com/jansemmelink/taxiching/lib/users"
)

type IWallets interface {
	New(u users.IUser, name string, minBalance Amount) (IWallet, error)
	NewDepRef(w IWallet) (string, error)
	GetByDepRef(ref string) IWallet
	UserWallet(userID string, walletName string) IWallet
}

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
