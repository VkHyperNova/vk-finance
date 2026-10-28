package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"time"
)

func (f *Finance) Report() {

	clearScreen()

	fmt.Print(Blue + Bold + "\nVK FINANCE v2\n\n" + Reset)

	// Print current month
	currentMonth := time.Now().Format("January 2006")
	fmt.Println(Yellow + "\t" + currentMonth + "\n" + Reset)

	income, expenses, balance := f.calculate()
	fmt.Println(Green, "INCOME: "+"+"+strconv.FormatFloat(income, 'f', 2, 64)+" EUR", Reset)
	fmt.Println(Red, "EXPENSES: "+strconv.FormatFloat(expenses, 'f', 2, 64)+" EUR", Reset)
	fmt.Println(Bold, "BALANCE: "+strconv.FormatFloat(balance, 'f', 2, 64)+" EUR", Reset)
	fmt.Println()

	f.printItemsBySum()

	f.HistoryToday()

	fmt.Println("\n< log, history, undo, backup, help >")
}

func (f *Finance) HistoryToday() error {
	fmt.Println(Bold+Yellow, "\nHistory (today): \n", Reset)

	todayDate := time.Now().Format("02-01-2006")
	found := false

	for _, value := range f.Finance {
		if value.DATE != todayDate {
			continue
		}
		found = true
		fmt.Println(formatEntry(value, false))
	}

	if !found {
		fmt.Println(Gray + "\tNothing added today." + Reset)
	}
	return nil
}

func (f *Finance) HistoryAll() error {
	fmt.Println(Bold+Yellow, "\nHistory (all): \n", Reset)

	todayDate := time.Now().Format("02-01-2006")

	if len(f.Finance) == 0 {
		fmt.Println(Gray + "\tNothing to show." + Reset)
		return nil
	}

	for _, value := range f.Finance {
		line := formatEntry(value, true)
		if value.DATE == todayDate {
			line = Bold + line + Reset
		}
		fmt.Println(line)
	}
	return nil
}

func (f *Finance) printItemsBySum() {

	// Guard against empty slice panic
	if len(f.Finance) == 0 {
		return
	}

	// Group and Sum up Items
	totals := make(map[string]float64)
	for _, item := range f.Finance {
		totals[item.COMMENT] += item.VALUE
	}

	// Convert to sortable slice, skipping zero-sum entries
	type itemSum struct {
		name string
		sum  float64
	}
	pairs := make([]itemSum, 0, len(totals))
	for name, sum := range totals {
		if sum != 0 {
			pairs = append(pairs, itemSum{name, sum})
		}
	}

	// Sort highest to lowest
	sort.Slice(pairs, func(a, b int) bool {
		return pairs[a].sum > pairs[b].sum
	})

	// Last added item is used for highlight
	lastItem := f.Finance[len(f.Finance)-1]

	for _, p := range pairs {
		line := fmt.Sprintf("\t%s: %.2f EUR", p.name, p.sum)

		if p.name == lastItem.COMMENT {
			line += fmt.Sprintf(" | %.2f", lastItem.VALUE)
		}

		highlight := p.name == lastItem.COMMENT
		fmt.Println(colorize(line, p.sum, highlight))
	}
}

/* Hepler functions */

func clearScreen() {
	switch runtime.GOOS {
	case "linux": // check if the operating system is Linux
		cmd := exec.Command("clear") // execute the clear command
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls") // execute the cls command
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func formatEntry(value Item, withDate bool) string {
	color := Bold
	if value.VALUE > 0 {
		color = Bold + Green
	} else if value.VALUE < 0 {
		color = Bold + Red
	}

	sum := color + value.COMMENT + " " + strconv.FormatFloat(value.VALUE, 'f', -1, 64) + " EUR" + Reset

	if withDate {
		return " " + value.TIME + " | " + value.DATE + " " + sum
	}
	return " " + value.TIME + " | " + sum
}

func colorize(line string, value float64, highlight bool) string {
	if highlight {
		return Bold + Yellow + line + Reset
	}
	switch {
	case value > 0:
		return Green + line + Reset
	case value < 0:
		return Red + line + Reset
	default:
		return line
	}
}
