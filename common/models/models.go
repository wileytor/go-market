package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
}

type Credentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Product struct {
	UID         int     `json:"uid"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,min=1"`
	Delete      bool    `json:"delete"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
}

type Purchase struct {
	UID          int       `json:"uid"`
	UserID       int       `json:"userID" validate:"required"`
	ProductID    int       `json:"productID" validate:"required"`
	Quantity     int       `json:"quantity" validate:"required"`
	PurchaseDate time.Time `json:"purchase_date"`
}

/*{
	"name": "apple",
	"description": "red",
	"price": 100,
	"quantity": 10
  }*/