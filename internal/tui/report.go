package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/kraserh/energozvit/internal/storage"
)

type contentReport struct {
	tui   *Tui
	date  time.Time
	data  []*storage.Report
	table *tview.Table
}

func newContentReport(t *Tui) *contentReport {
	content := new(contentReport)
	content.tui = t
	content.date = t.stor.GetNextDate().AddDate(0, -1, 0)
	content.data = t.stor.GetReports(content.date)
	content.table = tview.NewTable().
		SetSelectable(false, false)
	content.setKeybinding()
	return content
}

func (c *contentReport) GetName() string {
	return "reports"
}

func (c *contentReport) GetMenuName() string {
	return "Звіт"
}

func (c *contentReport) GetTitle() string {
	return fmt.Sprintf("%d-%02d", c.date.Year(), c.date.Month())
}

func (c *contentReport) GetTable() *tview.Table {
	return c.table
}

func (c *contentReport) GetCell(row, column int) *tview.TableCell {
	// header
	var colName = []string{"Назва", "Номер", "Зона", "Теперешні",
		"Попередні", "Різниця", "Всього", "Примітка"}
	row -= 1 // -1 header
	var v string
	cell := new(tview.TableCell)

	if row < 0 {
		// header row
		v = colName[column]
		cell = tview.NewTableCell(v)

	} else if row < len(c.data) {
		// data rows
		switch column {
		case 0:
			v = c.data[row].Name
			cell = tview.NewTableCell(v)
		case 1:
			v = c.data[row].Serial
			cell = tview.NewTableCell(v)
		case 2:
			v = strconv.Itoa(c.data[row].Zone)
			cell = tview.NewTableCell(v)
		case 3:
			v = strconv.Itoa(c.data[row].CurKwh)
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		case 4:
			v = strconv.Itoa(c.data[row].PreKwh)
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		case 5:
			v = strconv.Itoa(c.data[row].Diff)
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		case 6:
			v = strconv.Itoa(c.data[row].Energy)
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		case 7:
			v = c.data[row].Annotation
			cell = tview.NewTableCell(v)
		}
	} else {

		// total row
		switch {
		case column == 6 && row == len(c.data):
			cell = tview.NewTableCell("------").
				SetAlign(tview.AlignRight)
		case column == 6 && row == len(c.data)+1:
			v = strconv.Itoa(c.tui.stor.GetTotal(c.date, c.date))
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		default:
			cell = tview.NewTableCell("")
		}
	}
	return cell
}

func (c *contentReport) GetRowCount() int {
	return len(c.data) + 3 // +1 header row and +2 total row
}

func (c *contentReport) GetColumnCount() int {
	return 8
}

func (c *contentReport) GetKeybindingString() string {
	return "m/M: Місяць,  y/Y: Рік,  z: Останній звіт  a: Додатково"
}

func (c *contentReport) NeedToSave() bool {
	return false
}

func (c *contentReport) RereadTable() {
	c.lastDate()
}

func (c *contentReport) setKeybinding() {
	c.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'm':
			c.prevMonth()
		case 'M':
			c.nextMonth()
		case 'y':
			c.prevYear()
		case 'Y':
			c.nextYear()
		case 'z':
			c.lastDate()
		case 'a':
			c.additional()
		}
		return event
	})
}

func (c *contentReport) nextMonth() {
	c.date = c.date.AddDate(0, 1, 0)
	c.data = c.tui.stor.GetReports(c.date)
	c.tui.updateTable(c)
}

func (c *contentReport) prevMonth() {
	c.date = c.date.AddDate(0, -1, 0)
	c.data = c.tui.stor.GetReports(c.date)
	c.tui.updateTable(c)
}

func (c *contentReport) nextYear() {
	c.date = c.date.AddDate(1, 0, 0)
	c.data = c.tui.stor.GetReports(c.date)
	c.tui.updateTable(c)
}

func (c *contentReport) prevYear() {
	c.date = c.date.AddDate(-1, 0, 0)
	c.data = c.tui.stor.GetReports(c.date)
	c.tui.updateTable(c)
}

func (c *contentReport) lastDate() {
	c.date = c.tui.stor.GetNextDate().AddDate(0, -1, 0)
	c.data = c.tui.stor.GetReports(c.date)
	c.tui.updateTable(c)
}

func (c *contentReport) additional() {
	dialog := newDialogAdditional(c.tui, c.date)
	c.tui.addAndSwitchToDialog(dialog)
}

////////////////////////////////////////////////////////////////////////

type dialogAdditional struct {
	list *tview.List
}

func newDialogAdditional(t *Tui, date time.Time) *dialogAdditional {
	dialog := &dialogAdditional{
		list: tview.NewList(),
	}

	files, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}

	var cmdFiles []string
	for _, file := range files {

		fileInfo, err := os.Stat(file.Name())
		if err != nil {
			continue
		}
		mode := fileInfo.Mode()
		if mode.IsRegular() && ((mode.Perm() & 0111) != 0) {
			cmdFiles = append(cmdFiles, file.Name())
		}
	}

	if len(cmdFiles) > 0 {
		list := dialog.list
		for i, name := range cmdFiles {
			shotcut := rune(i + int('1'))
			list.AddItem(name, "", shotcut, nil)
		}

		list.SetSelectedFunc(func(_ int, cmdName, _ string, _ rune) {
			yymm := fmt.Sprintf("%d-%02d",
				date.Year(), int(date.Month()))
			t.execCommand("./"+cmdName, yymm)
			t.closeDialog(dialog)

		})

		list.SetDoneFunc(func() {
			t.closeDialog(dialog)
		})
	}

	return dialog
}

func (d *dialogAdditional) GetTitle() string {
	return "Додаткові команди"
}

func (d *dialogAdditional) GetPrimitive() tview.Primitive {
	return d.list
}

func (d *dialogAdditional) GetBox() *tview.Box {
	return d.list.Box
}

func (d *dialogAdditional) SetOkFunc(f func()) {
}

func (d *dialogAdditional) SetCancelFunc(f func()) {
}

// Виконує команду операційної системи
// Перший аргумент завжди зберігай імʼя бази даних
func (t *Tui) execCommand(cmdName string, args ...string) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	argsMod := []string{t.stor.GetFilepath()}
	argsMod = append(argsMod, args...)
	t.app.Suspend(func() {
		cmd := exec.Command(cmdName, argsMod...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			t.ErrorShow(err)
		}
	})

	err = os.Chdir(pwd)
	if err != nil {
		panic(err)
	}
}
