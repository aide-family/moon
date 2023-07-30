package times

type UnixTime interface {
	Unix() int64
}

// TimeToUnix convert time.Time to unix timestamp
func TimeToUnix(t UnixTime) int64 {
	if t == nil {
		return 0
	}

	return t.Unix()
}
