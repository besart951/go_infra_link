package postgresjson

import "testing"

func TestDocumentReturnsNilForOptionalEmptyJSON(t *testing.T) {
	value, err := Document(nil).Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != nil {
		t.Fatalf("value = %#v, want nil", value)
	}
}

func TestDocumentRejectsInvalidJSON(t *testing.T) {
	if _, err := Document(`not json`).Value(); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
}

func TestDocumentScansAndCopiesJSON(t *testing.T) {
	var document Document
	if err := document.Scan([]byte(`{"value":42}`)); err != nil {
		t.Fatal(err)
	}
	if string(document.Bytes()) != `{"value":42}` {
		t.Fatalf("document = %s", document)
	}
}
