package repository

import rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

// RabbitAlert provides access to rabbit alert service.
type RabbitAlert interface {
	rabbitv1.AlertServer
}
