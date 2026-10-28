package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/peterh/liner"
)

type Item struct {
	DATE    string  `json:"date"`
	TIME    string  `json:"time"`
	COMMENT string  `json:"comment"`
	VALUE   float64 `json:"value"`
}

type Finance struct {
	Finance []Item `json:"finance"`
}

/* Main Commands */

func (f *Finance) Log() error {
	if len(os.Args) < 4 {
		return errors.New("usage: vk-finance log <item> <sum>")
	}

	sum, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		return fmt.Errorf("sum must be a valid number: %w", err)
	}

	return f.Add(os.Args[2], sum)
}

func (f *Finance) Add(item string, value float64) error {

	comment := strings.ToLower(item)

	now := time.Now()

	NewItem := Item{
		DATE:    now.Format("02-01-2006"),
		TIME:    now.Format("15:04:05"),
		COMMENT: comment,
		VALUE:   value,
	}

	f.Finance = append(f.Finance, NewItem)

	return f.Save()
}

func (f *Finance) Backup() error {
	if !confirmBackup() {
		return errors.New("backup canceled")
	}

	// Serialize current state
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling finance data: %w", err)
	}

	if err := mountBackup(); err != nil {
		return fmt.Errorf("backup drive: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(BackupFileWithDate), 0o755); err != nil {
		return fmt.Errorf("creating backup folder: %w", err)
	}

	// Never overwrite an existing monthly backup
	if _, err := os.Stat(BackupFileWithDate); err == nil {
		return fmt.Errorf("%s already exists, not overwriting", BackupFileWithDate)
	}

	if err := os.WriteFile(BackupFileWithDate, data, 0o644); err != nil {
		return fmt.Errorf("writing backup file: %w", err)
	}

	// Capture balance BEFORE resetting
	_, _, oldBalance := f.calculate()

	// Reset to an empty DB (only reached if the backup was written)
	f.Finance = []Item{}

	// Carry over debt, otherwise just save the empty state
	if oldBalance < 0 {
		return f.Add("debt", oldBalance) // Add saves
	}
	return f.Save()
}

func (f *Finance) Undo() error {
	if len(f.Finance) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	f.Finance = f.Finance[:len(f.Finance)-1]
	return f.Save()
}

/* Helper Functions */

func (f *Finance) calculate() (float64, float64, float64) {

	income := 0.0
	expenses := 0.0

	for _, item := range f.Finance {

		if item.VALUE < 0 {
			expenses += item.VALUE
		} else {
			income += item.VALUE
		}
	}

	return income, expenses, income + expenses
}

var confirmBackup = func() bool {
	line := liner.NewLiner()
	defer line.Close()
	answer, err := line.Prompt("Did you take a picture? (y/n): ")
	return err == nil && strings.ToLower(strings.TrimSpace(answer)) == "y"
}
