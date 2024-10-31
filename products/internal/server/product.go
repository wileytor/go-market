package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wileytor/go-market/common/models"
	"github.com/wileytor/go-market/products/internal/server/responses"
)

func (s *Server) deleter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("Deleter shutting down")
			return
		case <-time.After(time.Second):
			if len(s.deleteChan) >= 5 { // Проверьте, что количество для удаления >= 5
				for i := 0; i < 5; i++ {
					select {
					case uid := <-s.deleteChan:
						// Обрабатываем UID для удаления
						s.log.Info().Int("uid", uid).Msg("Marking product for deletion")
					default:
						// Если канал пуст, выходим из цикла
						return
					}
				}
				if err := s.Db.DeleteProducts(); err != nil {
					s.ErrorChan <- err
					s.log.Error().Err(err).Msg("Failed to delete products")

					return
				}
			}
		}
	}
}

// GetAllProductsHandler получает все продукты
// @Summary Получить список всех продуктов
// @Description Возвращает список всех продуктов
// @Tags Продукты
// @Produce json
// @Success 200 {object} responses.Success
// @Failure 500 {object} responses.Error
// @Router /products [get]
func (s *Server) GetAllProductsHandler(ctx *gin.Context) {
	products, err := s.Db.GetAllProducts()
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to retrieve products", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusOK, "List of products", products)
}

// GetProductByIDHandler получает проукты по id
// @Summary Получение списка продуктов по id
// @Description Получить продукт по ID
// @Tags Продукты
// @Param id path int true "Product ID"
// @Produce json
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Router /products/{id} [get]
func (s *Server) GetProductByIDHandler(ctx *gin.Context) {
	uid := ctx.Param("id")
	uIdInt, err := strconv.Atoi(uid)
	if err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}
	product, err := s.Db.GetProductByID(uIdInt)
	if err != nil {
		log.Println("Error retrieving product:", err) // Добавьте это
		if errors.Is(err, responses.ErrNotFound) {
			responses.SendError(ctx, http.StatusNotFound, "Product not found", err)
			return
		} else {
			responses.SendError(ctx, http.StatusInternalServerError, "message", err)
			return
		}
	}
	responses.SendSuccess(ctx, http.StatusOK, "Product found", product)
}

// AddProductHandler добавить новые продукт
// @Summary Добавление нового продукта
// @Description Создает новый продукт
// @Tags Продукты
// @Accept json
// @Produce json
// @Param product body models.Product true "Product data"
// @Success 201 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /products/add [post]
func (s *Server) AddProductHandler(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}
	if err := s.Valid.Struct(product); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Not a valid product", err)
		return
	}

	if product.Quantity < 0 {
		responses.SendError(ctx, http.StatusBadRequest, "Quantity cannot be negative", nil)
		return
	}
	exists, err := s.Db.IsProductUnique(product.Name)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	if !exists {
		responses.SendError(ctx, http.StatusBadRequest, "Product name already exists", nil)
		return
	}

	productUID, err := s.Db.AddProduct(product)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusCreated, "Product added", productUID)
}

// UpdateProductHandler обновляет данные продукта
// @Summary Обновление продукта
// @Description Обновить данные продукта
// @Tags Продукты
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body models.Product true "Product data"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /products/{id} [put]
func (s *Server) UpdateProductHandler(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if err := s.Valid.Struct(product); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Not a valid product", err)
		return
	}

	uid := ctx.Param("id")
	uIdInt, err := strconv.Atoi(uid)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	product.UID = uIdInt

	productUID, err := s.Db.UpdateProduct(uIdInt, product)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}

	responses.SendSuccess(ctx, http.StatusOK, "Product updated", productUID)
}

// / DeleteProductHandler удаление продукта
// @Summary Удаляет продукты по ID
// @Description Удалить продукт по ID
// @Tags Продукты
// @Param id path int true "Product ID"
// @Produce json
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /products/{id} [delete]
func (s *Server) DeleteProductHandler(ctx *gin.Context) {
	uid := ctx.Param("id")
	uIdInt, err := strconv.Atoi(uid)
	if err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid id", err)
		return
	}
	err = s.Db.SetDeleteStatus(uIdInt)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	s.deleteChan <- uIdInt

	responses.SendSuccess(ctx, http.StatusOK, "Product deleted", uIdInt)
}
