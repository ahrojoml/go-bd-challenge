package handler_test

import (
	"app/internal"
	"app/internal/handler"
	"app/internal/repository"
	"app/internal/service"
	"database/sql"
	"fmt"
	"net/http/httptest"
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

func TestGetTopProducts(t *testing.T) {
	testCases := []struct {
		name       string
		sales      []internal.SaleAttributes
		products   []internal.Product
		expectCode int
		expectBody string
	}{
		{
			name: "success retrieve top products",
			sales: []internal.SaleAttributes{
				{
					Quantity:  10,
					ProductId: 1,
					InvoiceId: 1,
				}, {
					Quantity:  5,
					ProductId: 2,
					InvoiceId: 1,
				}, {
					Quantity:  10,
					ProductId: 1,
					InvoiceId: 1,
				},
			},
			products: []internal.Product{
				{
					Id: 1,
					ProductAttributes: internal.ProductAttributes{
						Description: "Product 1",
					},
				}, {
					Id: 2,
					ProductAttributes: internal.ProductAttributes{
						Description: "Product 2",
					},
				},
			},
			expectCode: 200,
			expectBody: `{
				"data": [
					{"id": 1, "description": "Product 1", "total": 20},
					{"id": 2, "description": "Product 2", "total": 5}
				]
			}`,
		}, {
			name:       "success retrieve top products empty",
			sales:      []internal.SaleAttributes{},
			products:   []internal.Product{},
			expectCode: 200,
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

				_, err = db.Exec(
					"INSERT INTO invoices (`id`, `customer_id`, `datetime`, `total`) VALUES (?, ?, ?, ?)",
					1, 1, "2021-01-01 00:00:00", 42.00,
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
						"INSERT INTO products (`id`, `description`) VALUES (?, ?)",
						productAttr.Id, productAttr.Description,
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

			pr := repository.NewProductsMySQL(db)
			ps := service.NewProductsDefault(pr)
			h := handler.NewProductsDefault(ps)

			request := httptest.NewRequest("GET", "/products/top", nil)
			response := httptest.NewRecorder()

			h.GetTopProducts()(response, request)

			require.Equal(t, testCase.expectCode, response.Code)
			require.JSONEq(t, testCase.expectBody, response.Body.String())

		})
	}
}
