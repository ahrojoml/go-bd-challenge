package loader

import (
	"app/internal"
	"encoding/json"
	"os"
)

type InvoiceLoader struct {
	InvoiceJSONPath string
	ir              internal.RepositoryInvoice
}

func NewInvoiceLoader(InvoiceJSONPath string, ir internal.RepositoryInvoice) *InvoiceLoader {
	return &InvoiceLoader{
		InvoiceJSONPath: InvoiceJSONPath,
		ir:              ir,
	}
}

// {"id":1,"datetime":"2022-05-15","customer_id":19,"total":0.0},
type InvoiceJSON struct {
	ID         int     `json:"id"`
	Datetime   string  `json:"datetime"`
	CustomerID int     `json:"customer_id"`
	Total      float64 `json:"total"`
}

func JSONToInvoice(invoiceJSON InvoiceJSON) internal.Invoice {
	attr := internal.InvoiceAttributes{
		Datetime:   invoiceJSON.Datetime,
		CustomerId: invoiceJSON.CustomerID,
		Total:      invoiceJSON.Total,
	}
	return internal.Invoice{
		Id:                invoiceJSON.ID,
		InvoiceAttributes: attr,
	}
}

func (c *InvoiceLoader) LoadAndSave() error {
	file, err := os.Open(c.InvoiceJSONPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var invoices []InvoiceJSON
	err = json.NewDecoder(file).Decode(&invoices)
	if err != nil {
		return err
	}

	var internalInvoice internal.Invoice
	for _, invoice := range invoices {
		internalInvoice = JSONToInvoice(invoice)
		if err := c.ir.Save(&internalInvoice); err != nil {
			return err
		}
	}

	return nil
}
