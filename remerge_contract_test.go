package json_test

import (
	"errors"
	"testing"

	json "github.com/goccy/go-json"
)

type standardJSONContract struct {
	Value string `json:"value"`
}

func (*standardJSONContract) UnmarshalJSON([]byte) error {
	return errors.New("standard UnmarshalJSON must not be called")
}

func (standardJSONContract) MarshalJSON() ([]byte, error) {
	return nil, errors.New("standard MarshalJSON must not be called")
}

type standardTextContract struct {
	Value string `json:"value"`
}

func (standardTextContract) MarshalText() ([]byte, error) {
	return nil, errors.New("standard MarshalText must not be called")
}

type remergeJSONContract struct {
	Called bool
}

func (v *remergeJSONContract) RemergeUnmarshalJSON([]byte) error {
	v.Called = true
	return nil
}

func TestRemergeInterfaceRoutingContract(t *testing.T) {
	var decoded standardJSONContract
	if err := json.Unmarshal([]byte(`{"value":"decoded"}`), &decoded); err != nil {
		t.Fatalf("Unmarshal returned an error: %v", err)
	}
	if decoded.Value != "decoded" {
		t.Fatalf("Unmarshal decoded Value = %q, want %q", decoded.Value, "decoded")
	}

	encoded, err := json.Marshal(standardJSONContract{Value: "encoded"})
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}
	if got, want := string(encoded), `{"value":"encoded"}`; got != want {
		t.Fatalf("Marshal output = %s, want %s", got, want)
	}

	encoded, err = json.Marshal(standardTextContract{Value: "encoded"})
	if err != nil {
		t.Fatalf("Marshal returned an error: %v", err)
	}
	if got, want := string(encoded), `{"value":"encoded"}`; got != want {
		t.Fatalf("Marshal output = %s, want %s", got, want)
	}

	var custom remergeJSONContract
	if err := json.Unmarshal([]byte(`{"value":"ignored"}`), &custom); err != nil {
		t.Fatalf("Unmarshal with RemergeUnmarshalJSON returned an error: %v", err)
	}
	if !custom.Called {
		t.Fatal("RemergeUnmarshalJSON was not called")
	}
}
