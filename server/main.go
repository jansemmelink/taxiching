package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/pat"
	"github.com/jansemmelink/log"

	_ "github.com/jansemmelink/taxiching/lib/goods/memory"
	_ "github.com/jansemmelink/taxiching/lib/sessions/memory"
	_ "github.com/jansemmelink/taxiching/lib/users/memory"
	_ "github.com/jansemmelink/taxiching/lib/wallets/memory"
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
	createBank()

	fmt.Fprintf(os.Stdout, "Serving %s\n", *addrFlag)
	if err := http.ListenAndServe(*addrFlag, router()); err != nil {
		panic(log.Wrapf(err, "HTTP Server failed"))
	}
}

func router() http.Handler {
	r := pat.New()
	r.Get("/user/msisdn/{msisdn}", UserGetMsisdn)
	r.Get("/user/{id}/login/{pin}", UserLogin)
	r.Get("/user/{id}", UserGetID)
	r.Post("/user", UserAdd)

	//session: goods
	r.Post("/session/{id}/pay/goods/{goodsid}", SessionPayGoods)
	r.Delete("/session/{id}/goods/{goodsid}", SessionGoodsDel)
	r.Get("/session/{id}/goods", SessionGoodsList)
	r.Post("/session/{id}/goods", SessionGoodsAdd)

	r.Get("/session/{id}/keepalive", SessionKeepAlive)
	r.Get("/session/{id}/ministatement", SessionMiniStatement)

	r.Get("/session/{id}/logout", SessionLogout)

	//for demo:
	//EFT:
	r.Post("/session/{id}/deposit", SessionDeposit)
	return r
}
