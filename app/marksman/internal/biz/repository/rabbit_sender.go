package repository

import rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

// RabbitSender provides access to rabbit sender service via domain factories.
type RabbitSender interface {
	rabbitv1.SenderServer
}

