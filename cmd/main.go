package main

import (
	"app/internal/application"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// env
	err := godotenv.Load(".local_env")
	if err != nil {
		fmt.Println(err)
		return
	}
	// ...

	// app
	// - config
	cfg := &application.ConfigApplicationDefault{
		Db: &mysql.Config{
			User:   "root",
			Passwd: os.Getenv("SERVER_PASSWD"),
			Net:    "tcp",
			Addr:   "localhost:3306",
			DBName: "fantasy_products",
		},
		Addr: "127.0.0.1:8080",
	}

	// Comment this after load
	// cfgLoader := &application.ConfigApplicationLoader{
	// 	Db:           cfg.Db,
	// 	CustomerPath: os.Getenv("CUSTOMER_PATH"),
	// 	InvoicePath:  os.Getenv("INVOICE_PATH"),
	// 	ProductPath:  os.Getenv("PRODUCT_PATH"),
	// 	SalePath:     os.Getenv("SALE_PATH"),
	// }

	// loaderApp := application.NewApplicationLoader(cfgLoader)
	// loaderApp.SetUp()
	// loaderApp.Run()

	app := application.NewApplicationDefault(cfg)
	// - set up
	err = app.SetUp()
	if err != nil {
		fmt.Println(err)
		return
	}
	// - run
	err = app.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
