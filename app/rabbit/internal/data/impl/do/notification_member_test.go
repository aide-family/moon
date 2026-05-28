package do

import (
	"encoding/json"
	"testing"
)

func TestNotificationMember_UnmarshalJSON_camelCase(t *testing.T) {
	var member NotificationMember
	if err := json.Unmarshal([]byte(`{"memberUid":42,"isEmail":true}`), &member); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if member.MemberUID != 42 || !member.IsEmail {
		t.Fatalf("got memberUID=%d isEmail=%v", member.MemberUID, member.IsEmail)
	}
}

func TestNotificationMembers_Scan_camelCase(t *testing.T) {
	var members NotificationMembers
	if err := members.Scan([]byte(`[{"memberUid":7,"isEmail":true}]`)); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(members) != 1 || members[0].MemberUID != 7 || !members[0].IsEmail {
		t.Fatalf("got %+v", members)
	}
}
