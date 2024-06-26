package cabrisync

import (
	"fmt"
	"io"
	"sort"
)

type SyncReportEntry struct {
	IsNs      bool // entry is a namespace
	isSymLink bool
	LPath     string // content's path in left DSS
	RPath     string // content's path in right DSS
	isRTL     bool   // if BiDir is active, indicates the synchronization is reversed: right to left
	Created   bool   // content is created on target
	Updated   bool   // content is updated on target
	Removed   bool   // content is removed on target
	Kept      bool   // content is kept on target
	MUpdated  bool   // meta data is updated on target
	Excluded  bool   // content was excluded
	Err       error  // if entry synchronization has errors
}

// SyncReport provides the Synchronize execution result
type SyncReport struct {
	GErr    error             // global error if synchronization aborted
	Entries []SyncReportEntry // report information for each entry
}

// SyncStats provides SyncReport statistics
type SyncStats struct {
	CreNum  int // number of created entries
	UpdNum  int // number of updated entries
	RmvNum  int // number of removed entries
	KeptNum int // number of kept entries
	MUpNum  int // number of meta data updated entries
	ErrNum  int // number of errors (excl. GErr)
}

// HasErrors indicates if any synchronization error occurred
func (sr SyncReport) HasErrors() bool {
	if sr.GErr != nil {
		return true
	}
	for _, entry := range sr.Entries {
		if entry.Err != nil {
			return true
		}
	}
	return false
}

// GetStats evaluates stats from report
func (sr SyncReport) GetStats() SyncStats {
	var syst SyncStats
	for _, entry := range sr.Entries {
		if entry.Created {
			syst.CreNum++
		}
		if entry.Updated {
			syst.UpdNum++
		}
		if entry.Removed {
			syst.RmvNum++
		}
		if entry.Kept {
			syst.KeptNum++
		}
		if entry.MUpdated {
			syst.MUpNum++
		}
		if entry.Err != nil {
			syst.ErrNum++
		}
	}
	return syst
}

// SortByPath builds a report sorted by entry names from existing report
func (sr SyncReport) SortByPath() (ssr SyncReport) {
	ssr.GErr = sr.GErr
	for _, entry := range sr.Entries {
		ssr.Entries = append(ssr.Entries, entry)
	}
	sort.Slice(ssr.Entries, func(i, j int) bool {
		return ssr.Entries[i].LPath < ssr.Entries[j].LPath
	})
	return
}

func (sr SyncReport) doTextOutput(out io.Writer, summary, dispRight bool) {
	for _, entry := range sr.Entries {
		arrow := '>'
		if entry.isRTL {
			arrow = '<'
		}
		c := '.'
		switch true {
		case entry.MUpdated:
			c = 'm'
		case entry.Created:
			c = '+'
		case entry.Updated:
			c = '*'
		case entry.Removed:
			c = 'x'
		case entry.Kept:
			c = '~'
		case entry.Excluded:
			c = ';'
		}
		rpathOmitIf := "-"
		if entry.RPath != entry.LPath || dispRight {
			rpathOmitIf = entry.RPath
		}
		if entry.Err == nil && (!summary || (c != '.' && c != ';')) {
			out.Write([]byte(fmt.Sprintf("%c%c %s %s\n", arrow, c, entry.LPath, rpathOmitIf)))
		} else if entry.Err != nil {
			c = '?'
			out.Write([]byte(fmt.Sprintf("%c%c %s %s %v\n", arrow, c, entry.LPath, rpathOmitIf, entry.Err)))
		}
	}
}

// TextOutput displays human readable report on given output
func (sr SyncReport) TextOutput(out io.Writer, dispRight bool) {
	sr.doTextOutput(out, false, dispRight)
}

// TextOutput4Test displays human readable report on given output, legacy force right display
func (sr SyncReport) TextOutput4Test(out io.Writer) {
	sr.doTextOutput(out, false, true)
}

// SummaryOutput displays human readable summary report on given output: only differences are displayed
func (sr SyncReport) SummaryOutput(out io.Writer, dispRight bool) {
	sr.doTextOutput(out, true, dispRight)
}

// SyncRefDiag provides a reference report indexed by left and right paths for diagnosis purpose
type SyncRefDiag struct {
	Left  map[string]SyncReportEntry
	Right map[string]SyncReportEntry
}

// GetRefDiag creates a SyncRefDiag from the current report
func (sr SyncReport) GetRefDiag() (refDiag SyncRefDiag) {
	refDiag.Left = map[string]SyncReportEntry{}
	refDiag.Right = map[string]SyncReportEntry{}
	for _, entry := range sr.Entries {
		refDiag.Left[entry.LPath] = entry
		refDiag.Right[entry.RPath] = entry
	}
	return
}
