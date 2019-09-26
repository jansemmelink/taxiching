package main

import (
	"testing"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/ledger"
	"github.com/jansemmelink/taxiching/lib/sessions"
	_ "github.com/jansemmelink/taxiching/lib/sessions/memory"
	"github.com/jansemmelink/taxiching/lib/users"
	_ "github.com/jansemmelink/taxiching/lib/users/memory"
	"github.com/jansemmelink/taxiching/lib/wallets"
	_ "github.com/jansemmelink/taxiching/lib/wallets/memory"
)

func TestThings(t *testing.T) {
	log.DebugOn()

	//create wallet to represent bank account for EFT deposits
	adminUser, err := users.New("27824526299", "admin", "admin")
	if err != nil {
		panic(log.Wrapf(err, "Failed to create admin user"))
	}
	bankWallet, err := wallets.New(adminUser.ID(), "bank", -100000000)
	if err != nil {
		panic(log.Wrapf(err, "Failed to create bank wallet"))
	}

	//create wallet for all unknown deposits
	unknownUserWallet, err := wallets.New(adminUser.ID(), "unknown-user", 0)
	if err != nil {
		t.Fatalf("Failed to create unknown user wallet for id=%s: %v", adminUser.ID(), err)
	}

	adminSession, err := sessions.New(adminUser.ID(), "admin")
	if err != nil {
		panic(log.Wrapf(err, "failed to create admin session"))
	}

	user1, wallet1 := newUser("27111111111", "one", "aa11")
	_ /*user2*/, wallet2 := newUser("27222222222", "two", "bb22")

	//EFT deposits into our bank account with that wallet's deposit reference
	//are loaded from the bank statement into the user wallet
	//this is sent from admin account, so needs admin session to do this
	deposits := []struct {
		ref    string
		amount wallets.Amount
	}{
		{wallet2.DepositReference(), 10000},
		{"Joe Smith", 50000},
		{wallet1.DepositReference(), 12000},
		{"W-3AAA-5BBB", 6000},
	}
	for _, dep := range deposits {
		depWallet := wallets.GetByDepRef(dep.ref)
		if depWallet == nil {
			depWallet = unknownUserWallet
		}
		if _, err := ledger.Send(
			adminSession,
			bankWallet,
			depWallet,
			dep.amount,
			"EFT deposit"); err != nil {
			panic(log.Wrapf(err, "Failed to load deposit"))
		}
	} //for each deposit

	sessionUser1, err := sessions.New(user1.ID(), "aa11")
	if err != nil {
		panic(log.Wrapf(err, "failed to login as one"))
	}

	//user can transfer from own wallet to other user's wallet's deposit reference
	if _, err := ledger.Send(
		sessionUser1,
		wallet1,
		wallet2,
		50,
		"xyz"); err != nil {
		panic("Failed to send money")
	}

	//products are created per user

	//show status
	log.Debugf("Wallets:")
	for _, w := range wallets.All() {
		log.Debugf("%s %s %s %d", w.ID(), w.Owner().Name(), w.Name(), w.Balance())
	}

	//show transactions
	log.Debugf("Transactions:")
	for _, t := range ledger.All() {
		log.Debugf("%s %s %s %d %s %s", t.Timestamp(),
			t.DebitWallet().Name(),
			t.CreditWallet().Name(),
			t.Amount(),
			t.Description(),
			t.Reference())
	}
}

func newUser(m, n, p string) (users.IUser, wallets.IWallet) {
	//normal user session to register user and create new user-wallet
	//s := sessions.New()
	//log.Debugf("s:{id:%s}", s.ID())

	u, err := users.New(m, n, p)
	if err != nil {
		panic(log.Wrapf(err, "failed to create user(%s)", n))
	}
	w, err := wallets.New(u.ID(), "default", 0)
	if err != nil {
		panic(log.Wrapf(err, "failed to create user default wallet"))
	}

	// if user1.Auth("aaa") {
	// 	panic("User auth success on wrong password")
	// }
	// if !user1.Auth("defg") {
	// 	panic("User auth fail on correct password")
	// }
	// s.Set("user", user1.ID())
	// s.Set("wallet", w.ID())

	log.Debugf("user:{id:%s,name:%s}", u.ID(), u.Name())
	log.Debugf("wallet:{id:%s,user:%s}", w.ID(), w.Owner().ID())
	return u, w
}
