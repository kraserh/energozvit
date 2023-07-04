package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/kraserh/energozvit/internal/storage"
)

type contentMeters struct {
	tui   *Tui
	data  []*storage.Meter
	table *tview.Table
}

func newContentMeters(t *Tui) *contentMeters {
	content := new(contentMeters)
	content.tui = t
	content.data = t.stor.GetActiveMeters()
	content.table = tview.NewTable().
		SetSelectable(true, false)
	content.setKeybinding()
	return content
}

func (c *contentMeters) GetName() string {
	return "meters"
}

func (c *contentMeters) GetMenuName() string {
	return "Лічильники"
}

func (c *contentMeters) GetTitle() string {
	return ""
}

func (c *contentMeters) GetTable() *tview.Table {
	return c.table
}

func (c *contentMeters) GetCell(row, column int) *tview.TableCell {
	var colName = []string{"КТП", "EIC", "Назва", "Модель",
		"Рік", "Номер", "Розряди", "Множник"}
	row -= 1 // -1 header
	var v string

	if row < 0 {
		// header row
		v = colName[column]

	} else {
		// data rows
		switch column {
		case 0:
			v = strconv.Itoa(c.data[row].Substation)
		case 1:
			v = c.data[row].Eic
		case 2:
			v = c.data[row].Name
		case 3:
			v = c.data[row].Model
		case 4:
			v = strconv.Itoa(c.data[row].Year)
		case 5:
			v = c.data[row].Serial
		case 6:
			v = strconv.Itoa(c.data[row].Digits)
		case 7:
			v = strconv.Itoa(c.data[row].Ratio)
		}
	}
	return tview.NewTableCell(v)
}

func (c *contentMeters) GetRowCount() int {
	return len(c.data) + 1 // +1 header
}

func (c *contentMeters) GetColumnCount() int {
	return 8
}

func (c *contentMeters) GetKeybindingString() string {
	return "n: Додати  d: Видалити  e: Редагувати"
}

func (c *contentMeters) NeedToSave() bool {
	return false
}

func (c *contentMeters) RereadTable() {
	c.data = c.tui.stor.GetActiveMeters()
	c.tui.updateTable(c)
}

func (c *contentMeters) getSelection() (*storage.Meter, bool) {
	row, _ := c.table.GetSelection()
	row -= 1
	if row < len(c.data) {
		return c.data[row], true
	}
	return nil, false
}

func (c *contentMeters) setKeybinding() {
	c.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'n':
			c.create()
		case 'd':
			c.delete()
		case 'e':
			c.edit()
		}
		return event
	})
}

func (c *contentMeters) create() {
	meter, ok := c.getSelection()
	if !ok {
		meter = new(storage.Meter)
	}
	dialog := newDialogCreateMeter(meter)

	dialog.SetOkFunc(func() {
		err := c.tui.stor.AddMeter(dialog.meter, dialog.firstKwh)
		if err != nil {
			c.tui.ErrorShow(err)
		} else {
			c.data = c.tui.stor.GetActiveMeters()
			c.tui.closeDialog(dialog)
			c.updateMetersOnNewReports()
		}
	})

	dialog.SetCancelFunc(func() {
		c.tui.closeDialog(dialog)
	})

	c.tui.addAndSwitchToDialog(dialog)
}

func (c *contentMeters) delete() {
	meter, ok := c.getSelection()
	if !ok {
		return
	}
	message := fmt.Sprintf("Буде видалено лічильник %s № %s",
		meter.Name, meter.Serial)
	c.tui.Confirm(message,
		func() {
			err := c.tui.stor.RemoveMeter(meter)
			if err != nil {
				c.tui.ErrorShow(err)
			}
			c.data = c.tui.stor.GetActiveMeters()
			c.updateMetersOnNewReports()
		})
}

func (c *contentMeters) edit() {
	meter, ok := c.getSelection()
	if !ok {
		return
	}
	dialog := newDialogEditMeter(meter)

	dialog.SetOkFunc(func() {
		err := c.tui.stor.UpdateMeter(dialog.meter)
		if err != nil {
			c.tui.ErrorShow(err)
		} else {
			c.data = c.tui.stor.GetActiveMeters()
			c.tui.closeDialog(dialog)
			c.updateMetersOnNewReports()
		}
	})

	dialog.SetCancelFunc(func() {
		c.tui.closeDialog(dialog)
	})

	c.tui.addAndSwitchToDialog(dialog)
}

