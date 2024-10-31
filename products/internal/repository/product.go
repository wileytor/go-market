package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/wileytor/go-market/common/models"
)

func (db *DBstorage) GetAllProducts() ([]models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("*").From("products").BuildWithFlavor(sqlbuilder.PostgreSQL)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.UID, &product.Name, &product.Description, &product.Price, &product.Delete, &product.Quantity); err != nil {
			return nil, err
		}
		product.Name = strings.TrimSpace(product.Name)
		product.Description = strings.TrimSpace(product.Description)
		products = append(products, product)
	}
	return products, nil
}

func (db *DBstorage) GetProductByID(uid int) (models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("*").From("products").Where(sb.Equal("uid", uid)).BuildWithFlavor(sqlbuilder.PostgreSQL)

	row := db.Pool.QueryRow(ctx, query, args...)
	var product models.Product
	if err := row.Scan(&product.UID, &product.Name, &product.Description, &product.Price, &product.Delete, &product.Quantity); err != nil {
		return models.Product{}, err
	}
	product.Name = strings.TrimSpace(product.Name)
	product.Description = strings.TrimSpace(product.Description)
	return product, nil
}

func (db *DBstorage) AddProduct(product models.Product) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewInsertBuilder()
	query, args := sb.InsertInto("products").Cols("name", "description", "price", "quantity").
		Values(product.Name, product.Description, product.Price, product.Quantity).
		BuildWithFlavor(sqlbuilder.PostgreSQL)
	query += " RETURNING uid"

	var UID int
	err := db.Pool.QueryRow(ctx, query, args...).Scan(&UID)
	if err != nil {
		return -1, fmt.Errorf("failed to insert product: %w", err)
	}
	return UID, nil
}

func (db *DBstorage) UpdateProduct(uid int, product models.Product) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sb := sqlbuilder.NewUpdateBuilder()
	query, args := sb.Update("products").
		Set(
			sb.Assign("name", product.Name),
			sb.Assign("description", product.Description),
			sb.Assign("price", product.Price),
			sb.Assign("quantity", product.Quantity),
		).
		Where(sb.Equal("uid", uid)).
		BuildWithFlavor(sqlbuilder.PostgreSQL)
	fmt.Printf("Generated query: %s, args: %v\n", query, args)
	query += " RETURNING uid"

	var UID int
	err := db.Pool.QueryRow(ctx, query, args...).Scan(&UID)
	if err != nil {
		return -1, fmt.Errorf("update user failed: %w", err)
	}
	return UID, nil
}

func (db *DBstorage) SetDeleteStatus(uid int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewUpdateBuilder()
	query, args := sb.Update("products").
		Set(sb.Assign("delete", true)).Where(sb.Equal("uid", uid)).BuildWithFlavor(sqlbuilder.PostgreSQL)

	_, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to set delete status: %w", err)
	}
	return nil
}

func (db *DBstorage) DeleteProducts() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create transaction failed: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Получаем id продуктов для удаления
	sbProductsSelect := sqlbuilder.NewSelectBuilder()
	productQuery, productArgs := sbProductsSelect.Select("uid").
		From("products").
		Where(sbProductsSelect.Equal("delete", true)).
		BuildWithFlavor(sqlbuilder.PostgreSQL)

	var productIDs []int
	rows, err := tx.Query(ctx, productQuery, productArgs...)
	if err != nil {
		return fmt.Errorf("failed to select products for deletion: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan product id: %w", err)
		}
		productIDs = append(productIDs, id)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	var productIDInterfaces []interface{}
	for _, id := range productIDs {
		productIDInterfaces = append(productIDInterfaces, id)
	}

	// Удаляем записи из purchases по полученным productIDs
	if len(productIDs) > 0 {
		sbPurchases := sqlbuilder.NewDeleteBuilder()
		purchaseQuery, purchaseArgs := sbPurchases.DeleteFrom("purchases").
			Where(sbPurchases.In("product_id", productIDInterfaces...)).
			BuildWithFlavor(sqlbuilder.PostgreSQL)

		if _, err := tx.Exec(ctx, purchaseQuery, purchaseArgs...); err != nil {
			return fmt.Errorf("failed to delete purchases: %w", err)
		}
	}

	// Удаляем продукты
	sbDeleteProducts := sqlbuilder.NewDeleteBuilder()
	deleteQuery, deleteArgs := sbDeleteProducts.DeleteFrom("products").
		Where(sbDeleteProducts.Equal("delete", true)).
		BuildWithFlavor(sqlbuilder.PostgreSQL)

	if _, err := tx.Exec(ctx, deleteQuery, deleteArgs...); err != nil {
		return fmt.Errorf("failed to delete products: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Дополнительная функция для проверки уникальности имени продукта
func (db *DBstorage) IsProductUnique(name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("COUNT(*)").From("products").Where(sb.Equal("name", name)).BuildWithFlavor(sqlbuilder.PostgreSQL)

	fmt.Printf("Generated query: %s, args: %v\n", query, args)
	var count int
	row := db.Pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}
	return count == 0, nil
}
