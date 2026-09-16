package pb_test

import (
	"context"
	"net"
	"reflect"
	"testing"

	pb "github.com/housegate/rewriter-proto/gen/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldContract struct {
	name        protoreflect.Name
	number      protoreflect.FieldNumber
	kind        protoreflect.Kind
	cardinality protoreflect.Cardinality
	typeName    protoreflect.FullName
}

func assertMessageContract(t *testing.T, message proto.Message, want []fieldContract) {
	t.Helper()
	fields := message.ProtoReflect().Descriptor().Fields()
	if fields.Len() != len(want) {
		t.Fatalf("%s field count = %d, want %d", message.ProtoReflect().Descriptor().FullName(), fields.Len(), len(want))
	}
	for index, expected := range want {
		field := fields.Get(index)
		if field.Name() != expected.name || field.Number() != expected.number || field.Kind() != expected.kind || field.Cardinality() != expected.cardinality {
			t.Errorf("%s field[%d] = (%s, %d, %s, %s), want (%s, %d, %s, %s)",
				message.ProtoReflect().Descriptor().FullName(), index,
				field.Name(), field.Number(), field.Kind(), field.Cardinality(),
				expected.name, expected.number, expected.kind, expected.cardinality)
		}
		var gotType protoreflect.FullName
		switch field.Kind() {
		case protoreflect.MessageKind:
			gotType = field.Message().FullName()
		case protoreflect.EnumKind:
			gotType = field.Enum().FullName()
		}
		if gotType != expected.typeName {
			t.Errorf("%s.%s type name = %q, want %q", message.ProtoReflect().Descriptor().FullName(), field.Name(), gotType, expected.typeName)
		}
	}
}

func TestSnapshotQueryMessageContracts(t *testing.T) {
	optional := protoreflect.Optional
	repeated := protoreflect.Repeated
	assertMessageContract(t, &pb.SnapshotQueryColumn{}, []fieldContract{
		{"name", 1, protoreflect.StringKind, optional, ""},
		{"type", 2, protoreflect.StringKind, optional, ""},
	})
	assertMessageContract(t, &pb.SnapshotQueryCatalogTable{}, []fieldContract{
		{"database", 1, protoreflect.StringKind, optional, ""},
		{"table", 2, protoreflect.StringKind, optional, ""},
		{"table_id", 3, protoreflect.StringKind, optional, ""},
		{"schema_hash", 4, protoreflect.StringKind, optional, ""},
		{"columns", 5, protoreflect.MessageKind, repeated, "rewriter.SnapshotQueryColumn"},
	})
	assertMessageContract(t, &pb.AnalyzeSnapshotQueryRequest{}, []fieldContract{
		{"contract_version", 1, protoreflect.Uint32Kind, optional, ""},
		{"query_profile_id", 2, protoreflect.StringKind, optional, ""},
		{"sql", 3, protoreflect.StringKind, optional, ""},
		{"logical_database", 4, protoreflect.StringKind, optional, ""},
		{"catalog", 5, protoreflect.MessageKind, repeated, "rewriter.SnapshotQueryCatalogTable"},
		{"materialize", 6, protoreflect.BoolKind, optional, ""},
		{"inputs", 7, protoreflect.MessageKind, optional, "rewriter.MaterializationInputs"},
	})
	assertMessageContract(t, &pb.AnalyzeSnapshotQueryResponse{}, []fieldContract{
		{"contract_version", 1, protoreflect.Uint32Kind, optional, ""},
		{"query_profile_id", 2, protoreflect.StringKind, optional, ""},
		{"code", 3, protoreflect.EnumKind, optional, "rewriter.SnapshotQueryCode"},
		{"message", 4, protoreflect.StringKind, optional, ""},
		{"sql_after_materialization", 5, protoreflect.StringKind, optional, ""},
		{"target_table_id", 6, protoreflect.StringKind, optional, ""},
		{"target_columns", 7, protoreflect.StringKind, repeated, ""},
		{"read_table_ids", 8, protoreflect.StringKind, repeated, ""},
	})
	assertMessageContract(t, &pb.SnapshotScratchBinding{}, []fieldContract{
		{"table_id", 1, protoreflect.StringKind, optional, ""},
		{"scratch_database", 2, protoreflect.StringKind, optional, ""},
		{"scratch_table", 3, protoreflect.StringKind, optional, ""},
	})
	assertMessageContract(t, &pb.PrepareSnapshotQueryRequest{}, []fieldContract{
		{"analysis", 1, protoreflect.MessageKind, optional, "rewriter.AnalyzeSnapshotQueryRequest"},
		{"bindings", 2, protoreflect.MessageKind, repeated, "rewriter.SnapshotScratchBinding"},
	})
	assertMessageContract(t, &pb.PrepareSnapshotQueryResponse{}, []fieldContract{
		{"contract_version", 1, protoreflect.Uint32Kind, optional, ""},
		{"query_profile_id", 2, protoreflect.StringKind, optional, ""},
		{"code", 3, protoreflect.EnumKind, optional, "rewriter.SnapshotQueryCode"},
		{"message", 4, protoreflect.StringKind, optional, ""},
		{"select_sql", 5, protoreflect.StringKind, optional, ""},
		{"target_table_id", 6, protoreflect.StringKind, optional, ""},
		{"target_columns", 7, protoreflect.StringKind, repeated, ""},
		{"read_table_ids", 8, protoreflect.StringKind, repeated, ""},
	})
}

