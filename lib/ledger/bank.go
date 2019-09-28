package ledger

import (
	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/goods"
	mongogoods "github.com/jansemmelink/taxiching/lib/goods/mongo"
	"github.com/jansemmelink/taxiching/lib/sessions"
	memorysessions "github.com/jansemmelink/taxiching/lib/sessions/memory"
	"github.com/jansemmelink/taxiching/lib/users"
	mongousers "github.com/jansemmelink/taxiching/lib/users/mongo"
	"github.com/jansemmelink/taxiching/lib/wallets"
	mongowallets "github.com/jansemmelink/taxiching/lib/wallets/mongo"
)

type Bank struct {
	Users     users.IUsers
	adminUser users.IUser

	Wallets    wallets.IWallets
	BankWallet wallets.IWallet

	Goods goods.IProducts

	Sessions sessions.ISessions
}

func New(mongoURI string) *Bank {
	var err error
	b := &Bank{}

	//users database
	b.Users, err = mongousers.Users(mongoURI, "taxiching")
	if err != nil {
		panic(log.Wrapf(err, "failed to create users"))
	}

	b.adminUser = b.Users.GetMsisdn("27824526299")
	if b.adminUser == nil {
		b.adminUser, err = b.Users.New("27824526299", "Jan", "1155")
		if err != nil {
			panic("Failed to create admin user: " + err.Error())
		}
	}

	b.Wallets, err = mongowallets.Wallets(mongoURI, "taxiching", b.Users)
	if err != nil {
		panic(log.Wrapf(err, "failed to create wallets"))
	}

	b.BankWallet, err = b.Wallets.New(b.adminUser, "bank", -10000000)
	if err != nil {
		panic("Failed to create bank wallet: " + err.Error())
	}

	b.Goods, err = mongogoods.Products(mongoURI, "taxiching", b.Users)

	b.Sessions, err = memorysessions.New(b.Users)
	if err != nil {
		panic("Failed to create bank wallet: " + err.Error())
	}

	log.Debugf("Created bank account")
	return b
}
