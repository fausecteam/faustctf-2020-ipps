package main

import (
	"flag"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/grpc"
	"log"

	"github.com/BurntSushi/toml"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/http"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/internal/session"
	"gitlab.cs.fau.de/faust/faustctf-2020/ipps/pkg/postgres"
)

type config struct {
	Database *postgres.Config
	Server   *http.Config
	Session  *session.Config
	GRPC     *grpc.Config
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.toml",
		"use another configuration file")
	flag.Parse()

	conf := &config{}
	_, err := toml.DecodeFile(configPath, conf)
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.Connect(conf.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = postgres.InstallTables(db)
	if err != nil {
		log.Fatalf("error installing tables: %v\n", err)
	}
	as, err := postgres.NewAddressStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	cs, err := postgres.NewCreditCardStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer cs.Close()
	es, err := postgres.NewEventStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer es.Close()
	fs, err := postgres.NewFeedbackStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()
	ps, err := postgres.NewParcelStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	us, err := postgres.NewUserStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer us.Close()

	go runGRPCServer(conf)
	s := http.Server{
		AddressStorage:  as,
		CreditStorage:   cs,
		EventStorage:    es,
		FeedbackStorage: fs,
		ParcelStorage:   ps,
		UserStorage:     us,
	}
	log.Fatal(s.ListenAndServe(conf.Server, conf.Session))
}

func runGRPCServer(c *config) {
	db, err := postgres.Connect(c.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	as, err := postgres.NewAddressStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	cs, err := postgres.NewCreditCardStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer cs.Close()
	us, err := postgres.NewUserStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	defer us.Close()

	s, err := grpc.NewServer(c.GRPC, as, cs, us)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.ListenAndServe())
}