func analysisFixture() *pb.AnalyzeSnapshotQueryRequest {
	now := int64(1700000000123456789)
	return &pb.AnalyzeSnapshotQueryRequest{
		ContractVersion: 1,
		QueryProfileId:  "snapshot-profile-v1",
		Sql:             "INSERT INTO tenant.copy SELECT value FROM tenant.events",
		LogicalDatabase: "tenant",
		Catalog: []*pb.SnapshotQueryCatalogTable{
			{
				Database:   "tenant",
				Table:      "copy",
				TableId:    "table-copy-id",
				SchemaHash: "copy-schema-sha256",
				Columns: []*pb.SnapshotQueryColumn{
					{Name: "copy_a", Type: "UInt64"},
					{Name: "copy_b", Type: "String"},
				},
			},
			{
				Database:   "tenant",
				Table:      "events",
				TableId:    "table-events-id",
				SchemaHash: "events-schema-sha256",
				Columns: []*pb.SnapshotQueryColumn{
					{Name: "event_value", Type: "Int32"},
					{Name: "event_time", Type: "DateTime64(3)"},
				},
			},
		},
		Materialize: true,
		Inputs: &pb.MaterializationInputs{
			NowUnixNs:           &now,
			RandomUint64Values:  []uint64{11, 22},
			UuidValues:          []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"},
			RandomFloat64Values: []float64{0.125, 0.875},
		},
	}
}

func roundTrip(t *testing.T, original, decoded proto.Message) {
	t.Helper()
	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(wire, decoded); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(original, decoded) {
		t.Fatalf("round trip mismatch:\noriginal: %v\ndecoded: %v", original, decoded)
	}
}

