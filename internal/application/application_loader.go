package application

import (
	"app/internal/loader"
	"app/internal/repository"
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

type ConfigApplicationLoader struct {
	Db           *mysql.Config
	CustomerPath string
	InvoicePath  string
	ProductPath  string
	SalePath     string
}

type ApplicationLoader struct {
	config         *ConfigApplicationLoader
	db             *sql.DB
	customerLoader *loader.CustomerLoader
	invoiceLoader  *loader.InvoiceLoader
	productLoader  *loader.ProductLoader
	saleLoader     *loader.SaleLoader
}

func NewApplicationLoader(config *ConfigApplicationLoader) *ApplicationLoader {
	return &ApplicationLoader{
		config: config,
	}
}

func (a *ApplicationLoader) SetUp() error {
	db, err := sql.Open("mysql", a.config.Db.FormatDSN())
	if err != nil {
		return err
	}

	a.db = db

	rc := repository.NewCustomersMySQL(a.db)
	cl := loader.NewCustomerLoader(a.config.CustomerPath, rc)
	a.customerLoader = cl

	ri := repository.NewInvoicesMySQL(a.db)
	il := loader.NewInvoiceLoader(a.config.InvoicePath, ri)
	a.invoiceLoader = il

	rp := repository.NewProductsMySQL(a.db)
	pl := loader.NewProductLoader(a.config.ProductPath, rp)
	a.productLoader = pl

	rs := repository.NewSalesMySQL(a.db)
	sl := loader.NewSaleLoader(a.config.SalePath, rs)
	a.saleLoader = sl

	return nil
}

func (a *ApplicationLoader) Run() error {
	if err := a.customerLoader.LoadAndSave(); err != nil {
		return err
	}

	if err := a.invoiceLoader.LoadAndSave(); err != nil {
		return err
	}

	if err := a.productLoader.LoadAndSave(); err != nil {
		return err
	}

	if err := a.saleLoader.LoadAndSave(); err != nil {
		return err
	}

	return nil
}
