package egress

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func NewDummyRPCRequest() *RPCRequest {
	return &RPCRequest{
		ModuleName: "dummyModule",
		MethodName: "dummyMethod",
		Params:     append(make([]any, 0), 1, 2),
		ID:         1,
		Method:     "dummyModule_dummyMethod",
		JSONRPC:    "2.0",
	}
}

func TestRPCRequest_GetRequestKeyIntParams(t *testing.T) {
	r := NewDummyRPCRequest()
	want := "dummyModule_dummyMethod12"
	actual := r.GetRequestKey()
	if actual != want {
		t.Fatal("invalid request key:", want, actual)
	}
}

func TestRPCRequest_GetRequestKeyDifferentTypeParams(t *testing.T) {
	r := NewDummyRPCRequest()
	r.Params = append(make([]any, 0), 1, "2")
	want := "dummyModule_dummyMethod1\"2\""
	actual := r.GetRequestKey()
	if actual != want {
		t.Fatal("invalid request key:", want, actual)
	}
}

func TestRPCRequest_GetRequestKeyEmptyParams(t *testing.T) {
	r := NewDummyRPCRequest()
	r.Params = make([]any, 0)
	want := "dummyModule_dummyMethod"
	actual := r.GetRequestKey()
	if actual != want {
		t.Fatal("invalid request key:", want, actual)
	}
}

func TestRPCRequest_GetFullMethodName(t *testing.T) {
	r := NewDummyRPCRequest()
	want := "dummyModule_dummyMethod"
	actual := r.GetFullMethodName()
	if actual != want {
		t.Fatal("invalid full method name:", want, actual)
	}
}

func TestParseCall(t *testing.T) {
	data := []byte(`{"id": 1, "jsonrpc": "2.0", "method": "dummyModule_dummyMethod", "params": [1,2]}`)
	want := NewDummyRPCRequest()
	actual, err := ParseCall(data)
	if err != nil {
		t.Fatal("error in parse call when there shouldn't be any", err)
	}
	if cmp.Equal(*want, *actual) {
		t.Fatal("invalid parse call result", want, actual)
	}
}

func TestParseCallInvalidRequests(t *testing.T) {
	cases := []string{
		`{"jsonrpc": "2.0", "method": "dummyModule_dummyMethod", "params": [1,2]}`,
		`{"id": null, "method": "dummyModule_dummyMethod", "params": [1,2]}`,
		`{"id": 1, "jsonrpc": "1.0", method": "dummyModule_dummyMethod", "params": [1,2]}`,
		`{"id": 1, "jsonrpc": "2.0", method": "dummyModuledummyMethod", "params": [1,2]}`,
	}
	for _, badCase := range cases {
		_, err := ParseCall([]byte(badCase))
		if err == nil {
			t.Fatal("no error in parse call when there should be one", badCase)
		}
	}
}
