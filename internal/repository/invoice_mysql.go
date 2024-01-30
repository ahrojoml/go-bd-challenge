package repository

import (
	"database/sql"

	"app/internal"
)

const (
	UpdateInvoicesTotalQuery                 = "UPDATE invoices AS i SET i.`total` = (SELECT SUM(s.`quantity` * p.`price`) FROM sales AS s INNER JOIN products AS p ON s.`product_id` = p.`id` WHERE i.`id` = s.`invoice_id`)"
	GetInvoicesTotalByCustomerConditionQuery = "SELECT c.`condition`, SUM(i.`total`) FROM (customers as c INNER JOIN invoices as i) GROUP BY c.`condition`"
)

// NewInvoicesMySQL creates new mysql repository for invoice entity.
func NewInvoicesMySQL(db *sql.DB) *InvoicesMySQL {
	return &InvoicesMySQL{db}
}

// InvoicesMySQL is the MySQL repository implementation for invoice entity.
type InvoicesMySQL struct {
	// db is the database connection.
	db *sql.DB
}

// FindAll returns all invoices from the database.
func (r *InvoicesMySQL) FindAll() (i []internal.Invoice, err error) {
	// execute the query
	rows, err := r.db.Query("SELECT `id`, `datetime`, `total`, `customer_id` FROM invoices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var iv internal.Invoice
		// scan the row into the invoice
		err := rows.Scan(&iv.Id, &iv.Datetime, &iv.Total, &iv.CustomerId)
		if err != nil {
			return nil, err
		}
		// append the invoice to the slice
		i = append(i, iv)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	return
}

// Save saves the invoice into the database.
func (r *InvoicesMySQL) Save(i *internal.Invoice) (err error) {
	// execute the query
	res, err := r.db.Exec(
		"INSERT INTO invoices (`datetime`, `total`, `customer_id`) VALUES (?, ?, ?)",
		(*i).Datetime, (*i).Total, (*i).CustomerId,
	)
	if err != nil {
		return err
	}

	// get the last inserted id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set the id
	(*i).Id = int(id)

	return
}

func (r *InvoicesMySQL) UpdateInvoicesTotal() error {
	_, err := r.db.Exec(UpdateInvoicesTotalQuery)
	if err != nil {
		return err
	}
	return nil
}

func (r *InvoicesMySQL) GetInvoicesTotalByCustomerCondition() ([]internal.InvoiceTotalByCustomerCondition, error) {
	rows, err := r.db.Query(GetInvoicesTotalByCustomerConditionQuery)
	if err != nil {
		return nil, err
	}

	invoicesTotalByCustomerCondition := make([]internal.InvoiceTotalByCustomerCondition, 0)
	for rows.Next() {
		var invoiceTotalByCustomerCondition internal.InvoiceTotalByCustomerCondition
		err := rows.Scan(&invoiceTotalByCustomerCondition.Condition, &invoiceTotalByCustomerCondition.Total)
		if err != nil {
			return nil, err
		}
		invoicesTotalByCustomerCondition = append(invoicesTotalByCustomerCondition, invoiceTotalByCustomerCondition)
	}

	return invoicesTotalByCustomerCondition, nil
}
