package handler_test

import (
	"app/internal"
	"app/internal/handler"
	"app/internal/repository"
	"app/internal/service"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func init() {
	cfg := mysql.Config{
		User:   "root",
		Passwd: "123",
		Net:    "tcp",
		Addr:   "localhost:3306",
		DBName: "fantasy_products_test",
	}

	txdb.Register("txdb", "mysql", cfg.FormatDSN())
}

func TestInvoicesTotalByCondition(t *testing.T) {
	testCases := []struct {
		name       string
		customers  []internal.Customer
		invoices   []internal.InvoiceAttributes
		expectCode int
		expectBody string
	}{
		{
			name: "success retrieve invoices total by condition",
			customers: []internal.Customer{
				{
					Id: 1,
					CustomerAttributes: internal.CustomerAttributes{
						FirstName: "John",
						LastName:  "Doe",
						Condition: 1,
					},
				}, {
					Id: 2,
					CustomerAttributes: internal.CustomerAttributes{
						FirstName: "Jane",
						LastName:  "Doe",
						Condition: 1,
					},
				}, {
					Id: 3,
					CustomerAttributes: internal.CustomerAttributes{
						FirstName: "Johnny",
						LastName:  "Doe",
						Condition: 0,
					},
				},
			},
			invoices: []internal.InvoiceAttributes{
				{
					Datetime:   "2022-05-15 00:00:00",
					Total:      32.00,
					CustomerId: 1,
				}, {
					Datetime:   "2022-05-15 00:00:00",
					Total:      10.00,
					CustomerId: 2,
				}, {
					Datetime:   "2022-05-15 00:00:00",
					Total:      5.00,
					CustomerId: 3,
				},
			},
			expectCode: http.StatusOK,
			expectBody: `{
				"data": [
					{"condition": 1, "total": 42.00},
					{"condition": 0, "total": 5.00}
				]
			}`,
		}, {
			name:       "success no customers",
			expectCode: http.StatusOK,
			customers:  []internal.Customer{},
			invoices:   []internal.InvoiceAttributes{},
			expectBody: `{"data": []}`,
		},
	}

	for idx, testCase := range testCases {
		t.Run(fmt.Sprintf("%d - %s", idx, testCase.name), func(t *testing.T) {
			db, err := sql.Open("txdb", "fantasy_products_test")
			require.NoError(t, err)
			defer db.Close()

			defer func(db *sql.DB) {
				// delete records
				_, err := db.Exec("DELETE FROM invoices")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("DELETE FROM customers")
				if err != nil {
					panic(err)
				}
				// reset auto increment
				_, err = db.Exec("ALTER TABLE invoices AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("ALTER TABLE customers AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
			}(db)

			err = func(db *sql.DB) error {
				for _, customerAttr := range testCase.customers {
					_, err := db.Exec(
						"INSERT INTO customers (`id`, `first_name`, `last_name`, `condition`) VALUES (?, ?, ?, ?)",
						customerAttr.Id, customerAttr.FirstName, customerAttr.LastName, customerAttr.Condition,
					)

					if err != nil {
						return err
					}
				}

				return nil
			}(db)
			require.NoError(t, err)

			err = func(db *sql.DB) error {
				for _, invoiceAttr := range testCase.invoices {
					_, err := db.Exec(
						"INSERT INTO invoices (`customer_id`, `datetime`, `total`) VALUES (?, ?, ?)",
						invoiceAttr.CustomerId, invoiceAttr.Datetime, invoiceAttr.Total,
					)
					if err != nil {
						return err
					}
				}
				return nil
			}(db)
			require.NoError(t, err)

			ir := repository.NewInvoicesMySQL(db)
			is := service.NewInvoicesDefault(ir)
			h := handler.NewInvoicesDefault(is)

			request := httptest.NewRequest("GET", "/total/condition", nil)
			response := httptest.NewRecorder()

			h.InvoicesTotalByCondition()(response, request)

			require.Equal(t, testCase.expectCode, response.Code)
			require.JSONEq(t, testCase.expectBody, response.Body.String())
		})
	}
}

func TestUpdateInvoicesTotal(t *testing.T) {
	testCases := []struct {
		name         string
		products     []internal.Product
		invoices     []internal.Invoice
		sales        []internal.Sale
		expectCode   int
		expectValues map[int]float64
	}{
		{
			name: "success retrieve invoices total by condition",
			products: []internal.Product{
				{
					Id: 1,
					ProductAttributes: internal.ProductAttributes{
						Description: "Product 1",
						Price:       10.00,
					},
				},
				{
					Id: 2,
					ProductAttributes: internal.ProductAttributes{
						Description: "Product 2",
						Price:       5.00,
					},
				},
			},
			invoices: []internal.Invoice{
				{
					Id: 1,
					InvoiceAttributes: internal.InvoiceAttributes{
						Datetime:   "2022-05-15 00:00:00",
						Total:      0.00,
						CustomerId: 1,
					},
				}, {
					Id: 2,
					InvoiceAttributes: internal.InvoiceAttributes{
						Datetime:   "2022-05-15 00:00:00",
						Total:      0.00,
						CustomerId: 1,
					},
				},
			},
			sales: []internal.Sale{
				{
					Id: 1,
					SaleAttributes: internal.SaleAttributes{
						ProductId: 1,
						InvoiceId: 1,
						Quantity:  10,
					},
				},
				{
					Id: 2,
					SaleAttributes: internal.SaleAttributes{
						ProductId: 2,
						InvoiceId: 1,
						Quantity:  10,
					},
				},
				{
					Id: 3,
					SaleAttributes: internal.SaleAttributes{
						ProductId: 1,
						InvoiceId: 2,
						Quantity:  20,
					},
				},
			},
			expectCode: http.StatusOK,
			expectValues: map[int]float64{
				1: 150.00,
				2: 200.00,
			},
		},
	}

	for idx, testCase := range testCases {
		t.Run(fmt.Sprintf("%d - %s", idx, testCase.name), func(t *testing.T) {
			db, err := sql.Open("txdb", "fantasy_products_test")
			require.NoError(t, err)
			defer db.Close()

			defer func(db *sql.DB) {
				// delete records
				_, err := db.Exec("DELETE FROM sales")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("DELETE FROM products")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("DELETE FROM invoices")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("DELETE FROM customers")
				if err != nil {
					panic(err)
				}
				// reset auto increment
				_, err = db.Exec("ALTER TABLE sales AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("ALTER TABLE products AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("ALTER TABLE invoices AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
				_, err = db.Exec("ALTER TABLE customers AUTO_INCREMENT = 0")
				if err != nil {
					panic(err)
				}
			}(db)

			err = func(db *sql.DB) error {
				_, err := db.Exec(
					"INSERT INTO customers (`id`, `first_name`, `last_name`, `condition`) VALUES (?, ?, ?, ?)",
					1, "John", "Doe", 1,
				)
				if err != nil {
					return err
				}

				return nil
			}(db)
			require.NoError(t, err)

			err = func(db *sql.DB) error {
				for _, productAttr := range testCase.products {
					_, err := db.Exec(
						"INSERT INTO products (`id`, `description`, `price`) VALUES (?, ?, ?)",
						productAttr.Id, productAttr.Description, productAttr.Price,
					)
					if err != nil {
						return err
					}
				}
				return nil
			}(db)
			require.NoError(t, err)

			err = func(db *sql.DB) error {
				for _, invoiceAttr := range testCase.invoices {
					_, err := db.Exec(
						"INSERT INTO invoices (`customer_id`, `datetime`, `total`) VALUES (?, ?, ?)",
						invoiceAttr.CustomerId, invoiceAttr.Datetime, invoiceAttr.Total,
					)
					if err != nil {
						return err
					}
				}
				return nil
			}(db)
			require.NoError(t, err)

			err = func(db *sql.DB) error {
				for _, saleAttr := range testCase.sales {
					_, err := db.Exec(
						"INSERT INTO sales (`quantity`, `product_id`, `invoice_id`) VALUES (?, ?, ?)",
						saleAttr.Quantity, saleAttr.ProductId, saleAttr.InvoiceId,
					)
					if err != nil {
						return err
					}
				}
				return nil
			}(db)
			require.NoError(t, err)

			ir := repository.NewInvoicesMySQL(db)
			is := service.NewInvoicesDefault(ir)
			h := handler.NewInvoicesDefault(is)

			request := httptest.NewRequest("PUT", "/update_total", nil)
			response := httptest.NewRecorder()

			h.UpdateInvoicesTotal()(response, request)

			require.Equal(t, testCase.expectCode, response.Code)

			rows, err := db.Query("SELECT id, total FROM invoices")
			require.NoError(t, err)

			var id int
			var total float64
			for rows.Next() {
				err = rows.Scan(&id, &total)
				require.NoError(t, err)
				total, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", total), 64)
				require.Equal(t, testCase.expectValues[id], total)
			}
		})

	}
}
