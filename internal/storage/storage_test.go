package storage

import (
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	sm := New()
	defer sm.Close()

	sm.Set("key1", []byte("value1"), 0)

	val, ok := sm.Get("key1")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if string(val) != "value1" {
		t.Fatalf("expected value1, got %s", string(val))
	}
}

func TestGetMissing(t *testing.T) {
	sm := New()
	defer sm.Close()

	_, ok := sm.Get("nonexistent")
	if ok {
		t.Fatal("expected key to not exist")
	}
}

func TestDel(t *testing.T) {
	sm := New()
	defer sm.Close()

	sm.Set("key1", []byte("value1"), 0)

	ok := sm.Del("key1")
	if !ok {
		t.Fatal("expected Del to return true for existing key")
	}

	_, exists := sm.Get("key1")
	if exists {
		t.Fatal("expected key to be deleted")
	}

	ok = sm.Del("nonexistent")
	if ok {
		t.Fatal("expected Del to return false for nonexistent key")
	}
}

func TestExists(t *testing.T) {
	sm := New()
	defer sm.Close()

	sm.Set("key1", []byte("value1"), 0)

	if !sm.Exists("key1") {
		t.Fatal("expected key to exist")
	}

	if sm.Exists("nonexistent") {
		t.Fatal("expected key to not exist")
	}
}

func TestIncr(t *testing.T) {
	sm := New()
	defer sm.Close()

	// Incr on nonexistent key should set to 1
	val, err := sm.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}

	// Incr again
	val, err = sm.Incr("counter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 2 {
		t.Fatalf("expected 2, got %d", val)
	}

	// Incr on non-integer value
	sm.Set("str", []byte("notanumber"), 0)
	_, err = sm.Incr("str")
	if err != ErrNotInteger {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestTTLExpiration(t *testing.T) {
	sm := New()
	defer sm.Close()

	sm.Set("expiring", []byte("value"), 50*time.Millisecond)

	val, ok := sm.Get("expiring")
	if !ok {
		t.Fatal("expected key to exist before expiration")
	}
	if string(val) != "value" {
		t.Fatalf("expected value, got %s", string(val))
	}

	time.Sleep(60 * time.Millisecond)

	_, ok = sm.Get("expiring")
	if ok {
		t.Fatal("expected key to be expired")
	}
}

func TestKeys(t *testing.T) {
	sm := New()
	defer sm.Close()

	sm.Set("key1", []byte("value1"), 0)
	sm.Set("key2", []byte("value2"), 0)
	sm.Set("key3", []byte("value3"), 0)

	keys := sm.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range []string{"key1", "key2", "key3"} {
		if !keyMap[expected] {
			t.Fatalf("expected key %s to be present", expected)
		}
	}
}

func TestWithShards(t *testing.T) {
	sm := New(WithShards(64))
	defer sm.Close()

	if sm.numShards != 64 {
		t.Fatalf("expected 64 shards, got %d", sm.numShards)
	}
}

func TestConcurrentAccess(t *testing.T) {
	sm := New()
	defer sm.Close()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOps := 1000

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "key"
				sm.Set(key, []byte("value"), 0)
				sm.Get(key)
				sm.Exists(key)
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkSet(b *testing.B) {
	sm := New()
	defer sm.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Set("key", []byte("value"), 0)
	}
}

func BenchmarkGet(b *testing.B) {
	sm := New()
	defer sm.Close()

	sm.Set("key", []byte("value"), 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Get("key")
	}
}

func BenchmarkConcurrentSet(b *testing.B) {
	sm := New()
	defer sm.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sm.Set("key", []byte("value"), 0)
		}
	})
}

func BenchmarkConcurrentGet(b *testing.B) {
	sm := New()
	defer sm.Close()

	sm.Set("key", []byte("value"), 0)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sm.Get("key")
		}
	})
}
