package field

import (
	"testing"
)

func TestStringField_Eq(t *testing.T) {
	t.Log(StringField("name").Eq("hello"))
	t.Log(StringField("name").Neq("hello"))
	t.Log(StringField("name").In("hello", "world"))
	t.Log(StringField("name").In())
	t.Log(StringField("name").NotIn("hello", "world"))
	t.Log(StringField("name").NotIn())
	t.Log(StringField("name").Like("a"))
	t.Log(StringField("name").NotLike("a"))
	t.Log(StringField("name").Prefix("a"))
	t.Log(StringField("name").Suffix("a"))

	t.Log(StringField("name").NotLike("a"))
	t.Log(StringField("name").NotPrefix("a"))
	t.Log(StringField("name").NotSuffix("a"))

}
