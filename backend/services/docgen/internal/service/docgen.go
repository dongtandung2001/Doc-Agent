package service

import (
	"log"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/clients"
)

type DocGenService struct {
	aiClient      *clients.AIClient
	gatewayClient *clients.GatewayClient
	// messageQueueConsumer - TODO: Add RabbitMQ/Kafka consumer
}

func NewDocGenService(
	aiClient *clients.AIClient,
	gatewayClient *clients.GatewayClient,
) *DocGenService {
	return &DocGenService{
		aiClient:      aiClient,
		gatewayClient: gatewayClient,
	}
}

// StartWorker starts the message queue consumer
// TODO: Implement message queue consumer that processes document generation tasks
func (s *DocGenService) StartWorker() error {
	log.Println("Document generation worker started")
	log.Println("Waiting for tasks from message queue...")

	// TODO: Subscribe to message queue
	// Example:
	// for msg := range messageQueue.Subscribe("doc-generation-tasks") {
	//     s.processTask(msg)
	// }

	// For now, just block
	select {}
}
