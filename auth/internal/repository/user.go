package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/lahnasti/go-market/common/models"
	"golang.org/x/crypto/bcrypt"
)

func (db *DBstorage) GetUserProfile(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("*").From("users").Where(sb.Equal("id", id)).BuildWithFlavor(sqlbuilder.PostgreSQL)

	row := db.Pool.QueryRow(ctx, query, args...)
	var user models.User
	//Нужно ли пароль выводить?
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (db *DBstorage) RegisterUser(user models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sb := sqlbuilder.NewInsertBuilder()
	query, args := sb.InsertInto("users").Cols("username", "email", "password").
		Values(user.Username, user.Email, user.Password).
		BuildWithFlavor(sqlbuilder.PostgreSQL)
	query += " RETURNING id"
	var ID int
	err := db.Pool.QueryRow(ctx, query, args...).Scan(&ID)
	if err != nil {
		return -1, fmt.Errorf("failes to insert user: %w", err)
	}
	return ID, nil
}

func (db *DBstorage) LoginUser(username string, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var userID int
	var hashedPassword string

	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("id", "password").From("users").Where(sb.Equal("username", username)).BuildWithFlavor(sqlbuilder.PostgreSQL)
	row := db.Pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&userID, &hashedPassword); err != nil {
		if err == sql.ErrNoRows {
			return -1, fmt.Errorf("user not found")
		}
		return -1, err
	}
	// Проверка введенного пароля
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Пароль неверный
		return -1, fmt.Errorf("invalid password")
	}

	// Успешная авторизация
	return userID, nil
}

func (db *DBstorage) IsUsernameUnique(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	sb := sqlbuilder.NewSelectBuilder()
	query, args := sb.Select("COUNT(*)").From("users").Where(sb.Equal("username", username)).BuildWithFlavor(sqlbuilder.PostgreSQL)
	err := db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return count == 0, nil
}

func (db *DBstorage) UserExists(id int)(bool, error){
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)"
	err := db.Pool.QueryRow(context.Background(), query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return exists, nil
}