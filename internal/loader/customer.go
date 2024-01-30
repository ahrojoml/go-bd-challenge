package loader

import (
	"app/internal"
	"encoding/json"
	"os"
)

type CustomerLoader struct {
	CustomerJSONPath string
	cr               internal.RepositoryCustomer
}

func NewCustomerLoader(CustomerJSONPath string, cr internal.RepositoryCustomer) *CustomerLoader {
	return &CustomerLoader{
		CustomerJSONPath: CustomerJSONPath,
		cr:               cr,
	}
}

// [{"id":1,"last_name":"Fifield","first_name":"Ike","condition":0},
type CustomerJSON struct {
	ID        int    `json:"id"`
	LastName  string `json:"last_name"`
	FirstName string `json:"first_name"`
	Condition int    `json:"condition"`
}

func JSONToCustomer(customerJSON CustomerJSON) internal.Customer {
	attr := internal.CustomerAttributes{
		LastName:  customerJSON.LastName,
		FirstName: customerJSON.FirstName,
		Condition: customerJSON.Condition,
	}

	return internal.Customer{
		Id:                 customerJSON.ID,
		CustomerAttributes: attr,
	}
}

func (c *CustomerLoader) LoadAndSave() error {
	file, err := os.Open(c.CustomerJSONPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var customers []CustomerJSON
	err = json.NewDecoder(file).Decode(&customers)
	if err != nil {
		return err
	}

	var internalCustomer internal.Customer
	for _, customer := range customers {
		internalCustomer = JSONToCustomer(customer)
		if err := c.cr.Save(&internalCustomer); err != nil {
			return err
		}
	}

	return nil
}
