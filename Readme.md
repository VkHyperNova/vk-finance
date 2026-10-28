# VK-Finance

A small command-line personal finance tracker written in Go. Log income and expenses from the terminal, see a colour-coded monthly report, and keep everything in a single local JSON file with automatic backup to a USB drive.

## Features

- **Quick logging**: one command adds an entry (`vk-finance log coffee -3.50`)
- **Monthly report**: income, expenses and balance, plus totals grouped by item
- **History**: view today's entries in the report, or the full list with `history`
- **Undo**: remove the last entry if you made a mistake
- **Automatic backup**: every save is also copied to the `VK DATA` drive when it is connected
- **Monthly archive**: `backup` archives the month to a dated file, resets the database and carries over any debt
- **Plain JSON storage**: no database required

## Requirements

- [Go](https://go.dev/dl/) 1.25 or newer
- Linux or Windows (the screen is cleared with `clear` / `cls`)
- Optional: a drive labelled `VK DATA` for backups

## Installation

```bash
git clone https://github.com/VkHyperNova/vk-finance.git
cd vk-finance
go build
```

This produces a `vk-finance` binary. You can also run it without building:

```bash
go run .
```

## Usage

```
vk-finance [command]
```

| Command | Description |
|---|---|
| *(none)* | Show the current report |
| `log <item> <sum>` | Add a new entry |
| `history`, `h` | Show all past entries |
| `undo` | Remove the last entry |
| `backup` | Archive the current month to the `VK DATA` drive and reset |
| `help`, `-h`, `--help` | Show usage |

### Examples

```bash
vk-finance log salary 2500      # income (positive value)
vk-finance log groceries -42.80 # expense (negative value)
vk-finance log "bus ticket" -2  # quote items that contain spaces
vk-finance undo                 # remove the last entry
vk-finance                      # show the report
```

Positive values count as income, negative values as expenses. Item names are lowercased, and entries with the same name are summed together in the report. All amounts are shown in EUR.

### The report

The default view shows the current month, then:

- `INCOME`, `EXPENSES` and `BALANCE` totals
- Per-item totals sorted from highest to lowest (items that sum to zero are hidden, and the most recently added item is highlighted)
- Today's history

## Data storage

Entries are stored in `DATABASES/FINANCE/finance.json`, relative to the directory you run the program from. The file is created on first save.

```json
{
  "finance": [
    {
      "date": "28-10-2026",
      "time": "20:05:00",
      "comment": "groceries",
      "value": -42.8
    }
  ]
}
```

Dates use the `DD-MM-YYYY` format.

## Backups

- **On every save**, the data is also written to `/media/veikko/VK DATA/DATABASES/FINANCE/finance.json` if the `VK DATA` drive can be mounted. If it is not connected, the backup is skipped with a warning and the local save still succeeds.
- **`vk-finance backup`** asks for confirmation (`Did you take a picture? (y/n)`), then writes a monthly archive named after the previous month (for example `September2026.json`) to the same folder on the drive. An existing archive is never overwritten. After a successful archive the local database is emptied. If the balance was negative, it is carried over as a single `debt` entry.

The drive label and mount point are constants at the top of `main.go`; change them to match your setup.

## Project structure

```
├── main.go           # Entry point, command dispatch, paths and constants
├── finance.go        # Data model and commands (log, undo, backup, totals)
├── report.go         # Report and history output
├── storage.go        # Loading and saving JSON, drive backup
├── *_test.go         # Tests
├── go.mod
└── go.sum
```

## Testing

```bash
go test ./...
```

## Dependencies

- [`peterh/liner`](https://github.com/peterh/liner) for the interactive confirmation prompt
- [`VkHyperNova/vkmount`](https://github.com/VkHyperNova/vkmount) for mounting the backup drive

## License

No license specified. All rights reserved by the author unless stated otherwise.