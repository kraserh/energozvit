package gui

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/kraserh/energozvit/dbvisual"
	"github.com/kraserh/energozvit/storage"
)

//
func Start(db *storage.DB) {
	gtk.Init(nil)
	tw := dbvisual.NewTopWindow("ЕнергоЗвіт")
	tw.DB = db

	// Main menu
	mainMenu := []dbvisual.MenuItem{
		dbvisual.MenuItem{"Розрахунковий звіт", onMReportClick},
		dbvisual.MenuItem{"Технічний звіт", onPReortClick},
		dbvisual.MenuItem{"Лічильники", onMetersClick},
		dbvisual.MenuItem{"Тех. точки", onPartsClick},
		dbvisual.MenuItem{"Ліміти", onLimitsClick},
		dbvisual.MenuItem{"Друк", onPrintClick},
		dbvisual.MenuItem{"Вихід", tw.Quit},
	}
	tw.SetMenu(mainMenu)
	tw.Start()
}

// onReportClick виводе розрахунковий звіт
func onMReportClick() {
}

// onSubReportClick виводе технічний звіт
func onPReortClick() {
}

// onMeterClick виводе лічильники
func onMetersClick() {
}

// onSubPlaceClick виводе технічні точки обліку
func onPartsClick() {

}

// onLimitsClick виводе ліміти
func onLimitsClick() {

}

// onPrintClick друкує звіти
func onPrintClick() {

}
