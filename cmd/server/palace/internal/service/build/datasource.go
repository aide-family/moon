package build

import (
	"encoding/json"

	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
)

type DatasourceBuild struct {
	*bizmodel.Datasource
}

func NewDatasourceBuild(datasource *bizmodel.Datasource) *DatasourceBuild {
	return &DatasourceBuild{
		Datasource: datasource,
	}
}

func (b *DatasourceBuild) ToApi() *admin.Datasource {
	if types.IsNil(b) || types.IsNil(b.Datasource) {
		return nil
	}
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(b.Config), &configMap); err != nil {
		log.Warnw("error", err)
	}
	return &admin.Datasource{
		Id:        b.ID,
		Name:      b.Name,
		Type:      api.DatasourceType(b.Category),
		Endpoint:  b.Endpoint,
		Status:    api.Status(b.Status),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		Config:    configMap,
		Remark:    b.Remark,
	}
}
