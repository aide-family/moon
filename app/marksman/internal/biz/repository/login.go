package repository

import goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

type LoginRepository interface {
	goddessv1.AuthServiceServer
}
