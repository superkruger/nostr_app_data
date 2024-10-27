package main

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/superkruger/nostr_app_data/app/utils/aws/secrets"
)

type config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Host     string `json:"host"`
}

func main() {
	var cnfg config
	err := secrets.NewService().GetAndUnmarshal("test/nostr/mongo/rw", &cnfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("got secret")
	m, err := migrate.New(
		"file://.",
		fmt.Sprintf("mongodb+srv://%v:%v@%v/%v", cnfg.Username, cnfg.Password, cnfg.Host, cnfg.Database),
	)
	log.Print("migrate initialized")
	if err != nil {
		log.Fatal(err)
	}
	err = m.Steps(1)
	if err != nil {
		log.Fatal(err)
	}
}
