package bo

import (
	"slices"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// DatasourceFilterBo scopes which datasources a strategy metric evaluates.
type DatasourceFilterBo struct {
	DatasourceUIDs          []int64
	ExcludeDatasourceUIDs   []int64
	DatasourceLabels        map[string]string
	ExcludeDatasourceLabels map[string]string
}

func NewDatasourceFilterBo(req *apiv1.DatasourceFilter) *DatasourceFilterBo {
	if req == nil {
		return &DatasourceFilterBo{}
	}
	return &DatasourceFilterBo{
		DatasourceUIDs:          req.GetDatasourceUids(),
		ExcludeDatasourceUIDs:   req.GetExcludeDatasourceUids(),
		DatasourceLabels:        req.GetDatasourceLabels(),
		ExcludeDatasourceLabels: req.GetExcludeDatasourceLabels(),
	}
}

func ToAPIV1DatasourceFilter(b *DatasourceFilterBo) *apiv1.DatasourceFilter {
	if b == nil {
		return nil
	}
	return &apiv1.DatasourceFilter{
		DatasourceUids:          b.DatasourceUIDs,
		ExcludeDatasourceUids:   b.ExcludeDatasourceUIDs,
		DatasourceLabels:        b.DatasourceLabels,
		ExcludeDatasourceLabels: b.ExcludeDatasourceLabels,
	}
}

// MatchesDatasource returns true when the datasource passes exclude/include rules.
// Empty filter matches all datasources (legacy: no datasource restriction).
func (f *DatasourceFilterBo) MatchesDatasource(datasourceUID int64, metadata map[string]string) bool {
	if f == nil {
		return true
	}
	if len(f.ExcludeDatasourceUIDs) > 0 && slices.Contains(f.ExcludeDatasourceUIDs, datasourceUID) {
		return false
	}
	if len(f.ExcludeDatasourceLabels) > 0 && metadataMatchesAll(f.ExcludeDatasourceLabels, metadata) {
		return false
	}
	hasIncludeUIDs := len(f.DatasourceUIDs) > 0
	hasIncludeLabels := len(f.DatasourceLabels) > 0
	if !hasIncludeUIDs && !hasIncludeLabels {
		return true
	}
	uidMatched := hasIncludeUIDs && slices.Contains(f.DatasourceUIDs, datasourceUID)
	labelsMatched := hasIncludeLabels && metadataMatchesAll(f.DatasourceLabels, metadata)
	return uidMatched || labelsMatched
}

func metadataMatchesAll(required map[string]string, metadata map[string]string) bool {
	if len(required) == 0 {
		return false
	}
	if len(metadata) == 0 {
		return false
	}
	for key, value := range required {
		actual, ok := metadata[key]
		if !ok || actual != value {
			return false
		}
	}
	return true
}

// IncludeFilterDatasourceUIDs returns explicitly bound include datasource UIDs (datasourceUids).
func (f *DatasourceFilterBo) IncludeFilterDatasourceUIDs() []int64 {
	if f == nil || len(f.DatasourceUIDs) == 0 {
		return nil
	}
	return slices.Clone(f.DatasourceUIDs)
}

// ExcludeFilterDatasourceUIDs returns explicitly bound exclude datasource UIDs (excludeDatasourceUids).
func (f *DatasourceFilterBo) ExcludeFilterDatasourceUIDs() []int64 {
	if f == nil || len(f.ExcludeDatasourceUIDs) == 0 {
		return nil
	}
	return slices.Clone(f.ExcludeDatasourceUIDs)
}

// FilterReferencedDatasourceUIDs returns deduplicated UIDs from include and exclude lists (include first).
func (f *DatasourceFilterBo) FilterReferencedDatasourceUIDs() []int64 {
	if f == nil {
		return nil
	}
	include := f.IncludeFilterDatasourceUIDs()
	exclude := f.ExcludeFilterDatasourceUIDs()
	if len(include) == 0 && len(exclude) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(include)+len(exclude))
	out := make([]int64, 0, len(include)+len(exclude))
	for _, uid := range include {
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	for _, uid := range exclude {
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	return out
}

// OrderDatasourceItemsByUIDs returns items in the same order as uids; missing UIDs are skipped.
func OrderDatasourceItemsByUIDs(uids []int64, items []*DatasourceItemBo) []*DatasourceItemBo {
	if len(uids) == 0 || len(items) == 0 {
		return nil
	}
	byUID := make(map[int64]*DatasourceItemBo, len(items))
	for _, item := range items {
		if item != nil {
			byUID[item.UID.Int64()] = item
		}
	}
	out := make([]*DatasourceItemBo, 0, len(uids))
	for _, uid := range uids {
		if item, ok := byUID[uid]; ok {
			out = append(out, item)
		}
	}
	return out
}

// SelectDatasourceUIDs returns datasource UIDs from the list that match the filter.
func (f *DatasourceFilterBo) SelectDatasourceUIDs(
	candidates []*DatasourceItemBo,
	allEnabledUIDs []int64,
) []int64 {
	if f == nil {
		if len(candidates) > 0 {
			out := make([]int64, 0, len(candidates))
			for _, ds := range candidates {
				if ds != nil {
					out = append(out, ds.UID.Int64())
				}
			}
			return out
		}
		return allEnabledUIDs
	}
	if len(candidates) == 0 {
		return nil
	}
	out := make([]int64, 0, len(candidates))
	for _, ds := range candidates {
		if ds == nil {
			continue
		}
		if f.MatchesDatasource(ds.UID.Int64(), ds.Metadata) {
			out = append(out, ds.UID.Int64())
		}
	}
	return out
}
