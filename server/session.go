package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/ledger"
	"github.com/jansemmelink/taxiching/lib/wallets"
)

type sessionData struct {
	SID     string            `json:"session-id,omitempty"`
	UID     string            `json:"user-id,omitempty"`
	Expiry  string            `json:"expiry,omitempty"`
	Balance int               `json:"balance,omitempty"`
	Recent  []transactionData `json:"recent"`
	Goods   []goodsData       `json:"goods"`
}

type transactionData struct {
	Error string `json:"error,omitempty"`

	ID          string         `json:"id,omitempty"`
	Time        string         `json:"time,omitempty"`
	Description string         `json:"description,omitempty"`
	Reference   string         `json:"reference,omitempty"`
	Amount      wallets.Amount `json:"amount,omitempty"`

	NewBalance wallets.Amount `json:"newBalance,omitempty"`
}

type goodsData struct {
	ID   string
	Name string
	Cost wallets.Amount
}

//r.Get("/session/{id}/ministatement", SessionMiniStatement)
func SessionMiniStatement(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}
	w := bank.Wallets.UserWallet(s.User().ID(), "default")
	if w == nil {
		http.Error(res, "Failed to get user wallet", http.StatusInternalServerError)
		return
	}

	//output
	sd := sessionData{
		SID:     s.ID(),
		Expiry:  s.Expire().Format(timeFormat),
		UID:     s.User().ID(),
		Balance: int(w.Balance()),
		Recent:  make([]transactionData, 0),
	}

	transactions := ledger.All()
	for _, t := range transactions {
		if t.DebitWallet().ID() == w.ID() || t.CreditWallet().ID() == w.ID() {
			sd.Recent = append(sd.Recent, transactionData{
				ID:          t.ID(),
				Time:        t.Timestamp().Format(timeFormat),
				Description: t.Description(),
				Reference:   t.Reference(),
				Amount:      t.Amount(),
			})
		}
	}

	j, _ := json.Marshal(sd)
	res.Write(j)
}

//r.Delete("/session/{id}/goods/{goodsid}", SessionGoodsDel)
func SessionGoodsDel(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}

	goodsID := req.URL.Query().Get(":goodsid")
	bank.Goods.DelID(goodsID)
}

//r.Get("/session/{id}/goods", SessionGoodsList)
func SessionGoodsList(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}

	//output
	sd := sessionData{
		SID:    s.ID(),
		Expiry: s.Expire().Format(timeFormat),
		UID:    s.User().ID(),
		Goods:  make([]goodsData, 0),
	}

	if ug, ok := bank.Goods.UserGoods(s.User().ID()); ok {
		for _, g := range ug {
			sd.Goods = append(sd.Goods, goodsData{
				ID:   g.ID(),
				Name: g.Name(),
				Cost: g.Cost(),
			})
		}
	}

	j, _ := json.Marshal(sd)
	res.Write(j)
} //SessionGoodsList()

//r.Post("/session/{id}/goods", SessionGoodsAdd)
func SessionGoodsAdd(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}

	//parse request
	var gd goodsData
	if err := json.NewDecoder(req.Body).Decode(&gd); err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(gd.ID) != 0 {
		http.Error(res, "id not allowed", http.StatusBadRequest)
		return
	}
	if len(gd.Name) == 0 {
		http.Error(res, "missing name", http.StatusBadRequest)
		return
	}
	if gd.Cost <= 0 {
		http.Error(res, fmt.Sprintf("invalid cost=%d", gd.Cost), http.StatusBadRequest)
		return
	}

	g, err := bank.Goods.New(s.User().ID(), gd.Name, gd.Cost)
	if err != nil {
		http.Error(res, fmt.Sprintf("failed to create goods: %v", err), http.StatusBadRequest)
		return
	}

	gd = goodsData{
		ID:   g.ID(),
		Name: g.Name(),
		Cost: g.Cost(),
	}
	j, _ := json.Marshal(gd)
	res.Write(j)
}

