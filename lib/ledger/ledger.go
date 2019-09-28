package ledger

import (
	"sync"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/sessions"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
)

//transact creates a new transaction
func (b Bank) transact(dt, ct wallets.IWallet, txTime time.Time, amount wallets.Amount, desc, ref string) (ITransaction, error) {
	if txTime.After(time.Now()) {
		return nil, log.Wrapf(nil, "future transaction time = %v", txTime)
	}
	if dt == nil {
		return nil, log.Wrapf(nil, "debit wallet not specified")
	}
	if ct == nil {
		return nil, log.Wrapf(nil, "credit wallet not specified")
	}
	if dt == ct {
		return nil, log.Wrapf(nil, "debit and credit wallet are the same")
	}
	if amount <= 0 {
		return nil, log.Wrapf(nil, "amount not positive")
	}
	if len(desc) == 0 || len(ref) == 0 {
		return nil, log.Wrapf(nil, "desc and ref are required")
	}

	//only one transaction at a time
	mutex.Lock()
	defer mutex.Unlock()

	//define the transaction and append
	dt.Debit(amount)
	ct.Credit(amount)
	t := transaction{
		id: uuid.NewV1().String(),
		ts: time.Now(),

		//details:
		timestamp:      time.Now(),
		dtWallet:       dt,
		dtBalanceAfter: 0,
		ctWallet:       ct,
		ctBalanceAfter: 0,
		amount:         amount,
		description:    desc,
		reference:      ref,
	}
	transactions = append(transactions, t)
	return t, nil
} //Transact()

var (
	mutex        sync.Mutex
	transactions = make([]ITransaction, 0)
)

type ITransaction interface {
	ID() string
	Timestamp() time.Time
	DebitWallet() wallets.IWallet
	CreditWallet() wallets.IWallet
	Amount() wallets.Amount
	Description() string
	Reference() string
}

type transaction struct {
	id string
	ts time.Time //time created

	dtWallet       wallets.IWallet
	dtBalanceAfter wallets.Amount

	ctWallet       wallets.IWallet
	ctBalanceAfter wallets.Amount

	timestamp   time.Time
	amount      wallets.Amount
	description string
	reference   string
}

func (t transaction) ID() string                    { return t.id }
func (t transaction) Timestamp() time.Time          { return t.timestamp }
func (t transaction) DebitWallet() wallets.IWallet  { return t.dtWallet }
func (t transaction) CreditWallet() wallets.IWallet { return t.ctWallet }
func (t transaction) Amount() wallets.Amount        { return t.amount }
func (t transaction) Description() string           { return t.description }
func (t transaction) Reference() string             { return t.reference }

func All() []ITransaction {
	return transactions
}

//Send money between wallets
func (b Bank) Send(s sessions.ISession, from wallets.IWallet, to wallets.IWallet, amount wallets.Amount, reference string) (ITransaction, error) {
	if !b.Sessions.IsValid(s) {
		return nil, log.Wrapf(nil, "Invalid session")
	}
	if from == nil {
		return nil, log.Wrapf(nil, "from wallet not specified")
	}
	if to == nil {
		return nil, log.Wrapf(nil, "to wallet not specified")
	}
	if from.ID() == to.ID() {
		return nil, log.Wrapf(nil, "cannot send to same wallet")
	}
	if amount <= 0 {
		return nil, log.Wrapf(nil, "send requires positive amount")
	}
	if len(reference) == 0 {
		return nil, log.Wrapf(nil, "send requires reference")
	}

	//check user permission
	u := s.User()
	log.Debugf("from=%v", from)
	log.Debugf("from.owner=%v", from.Owner())
	log.Debugf("from.owner.id=%v", from.Owner().ID())
	log.Debugf("session.user=%v", u)
	log.Debugf("session.user.id=%v", u.ID())
	if from.Owner().ID() != u.ID() {
		return nil, log.Wrapf(nil, "cannot send from other user's wallet")
	}

	//user wallets may not go negative, but bank account may
	//as we debit it with deposits
	if from.Balance()-amount < from.MinBalance() {
		return nil, log.Wrapf(nil, "insufficient funds (w:{id:%s,bal:%d,min-bal:%d} a:%d)", from.ID(), from.Balance(), from.MinBalance(), amount)
	}

	t, err := b.transact(
		from,
		to,
		time.Now(),
		amount,
		"send",
		reference)
	if err != nil {
		return nil, log.Wrapf(err, "failed to transact")
	}
	return t, nil
} //Send()
