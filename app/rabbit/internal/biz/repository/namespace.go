package repository

import magicboxapiv1 "github.com/aide-family/magicbox/api/v1"

type Namespace interface {
	magicboxapiv1.NamespaceServer
}