//r.Get("/session/{id}/keepalive", SessionKeepAlive)
func SessionKeepAlive(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}

	s.Extend()

	//output
	sd := sessionData{
		SID:    s.ID(),
		Expiry: s.Expire().Format(timeFormat),
		UID:    s.User().ID(),
		//Balance: int(w.Balance()),
		//Recent: make([]transactionData, 0),
	}
	j, _ := json.Marshal(sd)
	res.Write(j)
}

//r.Post("/session/{id}/pay/goods/{goodsid}", SessionPayGoods)
func SessionPayGoods(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}
	goodsID := req.URL.Query().Get(":goodsid")
	g := bank.Goods.GetID(goodsID)
	if g == nil {
		http.Error(res, "Unknown goods id", http.StatusNotFound)
		return
	}
	buyerWallet := bank.Wallets.UserWallet(s.User().ID(), "default")
	if buyerWallet == nil {
		http.Error(res, "Failed to get user wallet", http.StatusInternalServerError)
		return
	}
	td := transactionData{}
	if buyerWallet.Balance() < g.Cost() {
		http.Error(res, "Insufficient funds", http.StatusNotAcceptable)
		return
	}

	//wallet of seller
	seller := g.Owner()
	sellerWallet := bank.Wallets.UserWallet(seller.ID(), "default")
	if sellerWallet == nil {
		http.Error(res, "Failed to get seller wallet", http.StatusInternalServerError)
		return
	}
	if buyerWallet.ID() == sellerWallet.ID() {
		http.Error(res, "Cannot buy your own goods", http.StatusBadRequest)
		return
	}

	ref := fmt.Sprintf("%s buy %s", s.User().Name(), g.Name())
	t, err := bank.Send(s, buyerWallet, sellerWallet, g.Cost(), ref)
	if err != nil {
		http.Error(res, "Failed to transact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	td.ID = t.ID()
	td.Time = t.Timestamp().Format(timeFormat)
	td.Description = t.Reference()
	td.Amount = t.Amount()
	td.NewBalance = buyerWallet.Balance()

	j, _ := json.Marshal(td)
	log.Debugf("transactionData: %s", string(j))
	res.Write(j)
}

type depositRequest struct {
	Msisdn string         `json:"msisdn"`
	Amount wallets.Amount `json:"amount"`
}

//r.Post("/session/{id}/deposit", SessionDeposit)
func SessionDeposit(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	s := bank.Sessions.GetID(sessionID)
	if s == nil {
		http.Error(res, "Unknown session", http.StatusUnauthorized)
		return
	}
	if s.User().Msisdn() != "27824526299" {
		http.Error(res, "This function is restricted to admin user", http.StatusUnauthorized)
		return
	}

	var r depositRequest
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(r.Msisdn) == 0 {
		http.Error(res, "msisdn not specified", http.StatusBadRequest)
		return
	}
	if r.Amount <= 0 {
		http.Error(res, "amount not specified", http.StatusBadRequest)
		return
	}

	u := bank.Users.GetMsisdn(r.Msisdn)
	if u == nil {
		http.Error(res, "Unknown user msisdn="+r.Msisdn, http.StatusNotFound)
		return
	}
	userWallet := bank.Wallets.UserWallet(u.ID(), "default")
	if userWallet == nil {
		http.Error(res, "Failed to get user wallet", http.StatusInternalServerError)
		return
	}

	td := transactionData{}
	ref := fmt.Sprintf("deposit into %s", r.Msisdn)
	t, err := bank.Send(s, bank.BankWallet, userWallet, r.Amount, ref)
	if err != nil {
		http.Error(res, "Failed to transact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	td.ID = t.ID()
	td.Time = t.Timestamp().Format(timeFormat)
	td.Description = t.Reference()
	td.Amount = t.Amount()
	td.NewBalance = userWallet.Balance()

	j, _ := json.Marshal(td)
	log.Debugf("transactionData: %s", string(j))
	res.Write(j)

	return
} //SessionDeposit()

//r.Get("/session/{id}/logout", SessionLogout)
func SessionLogout(res http.ResponseWriter, req *http.Request, bank *ledger.Bank) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	sessionID := req.URL.Query().Get(":id")
	bank.Sessions.End(sessionID)
	return
}
