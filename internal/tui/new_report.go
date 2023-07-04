package tui

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/kraserh/energozvit/internal/storage"
)

type contentNewReport struct {
	tui      *Tui
	data     []*storage.Report
	table    *tview.Table
	modified bool
}

func newContentNewReport(t *Tui) *contentNewReport {
	content := new(contentNewReport)
	content.tui = t
	content.data = t.stor.GetNextReports()
	content.table = tview.NewTable().
		SetSelectable(true, false)
	content.setKeybinding()
	content.table.SetSelectedFunc(func(row, column int) {
		content.selectedRow(row)
	})
	return content
}

func (c *contentNewReport) GetName() string {
	return "newReport"
}

func (c *contentNewReport) GetMenuName() string {
	return "Нові дані"
}

func (c *contentNewReport) GetTitle() string {
	date := c.tui.stor.GetNextDate()
	title := fmt.Sprintf("%d-%02d", date.Year(), date.Month())
	if c.modified {
		title = title + " [:red](НЕ ЗБЕРЕЖЕНО)"
	}
	return title
}

func (c *contentNewReport) GetTable() *tview.Table {
	return c.table
}

func (c *contentNewReport) GetCell(row, column int) *tview.TableCell {
	// header
	var colName = []string{"Назва", "Номер", "Зона", "Теперешні",
		"Попередні", "Всього", "Примітка"}
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
			v = strconv.Itoa(c.data[row].Energy)
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		case 6:
			v = c.data[row].Annotation
			cell = tview.NewTableCell(v)
		}
	} else {

		// total row
		switch {
		case column == 5 && row == len(c.data):
			cell = tview.NewTableCell("------").
				SetAlign(tview.AlignRight)
		case column == 5 && row == len(c.data)+1:
			v = strconv.Itoa(c.tui.stor.GetNextTotal(c.data))
			cell = tview.NewTableCell(v).
				SetAlign(tview.AlignRight)
		default:
			cell = tview.NewTableCell("")
		}
	}
	return cell
}

func (c *contentNewReport) GetRowCount() int {
	return len(c.data) + 3 // +1 header row and +2 total row
}

func (c *contentNewReport) GetColumnCount() int {
	return 7
}

func (c *contentNewReport) GetKeybindingString() string {
	return "s: Зберегти,  u: Відміна"
}

func (c *contentNewReport) NeedToSave() bool {
	return c.modified
}

func (c *contentNewReport) RereadTable() {
	c.data = c.tui.stor.GetNextReports()
	c.modified = false
	c.tui.updateTable(c)
}

func (c *contentNewReport) selectedRow(row int) {
	row-- // -1 header
	if row >= len(c.data) || row < 0 {
		return
	}

	backupKwh := c.data[row].CurKwh
	backupAnnotation := c.data[row].Annotation
	dialog := newDialogNewReport(c.data[row])

	dialog.SetOkFunc(func() {
		c.modified = true
		c.tui.closeDialog(dialog)
		if row+1 < len(c.data) {
			c.table.Select(row+2, 0) // +1 header, +1 next
		}
	})

	dialog.SetCancelFunc(func() {
		c.data[row].CurKwh = backupKwh
		c.data[row].Annotation = backupAnnotation
		c.tui.closeDialog(dialog)
		c.table.Select(row+1, 0) // +1 header
	})

	c.tui.addAndSwitchToDialog(dialog)
}

func (c *contentNewReport) setKeybinding() {
	c.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 's':
			c.save()
		case 'u':
			c.undo()
		}
		return event
	})
}

func (c *contentNewReport) save() {
	err := c.tui.stor.SaveReports(c.data)
	if err != nil {
		c.tui.ErrorShow(err)
	} else {
		c.modified = false
		c.RereadTable()
		c.updateMetersOnReports()
	}
}

func (c *contentNewReport) undo() {
	c.modified = false
	c.RereadTable()
}

func (c *contentNewReport) updateMetersOnReports() {
	content, ok := c.tui.searchContent("reports")
	if ok {
		content.RereadTable()
	}
}

////////////////////////////////////////////////////////////////////////

type dialogNewReport struct {
	form       *tview.Form
	report     *storage.Report
	okFunc     func()
	cancelFunc func()
}

func newDialogNewReport(report *storage.Report) *dialogNewReport {
	dialog := &dialogNewReport{
		form:   tview.NewForm(),
		report: report,
	}
	dialog.addCurKwhField()
	dialog.addAnnotationField()
	dialog.addButtonOk()
	dialog.addButtonCancel()
	return dialog
}

func (d *dialogNewReport) GetTitle() string {
	return ""
}

func (d *dialogNewReport) GetPrimitive() tview.Primitive {
	return d.form
}

func (d *dialogNewReport) GetBox() *tview.Box {
	return d.form.Box
}

func (d *dialogNewReport) SetOkFunc(f func()) {
	d.okFunc = f
	d.form.GetButton(0).SetSelectedFunc(f)
}

func (d *dialogNewReport) SetCancelFunc(f func()) {
	d.cancelFunc = f
	d.form.GetButton(1).SetSelectedFunc(f)
}

// Поле вводу показників лічильника
func (d *dialogNewReport) addCurKwhField() {
	curKwhField := tview.NewInputField()
	curKwhField.
		SetLabel("Показник лічильника").
		SetFieldWidth(inputWidth).
		SetPlaceholder(strconv.Itoa(d.report.CurKwh)).
		// Дозволено ввод лише чисел
		SetAcceptanceFunc(func(_ string, lastChar rune) bool {
			return lastChar >= '0' && lastChar <= '9'
		}).
		// Натискання Enter в полі вводу
		SetDoneFunc(func(key tcell.Key) {
			newCurKwh, err := strconv.Atoi(
				curKwhField.GetText())
			if err == nil {
				d.report.CurKwh = newCurKwh
			}
			d.report.Calculate()
			//d.changeInfo(t)
		})
	d.form.AddFormItem(curKwhField)
}

// Поле вводу примітки
func (d *dialogNewReport) addAnnotationField() {
	annotationField := tview.NewInputField()
	annotationField.
		SetLabel("Примітка").
		SetFieldWidth(inputWidth).
		SetPlaceholder(d.report.Annotation).
		// Обмеження довжини примітки
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= 32
		}).
		// Натискання Enter в полі примітки
		SetDoneFunc(func(key tcell.Key) {
			newAnnotation := annotationField.GetText()
			if newAnnotation != "" {
				d.report.Annotation = newAnnotation
			}
		})
	d.form.AddFormItem(annotationField)
}

// Кнопка ОК
func (d *dialogNewReport) addButtonOk() {
	d.form.AddButton("OK", d.okFunc)
}

// Кнопка Відміна
func (d *dialogNewReport) addButtonCancel() {
	d.form.AddButton("Відміна", d.cancelFunc)
}
