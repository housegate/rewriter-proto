package pb_test

import (
	"testing"

	pb "github.com/housegate/rewriter-proto/gen/pb"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestStorageIntegrityContractVersionValues pins the SI contract enum. V2 is
// the dynamic-table-set contract (housegate sub-project 3): DROP TABLE of a
// logical SI table succeeds, a V2 request activates the SI surface even when
// StorageIntegrityArgs.tables is empty, and reserved_databases joins the
// protected physical namespace.
func TestStorageIntegrityContractVersionValues(t *testing.T) {
	enum := pb.StorageIntegrityContractVersion(0).Descriptor()
	want := []protoreflect.Name{
		"STORAGE_INTEGRITY_CONTRACT_UNSPECIFIED",
		"STORAGE_INTEGRITY_CONTRACT_V1",
		"STORAGE_INTEGRITY_CONTRACT_V2",
	}
	if enum.Values().Len() != len(want) {
		t.Fatalf("%s has %d values, want %d", enum.FullName(), enum.Values().Len(), len(want))
	}
	for number, name := range want {
		value := enum.Values().ByNumber(protoreflect.EnumNumber(number))
		if value == nil || value.Name() != name {
			t.Fatalf("%s value %d = %v, want %s", enum.FullName(), number, value, name)
		}
	}
	if pb.StorageIntegrityContractVersion_STORAGE_INTEGRITY_CONTRACT_V2 != 2 {
		t.Fatalf("STORAGE_INTEGRITY_CONTRACT_V2 = %d, want 2", pb.StorageIntegrityContractVersion_STORAGE_INTEGRITY_CONTRACT_V2)
	}
}

// TestStorageIntegrityArgsReservedDatabases pins the V2 reserved-namespace
// field: repeated string reserved_databases = 5.
func TestStorageIntegrityArgsReservedDatabases(t *testing.T) {
	field := (&pb.StorageIntegrityArgs{}).ProtoReflect().Descriptor().Fields().ByName("reserved_databases")
	if field == nil || field.Number() != 5 || field.Kind() != protoreflect.StringKind || !field.IsList() {
		t.Fatalf("StorageIntegrityArgs.reserved_databases = %v, want repeated string field 5", field)
	}
	args := &pb.StorageIntegrityArgs{ReservedDatabases: []string{"hg_safe", "hg_unsafe", "hg_promote"}}
	if got := args.GetReservedDatabases(); len(got) != 3 || got[2] != "hg_promote" {
		t.Fatalf("GetReservedDatabases() = %v", got)
	}
}
