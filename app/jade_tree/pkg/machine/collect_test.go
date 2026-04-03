package machine

import "testing"

func TestLocalMachineIdentityMatchesMachineUUID(t *testing.T) {
	u, _, _ := LocalMachineIdentity()
	if u != MachineUUID() {
		t.Fatalf("LocalMachineIdentity machine UUID %q != MachineUUID() %q", u, MachineUUID())
	}
}
