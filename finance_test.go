package main

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"vk-finance"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCalculate(t *testing.T) {
	cases := []struct {
		name                      string
		items                     []Item
		income, expenses, balance float64
	}{
		{"empty", nil, 0, 0, 0},
		{"income only", []Item{{VALUE: 100}, {VALUE: 50.5}}, 150.5, 0, 150.5},
		{"expenses only", []Item{{VALUE: -20}, {VALUE: -5.25}}, 0, -25.25, -25.25},
		{"mixed", sample().Finance, 1500, -4.5, 1495.5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &Finance{Finance: c.items}
			income, expenses, balance := f.calculate()
			if !near(income, c.income) || !near(expenses, c.expenses) || !near(balance, c.balance) {
				t.Fatalf("got %v/%v/%v, want %v/%v/%v",
					income, expenses, balance, c.income, c.expenses, c.balance)
			}
		})
	}
}

func TestAddNormalizesAndSaves(t *testing.T) {
	local, _ := setup(t, func() error { return nil })

	f := &Finance{}
	if err := f.Add("COFFEE Beans", -4.5); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if len(f.Finance) != 1 {
		t.Fatalf("expected 1 item, got %d", len(f.Finance))
	}
	item := f.Finance[0]
	if item.COMMENT != "coffee beans" {
		t.Errorf("comment should be lowercased, got %q", item.COMMENT)
	}
	if !near(item.VALUE, -4.5) {
		t.Errorf("value = %v, want -4.5", item.VALUE)
	}
	if _, err := time.Parse("02-01-2006", item.DATE); err != nil {
		t.Errorf("date %q is not DD-MM-YYYY: %v", item.DATE, err)
	}
	if _, err := time.Parse("15:04:05", item.TIME); err != nil {
		t.Errorf("time %q is not HH:MM:SS: %v", item.TIME, err)
	}

	// It must also be on disk
	loaded := &Finance{}
	if err := loaded.Load(local); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Finance) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(loaded.Finance))
	}
}

func TestAddAppendsToExisting(t *testing.T) {
	setup(t, func() error { return nil })

	f := sample()
	if err := f.Add("bread", -2); err != nil {
		t.Fatal(err)
	}
	if len(f.Finance) != 3 || f.Finance[2].COMMENT != "bread" {
		t.Fatalf("new item should be appended last, got %+v", f.Finance)
	}
}

func TestUndoEmpty(t *testing.T) {
	setup(t, func() error { return nil })

	if err := (&Finance{}).Undo(); err == nil {
		t.Fatal("undo on empty data should return an error")
	}
}

func TestUndoRemovesLastAndSaves(t *testing.T) {
	local, _ := setup(t, func() error { return nil })

	f := sample()
	if err := f.Undo(); err != nil {
		t.Fatal(err)
	}
	if len(f.Finance) != 1 || f.Finance[0].COMMENT != "salary" {
		t.Fatalf("expected only salary left, got %+v", f.Finance)
	}

	loaded := &Finance{}
	if err := loaded.Load(local); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Finance) != 1 {
		t.Fatalf("undo was not saved, file has %d items", len(loaded.Finance))
	}
}

func TestLog(t *testing.T) {
	t.Run("too few arguments", func(t *testing.T) {
		setup(t, func() error { return nil })
		withArgs(t, "log", "coffee")

		if err := (&Finance{}).Log(); err == nil {
			t.Fatal("expected a usage error")
		}
	})

	t.Run("invalid sum", func(t *testing.T) {
		setup(t, func() error { return nil })
		withArgs(t, "log", "coffee", "abc")

		f := &Finance{}
		if err := f.Log(); err == nil {
			t.Fatal("expected an error for a non-numeric sum")
		}
		if len(f.Finance) != 0 {
			t.Fatal("nothing should be added on error")
		}
	})

	t.Run("valid", func(t *testing.T) {
		setup(t, func() error { return nil })
		withArgs(t, "log", "Coffee", "-4.5")

		f := &Finance{}
		if err := f.Log(); err != nil {
			t.Fatal(err)
		}
		if len(f.Finance) != 1 || f.Finance[0].COMMENT != "coffee" || !near(f.Finance[0].VALUE, -4.5) {
			t.Fatalf("unexpected result: %+v", f.Finance)
		}
	})
}

/* Backup tests — these need the prompt to be replaceable, see below */

func setupBackup(t *testing.T, confirm bool) (monthly string) {
	t.Helper()
	setup(t, func() error { return nil })

	monthly = filepath.Join(t.TempDir(), "drive", "FINANCE", "September2026.json")

	oldFile, oldConfirm := BackupFileWithDate, confirmBackup
	BackupFileWithDate = monthly
	confirmBackup = func() bool { return confirm }
	t.Cleanup(func() { BackupFileWithDate, confirmBackup = oldFile, oldConfirm })

	return monthly
}

func TestBackupCanceled(t *testing.T) {
	monthly := setupBackup(t, false)

	f := sample()
	if err := f.Backup(); err == nil {
		t.Fatal("expected a cancel error")
	}
	if _, err := os.Stat(monthly); err == nil {
		t.Fatal("no backup file should be written when canceled")
	}
	if len(f.Finance) != 2 {
		t.Fatal("data must stay untouched when canceled")
	}
}

func TestBackupPositiveBalanceResets(t *testing.T) {
	monthly := setupBackup(t, true)

	f := sample() // balance +1495.5
	if err := f.Backup(); err != nil {
		t.Fatal(err)
	}

	saved := &Finance{}
	if err := saved.Load(monthly); err != nil {
		t.Fatal(err)
	}
	if len(saved.Finance) != 2 {
		t.Fatalf("backup should hold the old 2 items, got %d", len(saved.Finance))
	}
	if len(f.Finance) != 0 {
		t.Fatalf("data should be reset to empty, got %d items", len(f.Finance))
	}
}

func TestBackupNegativeBalanceCarriesDebt(t *testing.T) {
	setupBackup(t, true)

	f := &Finance{Finance: []Item{{COMMENT: "rent", VALUE: -300}, {COMMENT: "pay", VALUE: 100}}}
	if err := f.Backup(); err != nil {
		t.Fatal(err)
	}
	if len(f.Finance) != 1 || f.Finance[0].COMMENT != "debt" || !near(f.Finance[0].VALUE, -200) {
		t.Fatalf("expected a single debt of -200, got %+v", f.Finance)
	}
}

func TestBackupNeverOverwrites(t *testing.T) {
	monthly := setupBackup(t, true)

	if err := os.MkdirAll(filepath.Dir(monthly), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(monthly, []byte("old backup"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := sample()
	if err := f.Backup(); err == nil {
		t.Fatal("expected an error because the monthly backup already exists")
	}
	got, _ := os.ReadFile(monthly)
	if string(got) != "old backup" {
		t.Fatal("existing backup was overwritten")
	}
	if len(f.Finance) != 2 {
		t.Fatal("data must not be reset when the backup failed")
	}
}

func TestBackupDriveUnavailable(t *testing.T) {
	setupBackup(t, true)
	mountBackup = func() error { return errors.New("no drive") }

	f := sample()
	if err := f.Backup(); err == nil {
		t.Fatal("expected an error when the drive can't be mounted")
	}
	if len(f.Finance) != 2 {
		t.Fatal("data must not be reset when the backup failed")
	}
}
