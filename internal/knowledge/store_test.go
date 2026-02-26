package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Fatal("Database not initialized")
	}
	if store.conn == nil {
		t.Fatal("Connection not initialized")
	}
	if store.path != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, store.path)
	}
}

func TestStoreSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}
	defer store.Close()

	// Verify Entity table exists by trying to query it
	result, err := store.query("MATCH (e:Entity) RETURN count(e)")
	if err != nil {
		t.Fatalf("Entity table not created: %v", err)
	}
	defer result.Close()

	// Verify Observation table exists
	result, err = store.query("MATCH (o:Observation) RETURN count(o)")
	if err != nil {
		t.Fatalf("Observation table not created: %v", err)
	}
	defer result.Close()
}

func TestStoreClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestStoreCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

	// Parent directory should be created automatically
	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("OpenStore failed: %v", err)
	}
	defer store.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Parent directory was not created")
	}
}
