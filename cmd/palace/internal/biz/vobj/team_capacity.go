package vobj

// TeamCapacity defines the tenant capacity (number of people) gradient
//
//go:generate stringer -type=TeamCapacity -linecomment -output=team_capacity.string.go
type TeamCapacity int8

const (
	TeamCapacityUnknown TeamCapacity = iota // 0
	TeamCapacityDefault                     // 10
	TeamCapacityMini                        // 20
	TeamCapacitySmall                       // 50
	TeamCapacityMedium                      // 100
	TeamCapacityLarge                       // 500
	TeamCapacityXLarge                      // 1000
	TeamCapacityMore                        // >1000
)

func (i TeamCapacity) AllowGroup() bool {
	return i >= TeamCapacitySmall
}
