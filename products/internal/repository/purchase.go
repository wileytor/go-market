package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/wileytor/go-market/common/models"
)

func (db *DBstorage) MakePurchase(purchase models.Purchase) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer func(tx pgx.Tx) {
		if p := recover(); p != nil {
			tx.Rollback(ctx) // Восстановление после паники
			panic(p)         // Повторный вызов паники
		} else if err != nil {
			tx.Rollback(ctx) // Откат транзакции в случае ошибки
		}
	}(tx)

	/*sb := sqlbuilder.NewUpdateBuilder()
	updQuery, args := sb.Update("products").
		Set("quantity", sb.Add("quantity", -purchase.Quantity)).
		Where(
			sb.Equal("uid", purchase.ProductID),
			sb.GreaterEqualThan("quantity", purchase.Quantity),
		).Build()

	result, err := tx.Exec(ctx, updQuery, args...)
	if err != nil {
		return -1, err
	}*/
	updQuery := `UPDATE products SET quantity = quantity - $1 WHERE uid = $2 AND quantity >= $3`
	result, err := tx.Exec(ctx, updQuery, purchase.Quantity, purchase.ProductID, purchase.Quantity)
	if err != nil {
		return -1, err
	}
	// Проверяем, затронута ли строка (т.е. продукт в наличии в нужном количестве)
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return -1, fmt.Errorf("not enough product quantity available or product does not exist")
	}

	/*insertSb := sqlbuilder.NewInsertBuilder()
	query, args := insertSb.InsertInto("purchases").
		Cols("user_id", "product_id", "quantity").
		Values(purchase.UserID, purchase.ProductID, purchase.Quantity).
		Build()
			row := tx.QueryRow(ctx, query, args...)

	*/
	insertQuery := `INSERT INTO purchases (user_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING uid`
	row := tx.QueryRow(ctx, insertQuery, purchase.UserID, purchase.ProductID, purchase.Quantity)

	var UID int
	if err := row.Scan(&UID); err != nil {
		return -1, err
	}
	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return -1, err
	}
	return UID, nil
}

func (db *DBstorage) GetUserPurchases(userID int) ([]models.Purchase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("*").From("purchases").Where(sb.Equal("user_id", userID)).Build()

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var purchases []models.Purchase
	for rows.Next() {
		var purchase models.Purchase
		if err := rows.Scan(&purchase.UID, &purchase.UserID, &purchase.ProductID, &purchase.Quantity); err != nil {
			return nil, err
		}
		purchases = append(purchases, purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return purchases, nil
}

func (db *DBstorage) GetProductPurchases(productID int) ([]models.Purchase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("*").From("purchases").Where(sb.Equal("product_id", productID)).Build()

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var purchases []models.Purchase
	for rows.Next() {
		var purchase models.Purchase
		if err := rows.Scan(&purchase.UID, &purchase.UserID, &purchase.ProductID, &purchase.Quantity); err != nil {
			return nil, err
		}
		purchases = append(purchases, purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return purchases, nil
}
