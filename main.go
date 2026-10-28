package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
	Italic = "\033[3m"
)

const usage = Bold + "vk-finance" + Reset + ` - personal finance tracker

` + Bold + "Usage:" + Reset + `
  vk-finance [command]

` + Bold + "Commands:" + Reset + `
  ` + Cyan + "log" + Reset + `           add a new entry
  ` + Cyan + "history, h" + Reset + `    show past entries
  ` + Cyan + "undo" + Reset + `          remove the last entry
  ` + Cyan + "backup" + Reset + `        copy the data to the "` + driveLabel + `" drive
  ` + Cyan + "help" + Reset + `          show this message

Run without a command to show the current report.
`

const (
	driveLabel      = "VK DATA"
	driveMountPoint = "/media/veikko/VK DATA"
	BaseDB          = "FINANCE"
	File            = "finance.json"
	BaseLocal       = "DATABASES"
	// BaseLocal = "DEVELOPMENT"
)

var LocalFile = filepath.Join(BaseLocal, BaseDB, File)
var BackupFile = filepath.Join(driveMountPoint, BaseLocal, BaseDB, File)
var BackupFileWithDate = filepath.Join(driveMountPoint, BaseLocal, BaseDB, time.Now().AddDate(0, -1, 0).Format("January2006")+".json")

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		}
	}

	f := &Finance{}
	if err := f.Load(LocalFile); err != nil {
		log.Fatalf("load from file: %v", err)
	}

	var err error
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "log":
			err = f.Log()
		case "history", "h":
			err = f.HistoryAll()
			return
		case "undo":
			err = f.Undo()
		case "backup":
			err = f.Backup()
		default:
			fmt.Printf("%serror: unknown command: %s%s\n\n", Red, os.Args[1], Reset)
			fmt.Print(usage)
			os.Exit(1)
		}
	}

	f.Report()

	if err != nil {
		fmt.Printf("%serror: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}
}
