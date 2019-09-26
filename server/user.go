package main

import (
	"encoding/json"
	"net/http"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/sessions"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
)

type userData struct {
	ID     string `json:"user-id,omitempty"`
	Msisdn string `json:"msisdn"`
	Name   string `json:"name"`
	Pin    string `json:"pin,omitempty"`
}

func UserAdd(res http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	if req.URL.Path != "/user" || req.Method != http.MethodPost {
		http.Error(res, "create user with POST /user", http.StatusBadRequest)
		return
	}

	var u userData
	err := json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		http.Error(res, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(u.Msisdn) == 0 {
		http.Error(res, "invalid request: missing msisdn", http.StatusBadRequest)
		return
	}
	if len(u.Name) == 0 {
		http.Error(res, "invalid request: missing name", http.StatusBadRequest)
		return
	}
	if len(u.Pin) != 4 {
		http.Error(res, "invalid request: pin must be 4 digits", http.StatusBadRequest)
		return
	}
	user, err := users.New(u.Msisdn, u.Name, u.Pin)
	if err != nil {
		http.Error(res, "failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userDefaultWallet, err := wallets.New(user.ID(), "default", 0)
	if err != nil {
		http.Error(res, "failed to create user wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Debugf("Created wallet:{id:%s,name:%s,balance:%v,depref:%s}", userDefaultWallet.ID(), userDefaultWallet.Name(), userDefaultWallet.Balance(), userDefaultWallet.DepositReference())

	u.ID = user.ID()
	u.Pin = "" //not show in response
	j, _ := json.Marshal(u)
	res.Write(j)
} //UserAdd()

func UserGetID(res http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	id := req.URL.Query().Get(":id")
	if len(id) == 0 {
		http.Error(res, "expecting /user/<id>", http.StatusBadRequest)
		return
	}
	user := users.GetByID(id)
	if user == nil {
		http.Error(res, "user "+id+" does not exist", http.StatusNotFound)
		return
	}

	u := userData{
		ID:     user.ID(),
		Msisdn: user.Msisdn(),
		Name:   user.Name(),
	}
	j, _ := json.Marshal(u)
	res.Write(j)
} //UserGetID()

func UserGetMsisdn(res http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	msisdn := req.URL.Query().Get(":msisdn")
	if len(msisdn) == 0 {
		http.Error(res, "expecting /user/msisdn/<msisdn>", http.StatusBadRequest)
		return
	}
	user := users.GetByMsisdn(msisdn)
	if user == nil {
		http.Error(res, "user "+msisdn+" does not exist", http.StatusNotFound)
		return
	}

	u := userData{
		ID:     user.ID(),
		Msisdn: user.Msisdn(),
		Name:   user.Name(),
	}
	j, _ := json.Marshal(u)
	res.Write(j)
} //UserGetMsisdn()

func UserLogin(res http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.URL.Path)
	userID := req.URL.Query().Get(":id")
	pin := req.URL.Query().Get(":pin")
	if len(userID) == 0 || len(pin) == 0 {
		http.Error(res, "Login with /user/<id>/login/<pin>", http.StatusBadRequest)
		return
	}
	session, err := sessions.New(userID, pin)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	s := sessionData{
		SID:    session.ID(),
		UID:    userID,
		Expiry: session.Expire().Format(timeFormat),
	}
	j, _ := json.Marshal(s)
	res.Write(j)
} //UserLogin()
