package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lahnasti/go-market/common/models"
	"github.com/streadway/amqp"
	"github.com/wileytor/go-market/products/internal/server/responses"
	"log"
	"net/http"
	"strconv"
	"time"
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
	if err := s.EnsureConnection(); err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to connect to RabbitMQ", err)
		return
	}

	if s.Rabbit.Channel == nil {
		responses.SendError(ctx, http.StatusInternalServerError, "RabbitMQ channel is not open", nil)
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
	tempQueueName := "temp_queue"

	ch := s.Rabbit.Channel

	_, err = ch.QueueDeclare(tempQueueName, false, true, true, false, nil)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to declare temp queue", err)
		return
	}
	defer func() {
		if _, err := ch.QueueDelete(tempQueueName, false, false, false); err != nil {
			log.Printf("Failed to delete temp queue: %v", err)
		}
	}()

	mes := models.TokenCheckMessage{
		Token:     tokenStr,
		TempQueue: tempQueueName,
	}
	mesBytes, err := json.Marshal(mes)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to marshal message", err)
		return
	}
	if err := s.Rabbit.PublishMessage("user_check_queue", mesBytes); err != nil {
		log.Printf("MakePurchaseHandler: ошибка публикации сообщения: %v", err)
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to publish message to queue", err)
		return
	}
	log.Println("MakePurchaseHandler: сообщение опубликовано")

	resultChan := make(chan bool)

	go func() {
		err := s.Rabbit.ConsumeMessage(tempQueueName, func(msg amqp.Delivery) {
			tempQueueHandler(ctx, msg, s, purchase, resultChan)
		})
		if err != nil {
			log.Printf("Failed to consume temp queue message: %v", err)
			resultChan <- false
		}
	}()

	select {
	case success := <-resultChan:
		if success {
			log.Println("MakePurchaseHandler: покупка успешно обработана")
			return // Ответ отправлен в tempQueueHandler
		}
		log.Println("MakePurchaseHandler: ошибка при обработке покупки")

		responses.SendError(ctx, http.StatusInternalServerError, "Ошибка при обработке покупки", nil)
	case <-time.After(15 * time.Second):
		log.Println("MakePurchaseHandler: таймаут при обработке покупки")

		responses.SendError(ctx, http.StatusInternalServerError, "Таймаут при обработке покупки", nil)
	}
}

// Обработчик временной очереди
func tempQueueHandler(ctx *gin.Context, msg amqp.Delivery, s *Server, purchase models.Purchase, resultChan chan bool) {
	defer close(resultChan) // Закрываем канал по завершении

	var response models.TokenCheckResponse
	if err := json.Unmarshal(msg.Body, &response); err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to unmarshal response", err)
		resultChan <- false
		return
	}

	if response.Valid {
		purchase.UserID = response.UserID
		purchaseID, err := s.Db.MakePurchase(purchase)
		if err != nil {
			responses.SendError(ctx, http.StatusInternalServerError, "Purchase failed", err)
			resultChan <- false
			return
		}
		responses.SendSuccess(ctx, http.StatusOK, "Purchase successful", purchaseID)
		resultChan <- true // Успешная обработка
	} else {
		responses.SendError(ctx, http.StatusUnauthorized, "Invalid token", nil)
		resultChan <- false // Некорректный токен
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