func TestSnapshotQueryMessagesRoundTripEveryField(t *testing.T) {
	analysis := analysisFixture()
	decodedAnalysis := &pb.AnalyzeSnapshotQueryRequest{}
	roundTrip(t, analysis, decodedAnalysis)
	if !decodedAnalysis.ProtoReflect().Has(decodedAnalysis.ProtoReflect().Descriptor().Fields().ByName("inputs")) || decodedAnalysis.GetInputs().GetNowUnixNs() != 1700000000123456789 {
		t.Fatalf("materialization inputs or optional time presence lost: %v", decodedAnalysis.GetInputs())
	}
	if got := []string{decodedAnalysis.Catalog[0].Columns[0].Name, decodedAnalysis.Catalog[0].Columns[1].Name}; !reflect.DeepEqual(got, []string{"copy_a", "copy_b"}) {
		t.Fatalf("catalog schema order = %v", got)
	}
	if got := decodedAnalysis.Inputs.RandomUint64Values; !reflect.DeepEqual(got, []uint64{11, 22}) {
		t.Fatalf("random input order = %v", got)
	}
	if got := decodedAnalysis.Inputs.UuidValues; !reflect.DeepEqual(got, []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}) {
		t.Fatalf("uuid input order = %v", got)
	}
	if got := decodedAnalysis.Inputs.RandomFloat64Values; !reflect.DeepEqual(got, []float64{0.125, 0.875}) {
		t.Fatalf("float input order = %v", got)
	}

	analysisResponse := &pb.AnalyzeSnapshotQueryResponse{
		ContractVersion:         1,
		QueryProfileId:          "snapshot-profile-v1",
		Code:                    pb.SnapshotQueryCode_SUCCESS,
		Message:                 "analysis-message",
		SqlAfterMaterialization: "INSERT INTO tenant.copy (copy_b, copy_a) SELECT event_value, 7 FROM tenant.events",
		TargetTableId:           "table-copy-id",
		TargetColumns:           []string{"copy_b", "copy_a"},
		ReadTableIds:            []string{"read-a-id", "read-b-id"},
	}
	decodedAnalysisResponse := &pb.AnalyzeSnapshotQueryResponse{}
	roundTrip(t, analysisResponse, decodedAnalysisResponse)
	if !reflect.DeepEqual(decodedAnalysisResponse.TargetColumns, []string{"copy_b", "copy_a"}) || !reflect.DeepEqual(decodedAnalysisResponse.ReadTableIds, []string{"read-a-id", "read-b-id"}) {
		t.Fatalf("analysis response order lost: columns=%v reads=%v", decodedAnalysisResponse.TargetColumns, decodedAnalysisResponse.ReadTableIds)
	}

	prepare := &pb.PrepareSnapshotQueryRequest{
		Analysis: analysis,
		Bindings: []*pb.SnapshotScratchBinding{
			{TableId: "read-a-id", ScratchDatabase: "scratch_db_a", ScratchTable: "scratch_table_a"},
			{TableId: "read-b-id", ScratchDatabase: "scratch_db_b", ScratchTable: "scratch_table_b"},
		},
	}
	decodedPrepare := &pb.PrepareSnapshotQueryRequest{}
	roundTrip(t, prepare, decodedPrepare)
	if got := []string{decodedPrepare.Bindings[0].TableId, decodedPrepare.Bindings[1].TableId}; !reflect.DeepEqual(got, []string{"read-a-id", "read-b-id"}) {
		t.Fatalf("binding order = %v", got)
	}
	if decodedPrepare.Analysis.GetInputs().GetNowUnixNs() != 1700000000123456789 {
		t.Fatalf("nested analysis inputs lost: %v", decodedPrepare.Analysis.GetInputs())
	}

	prepareResponse := &pb.PrepareSnapshotQueryResponse{
		ContractVersion: 1,
		QueryProfileId:  "snapshot-profile-v1",
		Code:            pb.SnapshotQueryCode_UNSUPPORTED,
		Message:         "prepare-message",
		SelectSql:       "SELECT event_value AS copy_a, 7 AS copy_b FROM scratch_db_a.scratch_table_a",
		TargetTableId:   "table-copy-id",
		TargetColumns:   []string{"copy_a", "copy_b"},
		ReadTableIds:    []string{"read-a-id", "read-b-id"},
	}
	decodedPrepareResponse := &pb.PrepareSnapshotQueryResponse{}
	roundTrip(t, prepareResponse, decodedPrepareResponse)
	if !reflect.DeepEqual(decodedPrepareResponse.TargetColumns, []string{"copy_a", "copy_b"}) || !reflect.DeepEqual(decodedPrepareResponse.ReadTableIds, []string{"read-a-id", "read-b-id"}) {
		t.Fatalf("prepare response order lost: columns=%v reads=%v", decodedPrepareResponse.TargetColumns, decodedPrepareResponse.ReadTableIds)
	}
}

func TestSnapshotQueryCodeValuesAreFrozen(t *testing.T) {
	want := map[pb.SnapshotQueryCode]int32{
		pb.SnapshotQueryCode_UNSPECIFIED:            0,
		pb.SnapshotQueryCode_SUCCESS:                1,
		pb.SnapshotQueryCode_UNSUPPORTED:            2,
		pb.SnapshotQueryCode_INVALID_INPUT:          3,
		pb.SnapshotQueryCode_PROFILE_UNAVAILABLE:    4,
		pb.SnapshotQueryCode_MATERIALIZATION_FAILED: 5,
		pb.SnapshotQueryCode_NOT_SNAPSHOT_QUERY:     6,
	}
	for value, number := range want {
		if int32(value) != number {
			t.Errorf("%s = %d, want %d", value, value, number)
		}
	}
}

