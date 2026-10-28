package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/VkHyperNova/vkmount"
)

// setup redirects the storage paths to a temp dir and stubs the drive mount.
// Everything is restored when the test ends.
func setup(t *testing.T, mount func() error) (local, backup string) {
	t.Helper()

	dir := t.TempDir()
	local = filepath.Join(dir, "local", "FINANCE", "finance.json")
	backup = filepath.Join(dir, "drive", "FINANCE", "finance.json")

	oldLocal, oldBackup, oldMount := LocalFile, BackupFile, mountBackup
	LocalFile, BackupFile, mountBackup = local, backup, mount
	t.Cleanup(func() {
		LocalFile, BackupFile, mountBackup = oldLocal, oldBackup, oldMount
	})

	return local, backup
}

func sample() *Finance {
	return &Finance{Finance: []Item{
		{DATE: "28-10-2026", TIME: "10:00:00", COMMENT: "salary", VALUE: 1500},
		{DATE: "28-10-2026", TIME: "12:30:00", COMMENT: "coffee", VALUE: -4.5},
	}}
}

func TestLoadMissingFile(t *testing.T) {
	f := &Finance{}
	err := f.Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should not be an error, got: %v", err)
	}
	if len(f.Finance) != 0 {
		t.Fatalf("expected empty data, got %d items", len(f.Finance))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := &Finance{}
	if err := f.Load(path); err == nil {
		t.Fatal("corrupted file must return an error, not be treated as empty")
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	local, _ := setup(t, func() error { return nil })

	want := sample()
	if err := want.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got := &Finance{}
	if err := got.Load(local); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(want.Finance, got.Finance) {
		t.Fatalf("round trip mismatch:\nwant %+v\ngot  %+v", want.Finance, got.Finance)
	}
}

func TestSaveCreatesLocalFolder(t *testing.T) {
	local, _ := setup(t, func() error { return nil })

	if err := sample().Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(local); err != nil {
		t.Fatalf("local file was not created: %v", err)
	}
}

func TestSaveWritesBackupWhenMounted(t *testing.T) {
	local, backup := setup(t, func() error { return nil })

	if err := sample().Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	localData, err := os.ReadFile(local)
	if err != nil {
		t.Fatal(err)
	}
	backupData, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("backup file was not written: %v", err)
	}
	if string(localData) != string(backupData) {
		t.Fatal("local and backup contents differ")
	}
}

func TestSaveSkipsBackupButSucceeds(t *testing.T) {
	cases := map[string]error{
		"user skipped":  vkmount.ErrSkipped,
		"mount failure": errors.New("udisksctl failed"),
	}

	for name, mountErr := range cases {
		t.Run(name, func(t *testing.T) {
			local, backup := setup(t, func() error { return mountErr })

			if err := sample().Save(); err != nil {
				t.Fatalf("Save must not fail when the backup drive is unavailable: %v", err)
			}
			if _, err := os.Stat(local); err != nil {
				t.Fatalf("local file must still be written: %v", err)
			}
			if _, err := os.Stat(backup); err == nil {
				t.Fatal("backup file must not be written when mounting failed")
			}
		})
	}
}
