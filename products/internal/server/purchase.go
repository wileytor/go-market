package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/lahnasti/go-market/common/models"
	"github.com/lahnasti/go-market/products/internal/server/responses"
	"github.com/streadway/amqp"
)

// MakePurchaseHandler обрабатывает создание новой покупки
// @Summary Создание покупки
// @Description Создает новую покупку для указанного продукта
// @Tags Покупки
// @Accept json
// @Produce json
// @Param purchase body models.Purchase true "Данные о покупке"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /purchases/add [post]
func (s *Server) MakePurchaseHandler(ctx *gin.Context) {
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		responses.SendError(ctx, http.StatusUnauthorized, "The header is missing", nil)
		return
	}
	var purchase models.Purchase
	if err := ctx.ShouldBindJSON(&purchase); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}
	if err := s.Valid.Struct(purchase); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Not a valid purchase", err)
		return
	}
	_, err := s.Db.GetProductByID(purchase.ProductID)
	if err != nil {
		responses.SendError(ctx, http.StatusNotFound, "Product not found", err)
		return
	}
	if purchase.Quantity <= 0 {
		responses.SendError(ctx, http.StatusBadRequest, "Quantity must be greater than 0", nil)
		return
	}
	tempQueueName := "temp_queue_" + uuid.New().String()
	ch := s.Rabbit.Channel

	_, err = ch.QueueDeclare(tempQueueName, false, true, true, false, nil)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to declare temp queue", err)
		return
	}
	defer ch.QueueDelete(tempQueueName, false, false, false)

	mes := models.TokenCheckMessage{
		Token: tokenStr,
	}
	mesBytes, err := json.Marshal(mes)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to marshal message", err)
		return
	}
	if err := s.Rabbit.PublishMessage("user_check_queue", mesBytes); err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to publish message to queue", err)
		return
	}

	go func() {
		err := s.Rabbit.ConsumeMessage(tempQueueName, func(msg amqp.Delivery) {
			tempQueueHandler(ctx, msg, s, purchase)
		})
		if err != nil {
			log.Printf("Failed to consume temp queue message: %v", err)
		}
	}()

	// Устанавливаем таймаут для ответа
	time.Sleep(5 * time.Second)
}

// Обработчик временной очереди
func tempQueueHandler(ctx *gin.Context, msg amqp.Delivery, s *Server, purchase models.Purchase) {
	var response models.TokenCheckResponse
	if err := json.Unmarshal(msg.Body, &response); err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to unmarshal response", err)
		return
	}

	if response.Valid {
		// Выполняем покупку, если токен валиден
		purchase.UserID = response.UserID
		purchaseID, err := s.Db.MakePurchase(purchase)
		if err != nil {
			responses.SendError(ctx, http.StatusInternalServerError, "Purchase failed", err)
			return
		}
		responses.SendSuccess(ctx, http.StatusOK, "Purchase successful", purchaseID)
	} else {
		responses.SendError(ctx, http.StatusUnauthorized, "Invalid token", nil)
	}
}

// GetUserPurchasesHandler получает покупки пользователя
// @Summary Получение списка покупок пользователя
// @Description Возвращает список покупок для указанного пользователя
// @Tags Покупки
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /purchases/user/{id} [get]
func (s *Server) GetUserPurchasesHandler(ctx *gin.Context) {
	userID := ctx.Param("id")
	uIdInt, err := strconv.Atoi(userID)
	if err != nil || uIdInt <= 0 {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid user id", err)
		return
	}
	purchases, err := s.Db.GetUserPurchases(uIdInt)
	if err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid user id", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusOK, "List purchase found", purchases)
}

// GetProductPurchasesHandler получает покупки по продукту
// @Summary Получение списка покупок по продукту
// @Description Возвращает список покупок для указанного продукта
// @Tags Покупки
// @Produce json
// @Param id path int true "ID продукта"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /purchases/product/{id} [get]
func (s *Server) GetProductPurchasesHandler(ctx *gin.Context) {
	productId := ctx.Param("id")
	uIdInt, err := strconv.Atoi(productId)
	if err != nil || uIdInt <= 0 {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid product id", err)
		return
	}
	purchases, err := s.Db.GetProductPurchases(uIdInt)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusOK, "List purchase found", purchases)
}
