package page

// Normalize 把页码和每页条数收进 1 到 100。缺省页码为 1，缺省每页 20 条。
func Normalize(page, pageSize int32) (int, int) {
	p := int(page)
	if p < 1 {
		p = 1
	}
	s := int(pageSize)
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	return p, s
}

// Limit 把下拉查询的条数收进 1 到 100，缺省 20。
func Limit(limit int32) int {
	n := int(limit)
	if n < 1 {
		n = 20
	}
	if n > 100 {
		n = 100
	}
	return n
}
