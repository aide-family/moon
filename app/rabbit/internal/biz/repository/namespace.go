package repository

import goddessv1 "github.com/aide-family/goddess/pkg/api/v1"

type Namespace interface {
	goddessv1.NamespaceServer
}
