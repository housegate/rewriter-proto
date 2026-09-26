package pb

import "testing"

func TestProtectedDatabasesFieldRoundTrips(t *testing.T) {
	in := &RewriteTableDynamicArgs{ProtectedDatabases: []string{"phys", "hg_safe"}}
	if got := in.GetProtectedDatabases(); len(got) != 2 || got[0] != "phys" || got[1] != "hg_safe" {
		t.Fatalf("protected_databases = %v, want [phys hg_safe]", got)
	}
}
