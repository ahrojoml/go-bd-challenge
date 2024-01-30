package internal

// ServiceProduct is the interface that wraps the basic Product methods.
type ServiceProduct interface {
	// FindAll returns all products.
	FindAll() (p []Product, err error)
	GetTopProducts() ([]TopProduct, error)
	// Save saves a product.
	Save(p *Product) (err error)
}
