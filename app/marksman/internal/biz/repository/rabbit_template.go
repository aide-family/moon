package repository

import rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

// RabbitTemplate provides access to rabbit template service via domain factories.
type RabbitTemplate interface {
	rabbitv1.TemplateServer
}

