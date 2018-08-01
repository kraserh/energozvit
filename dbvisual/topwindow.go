package dbvisual

import (
	"log"
	"strings"

	"github.com/gotk3/gotk3/gtk"
	"github.com/kraserh/energozvit/storage"
)

// TopWindow вікно верхнього рівня для показу контенту та керування ним
type TopWindow struct {
	*gtk.Window
	DB *storage.DB
	//	Date *dateControl
	//	Content     *Content
	QuitFunc   func()
	lockDialog bool
	menuBox    *gtk.Box
	dateBox    *gtk.Box
	contentBox *gtk.Box
}

// NewTopWindow створює основне вікно
func NewTopWindow(title string) *TopWindow {
	tw := new(TopWindow)
	var err error
	tw.Window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	check(err)
	tw.Connect("destroy", tw.Quit)
	tw.SetTitle(title)

	// Розмір та позиція вікна за замовчуванням
	tw.SetDefaultSize(1100, 600)
	tw.SetPosition(gtk.WIN_POS_CENTER)

	// Ділим вікно по вертикалі
	rootBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	check(err)
	tw.Add(rootBox)

	// Додаєм бокс головного меню та дати
	menuAndDateBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	rootBox.PackStart(menuAndDateBox, false, false, 5)
	separator, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	check(err)
	rootBox.PackStart(separator, false, false, 0)

	// Бокс головного меню
	tw.menuBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	menuAndDateBox.PackStart(tw.menuBox, false, false, 0)

	// Бокс дати
	tw.dateBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	menuAndDateBox.PackEnd(tw.dateBox, false, false, 10)

	// Блок даних
	tw.contentBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	check(err)
	rootBox.PackEnd(tw.contentBox, true, true, 5)

	return tw
}

// Start запускає графічний інтерфейс.
func (tw *TopWindow) Start() {
	tw.ShowAll()
	gtk.Main()
}

// Quit закриває всі відкриті вікна.
func (tw *TopWindow) Quit() {
	gtk.MainQuit()
	if tw.QuitFunc != nil {
		tw.QuitFunc()
	}
}

////////////////////////////////////////////////////////////////////////////////

// MenuItem пункт головного меню
type MenuItem struct {
	Label    string
	Callback func()
}

// SetMenu встановлює головне меню
func (tw *TopWindow) SetMenu(menu []MenuItem) {
	for _, item := range menu {
		button, err := gtk.ButtonNewWithLabel(item.Label)
		check(err)
		button.Connect("clicked", item.Callback)
		tw.menuBox.PackStart(button, true, true, 1)
	}
}

////////////////////////////////////////////////////////////////////////////////

// InfoDialog відкриває вікно для повідомлення.
func (tw *TopWindow) InfoDialog() {

}

// ConfirmDialog відкриває вікно для підтвердження.
func (tw *TopWindow) ConfirmDialog() {

}

// EntryDialog відкриває вікно для вводу одного значення.
func (tw *TopWindow) EntryDialog() {

}

// ListDialog відкриває вікно для редагування таблиці.
func (tw *TopWindow) ListDialog() {

}

// check перериває програму якщо err містить помилку. Якщо вказано msg,
// то заміть err виводиться вказане повідомлення.
func check(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			log.Fatal(strings.Join(msg, " "))
		} else {
			log.Fatal(err)
		}
	}
}
