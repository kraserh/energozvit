package tui

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/kraserh/energozvit/internal/storage"
)

// Ширина поля вводу
const inputWidth = 32

type Content interface {
	GetName() string                          // ідентефікатор
	GetMenuName() string                      // пункт меню
	GetTitle() string                         // заголовок таблиці
	GetTable() *tview.Table                   // таблиця з даними
	GetCell(row, column int) *tview.TableCell // вміст комірки
	GetRowCount() int                         // кількість рядків
	GetColumnCount() int                      // кількість колонок
	GetKeybindingString() string              // опис клавіш дій
	NeedToSave() bool                         // потрібно збереження
	RereadTable()                             // оновлює таблицю
}

type Dialog interface {
	GetTitle() string              // заголовок
	GetPrimitive() tview.Primitive // контент
	GetBox() *tview.Box            // бокс
	SetOkFunc(func())              // натиснуто ОК
	SetCancelFunc(func())          // натиснуто відміна
}

type Tui struct {
	app      *tview.Application
	pages    *tview.Pages
	tabBar   *tview.TextView
	contents []Content
	stor     *storage.Storage
}

// Start запускає інтерфейс.
func Start(stor *storage.Storage) {
	// створюєм структуру інтерфейсу.
	t := &Tui{
		app:   tview.NewApplication(),
		pages: tview.NewPages(),
		tabBar: tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			SetWrap(false),
		stor: stor,
	}

	// додаєм сторінки і перемикаємо на першу сторінку.
	content := newContentNewReport(t)
	t.addContent(content)
	t.addContent(newContentReport(t))
	t.addContent(newContentMeters(t))
	t.switchToContent(content)

	// створюєм верхній рядок табів і показ сторінки.
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.tabBar, 1, 1, false).
		AddItem(t.pages, 0, 16, true)
	t.app.SetRoot(flex, true)

	// запускаєм інтерфейс.
	if err := t.app.Run(); err != nil {
		log.Fatal(err)
	}
}

// Stop зупиняє інтерфейс.
func (t *Tui) Stop() {
	for _, content := range t.contents {
		if content.NeedToSave() {
			message := fmt.Sprintf(
				"Не збережені дані в панелі \"%s\"",
				content.GetMenuName())
			t.Message(message)
			return
		}
	}
	t.app.Stop()
}

// addContent додає сторінку в якій розміщено вказаний контент.
func (t *Tui) addContent(content Content) {
	// розміщуєм таблицю і рядок-підсказку клавіш
	t.initTable(content)
	table := content.GetTable()
	keybindingString := t.setKeybinding(content)
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 16, true).
		AddItem(keybindingString, 1, 1, false)

	// додаєм сторінку.
	t.pages.AddPage(content.GetName(), flex, true, true)
	t.contents = append(t.contents, content)
	num := len(t.contents)
	fmt.Fprintf(t.tabBar, `  ["%s"]%d %s[""] `, content.GetName(),
		num, content.GetMenuName())
}

// initTable ініциалізує таблицю.
func (t *Tui) initTable(content Content) {
	table := content.GetTable().
		SetBorders(false).
		SetFixed(1, 0)
	table.SetBorder(true)
	t.updateTable(content)
}

// updateTable оновлює таблицю.
func (t *Tui) updateTable(content Content) {
	table := content.GetTable().
		Clear().
		Select(0, 0)
	table.SetTitle(content.GetTitle())
	rows := content.GetRowCount()
	cols := content.GetColumnCount()
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := content.GetCell(row, col)
			if row == 0 {
				// header
				color := tview.Styles.InverseTextColor
				cell.SetTextColor(color).
					SetSelectable(false).
					SetAlign(tview.AlignLeft).
					SetAttributes(tcell.AttrBold)
			}
			table.SetCell(row, col, cell)
		}
	}
}

// setKeybinding встановлює додаткову прив'язку клавіш, в добавок до вже
// встановленої до таблиці прив'язки. Повертає текст з описом клавіш.
func (t *Tui) setKeybinding(content Content) *tview.TextView {
	// прив'язка клавіш
	table := content.GetTable()
	tableKeybinding := table.GetInputCapture()
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			t.Stop()
		}

		num, err := strconv.Atoi(string(event.Rune()))
		if err == nil {
			if (num > 0) && (num <= len(t.contents)) {
				t.switchToContent(t.contents[num-1])
			}
		}
		return tableKeybinding(event)
	})

	// рядок підсказка
	generalKeybinding := "  q: Вихід"
	keybindingString := tview.NewTextView().
		SetText(content.GetKeybindingString() + generalKeybinding)
	return keybindingString
}

// switchToContent перемикає сторінки з контентом.
func (t *Tui) switchToContent(content Content) {
	t.pages.SwitchToPage(content.GetName())
	t.tabBar.Highlight(content.GetName()).ScrollToHighlight()
}

// addAndSwitchToDialog додає діалог і перемикає на нього. Замінює інший
// відкритий діалог.
func (t *Tui) addAndSwitchToDialog(dialog Dialog) {
	box := dialog.GetBox()
	box.SetBorder(true).
		SetTitle(dialog.GetTitle())
	parent, _ := t.pages.GetFrontPage()
	t.pages.SendToFront(parent)
	t.pages.AddAndSwitchToPage("dialog", dialog.GetPrimitive(), true)
}

// closeDialog закриває діалог і повертається на останню контент сторінку.
func (t *Tui) closeDialog(dialog Dialog) {
	t.pages.RemovePage("dialog")
	name, _ := t.pages.GetFrontPage()
	content, ok := t.searchContent(name)
	if ok {
		t.updateTable(content)
	}
}

// searchContent шукає контент по назві.
func (t *Tui) searchContent(name string) (Content, bool) {
	for _, content := range t.contents {
		if content.GetName() == name {
			return content, true
		}
	}
	return nil, false
}

// centeredPage виводе вікно вказаного розміру по центру.
func (t *Tui) centeredPage(name string, p tview.Primitive, width, height int) {
	grid := tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
	parent, _ := t.pages.GetFrontPage()
	t.pages.SendToFront(parent)
	t.pages.AddAndSwitchToPage(name, grid, true)
}

// ErrorShow виводе помилку.
func (t *Tui) ErrorShow(err error) {
	t.Message(fmt.Sprint(err))
}

// message виводе повідомлення.
func (t *Tui) Message(text string) {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) {
			t.pages.RemovePage("message")
		})
	t.centeredPage("message", modal, 80, 29)
}

// Confirm запитує підтвердження і оновлює таблицю.
func (t *Tui) Confirm(text string, okFunc func()) {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Відміна"}).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(index int, label string) {
			t.pages.RemovePage("message")
			if label == "OK" {
				okFunc()
				name, _ := t.pages.GetFrontPage()
				content, ok := t.searchContent(name)
				if ok {
					t.updateTable(content)
				}
			}
		})
	t.centeredPage("message", modal, 80, 29)
}

// Контроль вводу тільки цифр
func isNumber(_ string, lastChar rune) bool {
	return lastChar >= '0' && lastChar <= '9'
}
