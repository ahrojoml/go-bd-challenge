package loader

import (
	"app/internal"
	"encoding/json"
	"os"
)

type ProductLoader struct {
	ProductJSONPath string
	pr              internal.RepositoryProduct
}

func NewProductLoader(ProductJSONPath string, pr internal.RepositoryProduct) *ProductLoader {
	return &ProductLoader{
		ProductJSONPath: ProductJSONPath,
		pr:              pr,
	}
}

// {"id":1,"description":"French Pastry - Mini Chocolate","price":97.01},
type ProductJSON struct {
	ID          int     `json:"id"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func JSONToProduct(productJSON ProductJSON) internal.Product {
	attr := internal.ProductAttributes{
		Description: productJSON.Description,
		Price:       productJSON.Price,
	}
	return internal.Product{
		Id:                productJSON.ID,
		ProductAttributes: attr,
	}
}

func (p *ProductLoader) LoadAndSave() error {
	file, err := os.Open(p.ProductJSONPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var products []ProductJSON
	err = json.NewDecoder(file).Decode(&products)
	if err != nil {
		return err
	}

	var internalProduct internal.Product
	for _, product := range products {
		internalProduct = JSONToProduct(product)
		if err := p.pr.Save(&internalProduct); err != nil {
			return err
		}
	}

	return nil
}
