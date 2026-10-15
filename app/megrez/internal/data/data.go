package data

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewData, NewLocker, NewGateway, NewRangeQuery)

type Data struct{}

func NewData() (*Data, func(), error) {
	return &Data{}, func() {}, nil
}