type compatibilityServer struct {
	pb.UnimplementedRewriterServiceServer
}

func (compatibilityServer) Rewrite(_ context.Context, request *pb.RewriteSQLRequest) (*pb.RewriteSQLResponse, error) {
	return &pb.RewriteSQLResponse{Code: pb.RewriteCode_Success, SqlAfterRewrite: request.Sql}, nil
}

func TestNewRPCsReturnTransportUnimplementedAndLegacyRewriteStillWorks(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterRewriterServiceServer(server, compatibilityServer{})
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	connection, err := grpc.NewClient("passthrough:///bufconn", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewRewriterServiceClient(connection)

	for name, call := range map[string]func() (proto.Message, error){
		"AnalyzeSnapshotQuery": func() (proto.Message, error) {
			return client.AnalyzeSnapshotQuery(context.Background(), analysisFixture())
		},
		"PrepareSnapshotQuery": func() (proto.Message, error) {
			return client.PrepareSnapshotQuery(context.Background(), &pb.PrepareSnapshotQueryRequest{Analysis: analysisFixture()})
		},
	} {
		t.Run(name, func(t *testing.T) {
			response, err := call()
			if status.Code(err) != codes.Unimplemented {
				t.Fatalf("status = %v, response = %v, want transport Unimplemented and nil response", status.Code(err), response)
			}
			if response != nil && !reflect.ValueOf(response).IsNil() {
				t.Fatalf("response = %v, want nil on transport Unimplemented", response)
			}
		})
	}

	legacy, err := client.Rewrite(context.Background(), &pb.RewriteSQLRequest{Sql: "SELECT 153"})
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Code != pb.RewriteCode_Success || legacy.SqlAfterRewrite != "SELECT 153" {
		t.Fatalf("legacy Rewrite response = %v", legacy)
	}
}

func TestServiceMethodsRemainAppendOnly(t *testing.T) {
	methods := pb.File_rewriter_proto.Services().ByName("RewriterService").Methods()
	want := []string{"MaterializeSQL", "Rewrite", "RewriteErrorMessage", "Optimize", "AnalyzeSnapshotQuery", "PrepareSnapshotQuery"}
	if methods.Len() != len(want) {
		t.Fatalf("method count = %d, want %d", methods.Len(), len(want))
	}
	for index, name := range want {
		if got := string(methods.Get(index).Name()); got != name {
			t.Errorf("method[%d] = %q, want %q", index, got, name)
		}
	}
}

func TestLegacyMessageFieldAndServiceSignaturesRemainCompatible(t *testing.T) {
	materializeFields := (&pb.MaterializeSQLRequest{}).ProtoReflect().Descriptor().Fields()
	inputs := materializeFields.ByName("inputs")
	if inputs == nil || inputs.Number() != 2 || inputs.Kind() != protoreflect.MessageKind || inputs.Message().FullName() != "rewriter.MaterializationInputs" {
		t.Fatalf("legacy MaterializeSQLRequest.inputs changed: %v", inputs)
	}

	methods := pb.File_rewriter_proto.Services().ByName("RewriterService").Methods()
	want := []struct {
		name   protoreflect.Name
		input  protoreflect.FullName
		output protoreflect.FullName
	}{
		{"MaterializeSQL", "rewriter.MaterializeSQLRequest", "rewriter.MaterializeSQLResponse"},
		{"Rewrite", "rewriter.RewriteSQLRequest", "rewriter.RewriteSQLResponse"},
		{"RewriteErrorMessage", "rewriter.RewriteErrorMessageRequest", "rewriter.RewriteErrorMessageResponse"},
		{"Optimize", "rewriter.OptimizeRequest", "rewriter.OptimizeResponse"},
	}
	for index, expected := range want {
		method := methods.Get(index)
		if method.Name() != expected.name || method.Input().FullName() != expected.input || method.Output().FullName() != expected.output || method.IsStreamingClient() || method.IsStreamingServer() {
			t.Errorf("legacy method[%d] = %s(%s) returns (%s), streaming=(%t,%t); want %s(%s) returns (%s), unary",
				index, method.Name(), method.Input().FullName(), method.Output().FullName(), method.IsStreamingClient(), method.IsStreamingServer(),
				expected.name, expected.input, expected.output)
		}
	}
}
