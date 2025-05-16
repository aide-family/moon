package vobj

// TypeDatasource represents the type of a datasource.
//
//go:generate stringer -type=DatasourceType -linecomment -output=type_datasource.string.go
type DatasourceType int8

const (
	DatasourceTypeUnknown DatasourceType = iota // unknown
	DatasourceTypeMetric                        // metric
	DatasourceTypeLog                           // log
	DatasourceTypeEvent                         // event
	DatasourceTypeTrace                         // trace
)
