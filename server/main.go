package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/pat"
	"github.com/jansemmelink/log"

	"github.com/jansemmelink/taxiching/lib/ledger"
)

const timeFormat = "2006-01-02T15:04:05+07:00"

func main() {
	debugFlag := flag.Bool("debug", false, "DEBUG Mode")
	addrFlag := flag.String("addr", "localhost:8080", "HTTP Server address")
	flag.Parse()
	if *debugFlag {
		log.DebugOn()
		log.Debugf("DEBUG Mode")
	}

	bank := ledger.New()

	fmt.Fprintf(os.Stdout, "Serving %s\n", *addrFlag)
	if err := http.ListenAndServe(*addrFlag, router(bank)); err != nil {
		panic(log.Wrapf(err, "HTTP Server failed"))
	}
}

func router(bank *ledger.Bank) http.Handler {
	r := pat.New()
	r.Get("/user/msisdn/{msisdn}", func(res http.ResponseWriter, req *http.Request) { UserGetMsisdn(res, req, bank) })
	r.Get("/user/{id}/login/{pin}", func(res http.ResponseWriter, req *http.Request) { UserLogin(res, req, bank) })
	r.Get("/user/{id}", func(res http.ResponseWriter, req *http.Request) { UserGetID(res, req, bank) })
	r.Post("/user", func(res http.ResponseWriter, req *http.Request) { UserAdd(res, req, bank) })

	//session: goods
	r.Post("/session/{id}/pay/goods/{goodsid}", func(res http.ResponseWriter, req *http.Request) { SessionPayGoods(res, req, bank) })
	r.Delete("/session/{id}/goods/{goodsid}", func(res http.ResponseWriter, req *http.Request) { SessionGoodsDel(res, req, bank) })
	r.Get("/session/{id}/goods", func(res http.ResponseWriter, req *http.Request) { SessionGoodsList(res, req, bank) })
	r.Post("/session/{id}/goods", func(res http.ResponseWriter, req *http.Request) { SessionGoodsAdd(res, req, bank) })

	r.Get("/session/{id}/keepalive", func(res http.ResponseWriter, req *http.Request) { SessionKeepAlive(res, req, bank) })
	r.Get("/session/{id}/ministatement", func(res http.ResponseWriter, req *http.Request) { SessionMiniStatement(res, req, bank) })

	r.Get("/session/{id}/logout", func(res http.ResponseWriter, req *http.Request) { SessionLogout(res, req, bank) })

	//for demo:
	//EFT:
	r.Post("/session/{id}/deposit", func(res http.ResponseWriter, req *http.Request) { SessionDeposit(res, req, bank) })
	return r
}
