package domain

import (
	"testing"
	"time"
)

func TestComputeUniquenessKey(t *testing.T) {
	date := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)

	// Same inputs should produce same key
	key1 := ComputeUniquenessKey(date, "AAPL", "Trade", 100.5, 5000.0)
	key2 := ComputeUniquenessKey(date, "AAPL", "Trade", 100.5, 5000.0)
	if key1 != key2 {
		t.Errorf("same inputs should produce same key, got %s and %s", key1, key2)
	}

	// Different quantity should produce different key
	key3 := ComputeUniquenessKey(date, "AAPL", "Trade", 100.6, 5000.0)
	if key1 == key3 {
		t.Errorf("different quantity should produce different key")
	}

	// Different amount should produce different key
	key4 := ComputeUniquenessKey(date, "AAPL", "Trade", 100.5, 5001.0)
	if key1 == key4 {
		t.Errorf("different amount should produce different key")
	}

	// Different instrument should produce different key
	key5 := ComputeUniquenessKey(date, "GOOGL", "Trade", 100.5, 5000.0)
	if key1 == key5 {
		t.Errorf("different instrument should produce different key")
	}

	// Different category should produce different key
	key6 := ComputeUniquenessKey(date, "AAPL", "Corporate Action", 100.5, 5000.0)
	if key1 == key6 {
		t.Errorf("different category should produce different key")
	}

	// Different date should produce different key
	date2 := time.Date(2024, time.January, 16, 0, 0, 0, 0, time.UTC)
	key7 := ComputeUniquenessKey(date2, "AAPL", "Trade", 100.5, 5000.0)
	if key1 == key7 {
		t.Errorf("different date should produce different key")
	}

	// High precision floats should be distinguished
	key8 := ComputeUniquenessKey(date, "AAPL", "Trade", 100.500001, 5000.0)
	if key1 == key8 {
		t.Errorf("similar quantity values should produce different keys")
	}
}
