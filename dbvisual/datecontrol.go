package dbvisual

import (
	"github.com/gotk3/gotk3/gtk"
)

// dateControl зберігає дані для блоку управління датою
type dateControl struct {
	frame     *gtk.Frame
	spinYear  *gtk.SpinButton
	spinMonth *gtk.SpinButton
	minYear   int
	minMonth  int
	maxYear   int
	maxMonth  int
	step      DateStep
	callback  func()
}

// newDateBox створює блок управління датою
func newDateControl() (*dateControl, *gtk.Frame) {
	dc := new(dateControl)

	// Рамка навколо панелі
	var err error
	dc.frame, err = gtk.FrameNew("Дата")
	check(err)
	dc.frame.SetLabelAlign(0.5, 0.5)

	// Розбиваєм по вертикалі для кнопок і спіну
	inFrame, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	dc.frame.Add(inFrame)

	// Spin
	boxSpin, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	check(err)

	// Year
	boxYear, spinYear := setupSpinBox("Рік:", 2000, 2100)
	boxSpin.PackStart(boxYear, true, true, 5)
	spinYear.Connect("value-changed", dc.changedYear)
	dc.spinYear = spinYear

	// Month
	boxMonth, spinMonth := setupSpinBox("Місяць:", 1, 12)
	boxSpin.PackStart(boxMonth, true, true, 5)
	spinMonth.Connect("value-changed", dc.updated)
	dc.spinMonth = spinMonth
	inFrame.PackEnd(boxSpin, false, false, 12)

	//Buttons box
	boxButtons, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	check(err)
	inFrame.PackEnd(boxButtons, false, false, 8)

	// Кнопка "Назад"
	backButton, err := gtk.ButtonNewFromIconName("pan-start-symbolic", 1)
	check(err)
	boxButtons.PackStart(backButton, true, true, 5)
	backButton.Connect("clicked", dc.Prev)

	// Кнопка "Вперед"
	forwardButton, err := gtk.ButtonNewFromIconName("pan-end-symbolic", 1)
	check(err)
	boxButtons.PackStart(forwardButton, true, true, 5)
	forwardButton.Connect("clicked", dc.Next)
	return dc, dc.frame
}

// setupSpinBox налаштовує кнопку-прокрутку
func setupSpinBox(label string, min, max int) (*gtk.Box, *gtk.SpinButton) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	gtklabel, err := gtk.LabelNew(label)
	check(err)
	gtklabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(gtklabel, false, true, 0)
	spin, err := gtk.SpinButtonNewWithRange(float64(min), float64(max), 1)
	check(err)
	box.PackStart(spin, false, true, 0)
	return box, spin
}

// changedYear змінено рік
func (dc *dateControl) changedYear() {
	year, _ := dc.Get()
	if year == dc.minYear {
		dc.spinMonth.SetRange(float64(dc.minMonth), 12)
	} else if year == dc.maxYear {
		dc.spinMonth.SetRange(1, float64(dc.maxMonth))
	} else {
		dc.spinMonth.SetRange(1, 12)
	}
	dc.updated()
}

// updated виконується при зміні дати
func (dc *dateControl) updated() {
	if dc.callback != nil {
		dc.callback()
	}
}

// setUpdateFunc встановлює функцію яка буде виконуватись при зміні дати
func (dc *dateControl) setUpdateFunc(f func()) {
	dc.callback = f
}

// SetMin встановлює мінімальну дату
func (dc *dateControl) SetMin(year, month int) {
	dc.minYear = year
	dc.minMonth = month
	dc.spinYear.SetRange(float64(dc.minYear), float64(dc.maxYear))
}

// SetMax встановлює максимальну дату
func (dc *dateControl) SetMax(year, month int) {
	dc.maxYear = year
	dc.maxMonth = month
	dc.spinYear.SetRange(float64(dc.minYear), float64(dc.maxYear))
}

// Prev дата на крок вперед
func (dc *dateControl) Prev() {
	year, month := dc.Get()
	switch dc.step {
	case DATE_STEP_MONTH:
		month--
		if month < 1 && year > dc.minYear {
			month = 12
			year--
		}
	case DATE_STEP_YEAR:
		year--
	}
	dc.Put(year, month)
}

// Next дата на крок назад
func (dc *dateControl) Next() {
	year, month := dc.Get()
	switch dc.step {
	case DATE_STEP_MONTH:
		month++
		if month > 12 && year < dc.maxYear {
			month = 1
			year++
		}
	case DATE_STEP_YEAR:
		year++
	}
	dc.Put(year, month)
}

// Put встановлює дату
func (dc *dateControl) Put(year, month int) {
	if dc.spinYear.GetValueAsInt() != year {
		dc.spinYear.SetValue(float64(year))
	}
	if dc.spinMonth.GetValueAsInt() != month {
		dc.spinMonth.SetValue(float64(month))
	}
}

// Get повертає дату
func (dc *dateControl) Get() (year, month int) {
	year = dc.spinYear.GetValueAsInt()
	month = dc.spinMonth.GetValueAsInt()
	return
}

// setSensitive встановлює активність блоку дати
func (dc *dateControl) SetSensitive(b bool) {
	dc.frame.SetSensitive(b)
}

// DateStep крок зміни дати. Month або Year
type DateStep int

// Константи для позначення кроку дати
const (
	DATE_STEP_MONTH DateStep = iota
	DATE_STEP_YEAR
)

// setStep встановлює крок зміни дати. Можливі варіанти Month та Year
func (dc *dateControl) SetStep(s DateStep) {
	dc.step = s
	switch s {
	case DATE_STEP_MONTH:
		dc.spinMonth.SetSensitive(true)
	case DATE_STEP_YEAR:
		dc.spinMonth.SetSensitive(false)
	}
}
