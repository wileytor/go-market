package server

import (
	"encoding/json"
	"log"

	"github.com/lahnasti/go-market/common/models"
	"github.com/streadway/amqp"
)

func StartListener(s *Server) {
	err := s.Rabbit.ConsumeMessage("user_check_queue", s.TokenCheck)
	if err != nil {
		log.Fatalf("failed to consume message: %v", err)
	}
}

func (s *Server) TokenCheck(msg amqp.Delivery) {
	var request models.TokenCheckMessage
	if err := json.Unmarshal(msg.Body, &request); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	userID, err := CheckToken(request.Token)
	valid := err == nil

	response := models.TokenCheckResponse{
		Valid:  valid,
		UserID: userID,
	}
	if !valid {
		response.Error = "Invalid token"
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// Отправляем ответ в указанную temp_queue
	if err := s.Rabbit.PublishMessage(request.TempQueue, responseBytes); err != nil {
		log.Printf("Failed to publish response to temp_queue: %v", err)
	}
}
