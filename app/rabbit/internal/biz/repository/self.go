package repository

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type Self interface {
	goddessv1.SelfServer
}
