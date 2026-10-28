package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VkHyperNova/vkmount"
)

/* Storage */

var mountBackup = func() error {
	return vkmount.Drive{Label: driveLabel}.Mount()
}

func (f *Finance) Save() error {
	copySlice := make([]Item, len(f.Finance))
	copy(copySlice, f.Finance)

	data, err := json.MarshalIndent(Finance{Finance: copySlice}, "", "  ")
	if err != nil {
		return err
	}

	// Save local
	if err := os.MkdirAll(filepath.Dir(LocalFile), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(LocalFile, data, 0o644); err != nil {
		return err
	}
	fmt.Println(Green + "Local save!" + Reset)

	// Save backup (never fails the whole save)
	if err := mountBackup(); err != nil {
		switch {
		case errors.Is(err, vkmount.ErrSkipped):
			fmt.Println(Yellow + "Backup skipped: " + driveLabel + " is not connected" + Reset)
		case errors.Is(err, vkmount.ErrSkipped):
			fmt.Println(Yellow + "Backup skipped: drive not mounted" + Reset)
		default:
			fmt.Println(Yellow + "Backup failed: " + err.Error() + Reset)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(BackupFile), 0o755); err != nil {
		fmt.Println(Yellow + "Backup failed: " + err.Error() + Reset)
		return nil
	}
	if err := os.WriteFile(BackupFile, data, 0o644); err != nil {
		fmt.Println(Yellow + "Backup write failed: " + err.Error() + Reset)
		return nil
	}
	fmt.Println(Green + "Backup save!" + Reset)

	return nil
}

func (f *Finance) Load(source string) error {
	data, err := os.ReadFile(source)

	// no file yet, start with empty data
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, f)
}
