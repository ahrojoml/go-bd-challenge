package loader

import (
	"app/internal"
	"encoding/json"
	"os"
)

type SaleLoader struct {
	SaleJSONPath string
	sr           internal.RepositorySale
}

func NewSaleLoader(SaleJSONPath string, sr internal.RepositorySale) *SaleLoader {
	return &SaleLoader{
		SaleJSONPath: SaleJSONPath,
		sr:           sr,
	}
}

// {"id":1,"product_id":58,"invoice_id":45,"quantity":22}
type SaleJSON struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id"`
	InvoiceID int `json:"invoice_id"`
	Quantity  int `json:"quantity"`
}

func JSONToSale(saleJSON SaleJSON) internal.Sale {
	attr := internal.SaleAttributes{
		ProductId: saleJSON.ProductID,
		InvoiceId: saleJSON.InvoiceID,
		Quantity:  saleJSON.Quantity,
	}

	return internal.Sale{
		Id:             saleJSON.ID,
		SaleAttributes: attr,
	}

}

func (p *SaleLoader) LoadAndSave() error {
	file, err := os.Open(p.SaleJSONPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sales []SaleJSON
	err = json.NewDecoder(file).Decode(&sales)
	if err != nil {
		return err
	}

	var internalSale internal.Sale
	for _, sale := range sales {
		internalSale = JSONToSale(sale)
		if err := p.sr.Save(&internalSale); err != nil {
			return err
		}
	}

	return nil
}