func (c *contentMeters) updateMetersOnNewReports() {
	content, ok := c.tui.searchContent("newReports")
	if ok {
		content.RereadTable()
	}
}

////////////////////////////////////////////////////////////////////////

type dialogMeter struct {
	form       *tview.Form
	meter      *storage.Meter
	firstKwh   []int
	isCreate   bool
	okFunc     func()
	cancelFunc func()
}

func newDialogCreateMeter(meter *storage.Meter) *dialogMeter {
	dialog := &dialogMeter{
		form:     tview.NewForm(),
		meter:    meter,
		isCreate: true,
	}
	dialog.addSubstationField()
	dialog.addEicField()
	dialog.addNameField()
	dialog.addModelField()
	dialog.addYearField()
	dialog.addSerialField()
	dialog.addDigitsField()
	dialog.addRatioField()
	dialog.addFirstKwhField()
	dialog.addButtonOk()
	dialog.addButtonCancel()
	return dialog
}

func newDialogEditMeter(meter *storage.Meter) *dialogMeter {
	dialog := &dialogMeter{
		form:  tview.NewForm(),
		meter: meter,
	}
	dialog.addSubstationField()
	dialog.addEicField()
	dialog.addNameField()
	dialog.addModelField()
	dialog.addYearField()
	dialog.addSerialField()
	dialog.addDigitsField()
	dialog.addRatioField()
	dialog.addButtonOk()
	dialog.addButtonCancel()
	return dialog
}

func (d *dialogMeter) GetTitle() string {
	var title string
	if d.isCreate {
		title = "Додавання лічильника"
	} else {
		title = "Редагування " + d.meter.Name
	}
	return title
}

func (d *dialogMeter) GetPrimitive() tview.Primitive {
	return d.form
}

func (d *dialogMeter) GetBox() *tview.Box {
	return d.form.Box
}

func (d *dialogMeter) SetOkFunc(f func()) {
	d.okFunc = f
	d.form.GetButton(0).SetSelectedFunc(f)
}

func (d *dialogMeter) SetCancelFunc(f func()) {
	d.cancelFunc = f
	d.form.GetButton(1).SetSelectedFunc(f)
}

////////////////////////////////////////////////////////////////////////

// Поле вводу номеру підстанції
func (d *dialogMeter) addSubstationField() {
	substationField := tview.NewInputField()
	substationField.
		SetLabel("Номер підстанції").
		SetFieldWidth(inputWidth).
		SetPlaceholder(strconv.Itoa(d.meter.Substation)).
		SetAcceptanceFunc(isNumber).
		SetDoneFunc(func(key tcell.Key) {
			newSubstation, err :=
				strconv.Atoi(substationField.GetText())
			if err == nil {
				d.meter.Substation = newSubstation
			}
		})
	d.form.AddFormItem(substationField)
}

// Поле вводу EIC коду
func (d *dialogMeter) addEicField() {
	eicCodeField := tview.NewInputField()
	eicCodeField.
		SetLabel("EIC Код").
		SetFieldWidth(inputWidth).
		SetPlaceholder(d.meter.Eic).
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= 16
		}).
		SetDoneFunc(func(key tcell.Key) {
			newEic := eicCodeField.GetText()
			if newEic != "" {
				d.meter.Eic = newEic
			}
		})
	d.form.AddFormItem(eicCodeField)
}

// Поле вводу назви точки обліку
func (d *dialogMeter) addNameField() {
	nameField := tview.NewInputField()
	nameField.
		SetLabel("Назва точки обліку").
		SetFieldWidth(inputWidth).
		SetPlaceholder(d.meter.Name).
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= 24
		}).
		SetDoneFunc(func(key tcell.Key) {
			newName := nameField.GetText()
			if newName != "" {
				d.meter.Name = newName
			}
		})
	d.form.AddFormItem(nameField)
}

