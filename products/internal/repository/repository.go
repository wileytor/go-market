package repository

import "github.com/wileytor/go-market/common/models"

type Repository interface {
	PurchaseRepository
	ProductRepository
}

type PurchaseRepository interface {
	MakePurchase(models.Purchase) (int, error)
	GetUserPurchases(int) ([]models.Purchase, error)
	GetProductPurchases(int) ([]models.Purchase, error)
}

type ProductRepository interface {
	GetAllProducts() ([]models.Product, error)
	GetProductByID(int) (models.Product, error)
	AddProduct(models.Product) (int, error)
	UpdateProduct(int, models.Product) (int, error)
	DeleteProducts() error
	SetDeleteStatus(int) error
	IsProductUnique(string) (bool, error)
}
