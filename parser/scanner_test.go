package parser

import (
	"testing"
)

func TestCreateMakefileScanner(t *testing.T) {
	_, err := NewMakefileScanner("../fixtures/simple.make")
	if err != nil {
		t.Errorf("Unable to create MakefileScanner for 'fixtures/simple.make': %s", err.Error())
	}
}

func TestCreateMakefileScannerFailing(t *testing.T) {
	_, err := NewMakefileScanner("fixtures/idontexist.make")

	if err == nil {
		t.Errorf("Unable to fail creating MakefileScanner for 'fixtures/idontexist.make'")
	}
}
