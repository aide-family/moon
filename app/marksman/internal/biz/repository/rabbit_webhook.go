package repository

import rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

// RabbitWebhook provides access to rabbit webhook service via domain factories.
type RabbitWebhook interface {
	rabbitv1.WebhookServer
}