// Поле вводу моделі лічильника
func (d *dialogMeter) addModelField() {
	modelField := tview.NewInputField()
	modelField.
		SetLabel("Модель лічильника").
		SetFieldWidth(inputWidth).
		SetPlaceholder(d.meter.Model).
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= 24
		}).
		SetDoneFunc(func(key tcell.Key) {
			newModel := modelField.GetText()
			if newModel != "" {
				d.meter.Model = newModel
			}
		})
	d.form.AddFormItem(modelField)
}

// Поле вводу рік виготовлення лічильника
func (d *dialogMeter) addYearField() {
	yearField := tview.NewInputField()
	yearField.
		SetLabel("Рік лічильника").
		SetFieldWidth(inputWidth).
		SetPlaceholder(strconv.Itoa(d.meter.Year)).
		SetAcceptanceFunc(isNumber).
		SetDoneFunc(func(key tcell.Key) {
			newYear, err :=
				strconv.Atoi(yearField.GetText())
			if err == nil {
				d.meter.Year = newYear
			}
		})
	d.form.AddFormItem(yearField)
}

// Поле вводу серійного номеру лічильника
func (d *dialogMeter) addSerialField() {
	serialNumField := tview.NewInputField()
	serialNumField.
		SetLabel("Серійний номер").
		SetFieldWidth(inputWidth).
		SetPlaceholder(d.meter.Serial).
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= 24
		}).
		SetDoneFunc(func(key tcell.Key) {
			newSerial := serialNumField.GetText()
			if newSerial != "" {
				d.meter.Serial = newSerial
			}
		})
	d.form.AddFormItem(serialNumField)
}

// Поле вводу кількості значучих розрядів
func (d *dialogMeter) addDigitsField() {
	digitsMaxField := tview.NewInputField()
	digitsMaxField.
		SetLabel("Значучі розряди").
		SetFieldWidth(inputWidth).
		SetPlaceholder(strconv.Itoa(d.meter.Digits)).
		SetAcceptanceFunc(isNumber).
		SetDoneFunc(func(key tcell.Key) {
			newDigitsMax, err :=
				strconv.Atoi(digitsMaxField.GetText())
			if err == nil {
				d.meter.Digits = newDigitsMax
			}
		})
	d.form.AddFormItem(digitsMaxField)
}

// Поле вводу коефіцієнту трансформації
func (d *dialogMeter) addRatioField() {
	ratioField := tview.NewInputField()
	ratioField.
		SetLabel("Коефіцієнт тр-ції").
		SetFieldWidth(inputWidth).
		SetPlaceholder(strconv.Itoa(d.meter.Ratio)).
		SetAcceptanceFunc(isNumber).
		SetDoneFunc(func(key tcell.Key) {
			newRatio, err :=
				strconv.Atoi(ratioField.GetText())
			if err == nil {
				d.meter.Ratio = newRatio
			}
		})
	d.form.AddFormItem(ratioField)
}

// Поле вводу початкових показників
func (d *dialogMeter) addFirstKwhField() {
	firstKwhField := tview.NewInputField()
	firstKwhField.
		SetLabel("Показники, через пробіл").
		SetFieldWidth(inputWidth).
		SetAcceptanceFunc(func(text string, _ rune) bool {
			return len(text) <= inputWidth
		}).
		SetDoneFunc(func(key tcell.Key) {
			newFirstKwh :=
				strings.Split(firstKwhField.GetText(), " ")
			if len(newFirstKwh) != 0 {
				d.firstKwh = nil
			}

			for _, firstKwh := range newFirstKwh {
				kwh, err := strconv.Atoi(firstKwh)
				if err == nil {
					d.firstKwh =
						append(d.firstKwh, kwh)
				}
			}
		})

	d.form.AddFormItem(firstKwhField)
}

// Кнопка ОК
func (d *dialogMeter) addButtonOk() {
	d.form.AddButton("OK", d.okFunc)
}

// Кнопка Відміна
func (d *dialogMeter) addButtonCancel() {
	d.form.AddButton("Відміна", d.cancelFunc)
}
