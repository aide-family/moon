package repository

import rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

// RabbitEmail provides access to rabbit email service.
type RabbitEmail interface {
	rabbitv1.EmailServer
}

