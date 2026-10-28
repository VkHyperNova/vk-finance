package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureOutput runs fn and returns everything it printed to stdout.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	w.Close()
	os.Stdout = old
	return <-done
}

func TestFormatEntry(t *testing.T) {
	cases := []struct {
		name     string
		item     Item
		withDate bool
		want     string
	}{
		{
			"income without date",
			Item{DATE: "28-10-2026", TIME: "10:00:00", COMMENT: "salary", VALUE: 1500},
			false,
			" 10:00:00 | " + Bold + Green + "salary 1500 EUR" + Reset,
		},
		{
			"expense without date",
			Item{DATE: "28-10-2026", TIME: "12:30:00", COMMENT: "coffee", VALUE: -4.5},
			false,
			" 12:30:00 | " + Bold + Red + "coffee -4.5 EUR" + Reset,
		},
		{
			"zero value uses plain bold",
			Item{DATE: "28-10-2026", TIME: "09:00:00", COMMENT: "note", VALUE: 0},
			false,
			" 09:00:00 | " + Bold + "note 0 EUR" + Reset,
		},
		{
			"with date",
			Item{DATE: "28-10-2026", TIME: "10:00:00", COMMENT: "salary", VALUE: 1500},
			true,
			" 10:00:00 | 28-10-2026 " + Bold + Green + "salary 1500 EUR" + Reset,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatEntry(c.item, c.withDate); got != c.want {
				t.Fatalf("got  %q\nwant %q", got, c.want)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	cases := []struct {
		name      string
		value     float64
		highlight bool
		want      string
	}{
		{"positive", 5, false, Green + "x" + Reset},
		{"negative", -5, false, Red + "x" + Reset},
		{"zero", 0, false, "x"},
		{"highlight beats positive", 5, true, Bold + Yellow + "x" + Reset},
		{"highlight beats negative", -5, true, Bold + Yellow + "x" + Reset},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := colorize("x", c.value, c.highlight); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestPrintItemsBySumEmpty(t *testing.T) {
	out := captureOutput(t, func() { (&Finance{}).printItemsBySum() })
	if out != "" {
		t.Fatalf("empty data should print nothing, got %q", out)
	}
}

func TestPrintItemsBySumGroupsItems(t *testing.T) {
	f := &Finance{Finance: []Item{
		{COMMENT: "coffee", VALUE: -2},
		{COMMENT: "coffee", VALUE: -3},
	}}

	out := captureOutput(t, func() { f.printItemsBySum() })

	// Same item is summed, and the last added item also shows its own value
	if !strings.Contains(out, "\tcoffee: -5.00 EUR | -3.00") {
		t.Fatalf("expected grouped total with last value, got %q", out)
	}
}

func TestPrintItemsBySumSortedHighestFirst(t *testing.T) {
	f := &Finance{Finance: []Item{
		{COMMENT: "a", VALUE: 10},
		{COMMENT: "b", VALUE: 50},
		{COMMENT: "c", VALUE: -5},
	}}

	out := captureOutput(t, func() { f.printItemsBySum() })

	b := strings.Index(out, "\tb:")
	a := strings.Index(out, "\ta:")
	c := strings.Index(out, "\tc:")
	if b < 0 || a < 0 || c < 0 {
		t.Fatalf("missing items in output: %q", out)
	}
	if !(b < a && a < c) {
		t.Fatalf("expected order b, a, c but got positions b=%d a=%d c=%d", b, a, c)
	}
}

func TestPrintItemsBySumSkipsZeroTotals(t *testing.T) {
	f := &Finance{Finance: []Item{
		{COMMENT: "x", VALUE: 5},
		{COMMENT: "y", VALUE: 3},
		{COMMENT: "x", VALUE: -5}, // x nets to zero
	}}

	out := captureOutput(t, func() { f.printItemsBySum() })

	if strings.Contains(out, "\tx:") {
		t.Fatalf("zero-sum item should be hidden, got %q", out)
	}
	if !strings.Contains(out, "\ty: 3.00 EUR") {
		t.Fatalf("y should be shown, got %q", out)
	}
}

func TestPrintItemsBySumHighlightsLastItem(t *testing.T) {
	f := &Finance{Finance: []Item{
		{COMMENT: "a", VALUE: 10},
		{COMMENT: "b", VALUE: 20}, // last added
	}}

	out := captureOutput(t, func() { f.printItemsBySum() })

	if !strings.Contains(out, Bold+Yellow+"\tb: 20.00 EUR | 20.00"+Reset) {
		t.Errorf("last item should be highlighted, got %q", out)
	}
	if !strings.Contains(out, Green+"\ta: 10.00 EUR"+Reset) {
		t.Errorf("other items should use plain color, got %q", out)
	}
}

func TestHistoryToday(t *testing.T) {
	today := time.Now().Format("02-01-2006")
	f := &Finance{Finance: []Item{
		{DATE: "01-01-2000", TIME: "08:00:00", COMMENT: "olditem", VALUE: -1},
		{DATE: today, TIME: "09:00:00", COMMENT: "newitem", VALUE: -2},
	}}

	out := captureOutput(t, func() {
		if err := f.HistoryToday(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "newitem") {
		t.Errorf("today's entry should be shown, got %q", out)
	}
	if strings.Contains(out, "olditem") {
		t.Errorf("old entry should be hidden, got %q", out)
	}
}

func TestHistoryTodayNothingAdded(t *testing.T) {
	f := &Finance{Finance: []Item{
		{DATE: "01-01-2000", TIME: "08:00:00", COMMENT: "olditem", VALUE: -1},
	}}

	out := captureOutput(t, func() { f.HistoryToday() })

	if !strings.Contains(out, "Nothing added today.") {
		t.Fatalf("expected the empty message, got %q", out)
	}
}

func TestHistoryAll(t *testing.T) {
	today := time.Now().Format("02-01-2006")
	oldItem := Item{DATE: "01-01-2000", TIME: "08:00:00", COMMENT: "olditem", VALUE: -1}
	newItem := Item{DATE: today, TIME: "09:00:00", COMMENT: "newitem", VALUE: -2}
	f := &Finance{Finance: []Item{oldItem, newItem}}

	out := captureOutput(t, func() {
		if err := f.HistoryAll(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, formatEntry(oldItem, true)) {
		t.Errorf("old entry should be shown with its date, got %q", out)
	}
	if !strings.Contains(out, Bold+formatEntry(newItem, true)+Reset) {
		t.Errorf("today's entry should be wrapped in bold, got %q", out)
	}
	if strings.Contains(out, Bold+formatEntry(oldItem, true)+Reset) {
		t.Errorf("old entry must not be wrapped in bold, got %q", out)
	}
}

func TestHistoryAllEmpty(t *testing.T) {
	out := captureOutput(t, func() { (&Finance{}).HistoryAll() })

	if !strings.Contains(out, "Nothing to show.") {
		t.Fatalf("expected the empty message, got %q", out)
	}
}

func TestReport(t *testing.T) {
	f := sample() // income 1500, expenses -4.5, balance 1495.5

	out := captureOutput(t, func() { f.Report() })

	want := []string{
		"VK FINANCE v2",
		time.Now().Format("January 2006"),
		"INCOME: +1500.00 EUR",
		"EXPENSES: -4.50 EUR",
		"BALANCE: 1495.50 EUR",
		"History (today):",
		"< log, history, undo, backup >",
	}
	for _, s := range want {
		if !strings.Contains(out, s) {
			t.Errorf("report is missing %q\n%s", s, out)
		}
	}
}
