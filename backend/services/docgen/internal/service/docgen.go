package service

import (
	"log"

	apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
)

type DocGenService struct {
	aiClient      apiv1.AIServiceClient
	gatewayClient apiv1.GatewayServiceClient
	// messageQueueConsumer - TODO: Add RabbitMQ/Kafka consumer
}

func NewDocGenService(
	aiClient apiv1.AIServiceClient,
	gatewayClient apiv1.GatewayServiceClient,
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
