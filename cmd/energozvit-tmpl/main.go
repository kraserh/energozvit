package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"

	"github.com/kraserh/energozvit/internal/storage"
)

var Version string

type tmpl struct {
	stor     *storage.Storage
	date     time.Time
	sortName []string
}

// Програма приймає шаблон із стандартного вводу і видає результат в
// стандартний вивід. Аргументи команди є імʼя бази даних та дата в
// форматі YYYY-MM
func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 3 {
		usageAndExit()
	}

	// база даних
	pathDB := os.Args[1]
	stor, err := storage.Open(pathDB)
	if err != nil {
		log.Fatal(err)
	}
	defer stor.Close()

	// дата
	yymm := os.Args[2]
	date, err := storage.DateParse(yymm)
	if err != nil {
		log.Fatal(err)
	}

	// обробка шаблону
	t := &tmpl{stor, date, nil}
	err = t.parse(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func usageAndExit() {
	fmt.Println("EnergoZvit templates handler")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Usage:\n  <template> | energozvit-tmpl db_file YYYY-MM > file\n")
	os.Exit(0)
}

// Обробка шаблону
func (t *tmpl) parse(in, out *os.File) error {
	// Функції які доступні в шаблонах
	funcs := template.FuncMap{
		"add":             func(x, y int) int { return x + y },
		"year":            t.date.Year,
		"month":           func() int { return int(t.date.Month()) },
		"monthName":       t.monthName,
		"query":           t.query,
		"reportMonth":     t.report,
		"sortReportMonth": t.setSortReport,
		"totalMonth":      t.totalMonth,
	}

	templateText, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	templ, err := template.New("main").
		Funcs(funcs).
		Parse(string(templateText))
	if err != nil {
		return err
	}

	err = templ.Execute(out, nil)
	if err != nil {
		return err
	}
	return nil
}

// Назви місяців
func (t *tmpl) monthName() string {
	monthMap := [13]string{
		"", "січень", "лютий", "березень", "квітень", "травень",
		"червень", "липень", "серпень", "вересень", "жовтень",
		"листопад", "грудень",
	}

	return monthMap[t.date.Month()]
}

// Запит до бази даних
func (t *tmpl) query(query string) [][]string {
	result, err := t.stor.QueryLines(query)
	if err != nil {
		panic(err)
	}
	return result
}

// Точка обліку
type Place struct {
	Substation int
	Eic        string
	Name       string
	Meters     []*Meter
	Lines      int // Кількість лічильників
}

// Лічильник
type Meter struct {
	Model  string
	Year   int
	Serial string
	Digits int
	Ratio  int
	Zones  []*Zone
	Lines  int // Кількість тарифних зон
}

// Тарифна зона
type Zone struct {
	CurKwh     int
	PrevKwh    int
	Diff       int
	Energy     int
	Annotation string
}

// Місячний звіт за вказану дату
func (t *tmpl) report() []*Place {
	reports := t.stor.GetReports(t.date)
	var places []*Place
	var place *Place
	var emeter *Meter
	for _, report := range reports {

		// точки обліку
		if place == nil || place.Name != report.Name {
			place = &Place{
				report.Substation,
				report.Eic,
				report.Name,
				[]*Meter{},
				0,
			}
			places = append(places, place)
		}

		// лічильники
		if emeter == nil || emeter.Serial != report.Serial {
			emeter = &Meter{
				report.Model,
				report.Year,
				report.Serial,
				report.Digits,
				report.Ratio,
				[]*Zone{},
				0,
			}
			place.Meters = append(place.Meters, emeter)
		}

		// тарифні зони
		zone := &Zone{
			report.CurKwh,
			report.PreKwh,
			report.Diff,
			report.Energy,
			report.Annotation,
		}
		emeter.Zones = append(emeter.Zones, zone)
		place.Lines++
		emeter.Lines++
	}
	return t.sortReport(places)
}

// Сортує точки обліку відповідно до t.sortName
func (t *tmpl) sortReport(places []*Place) []*Place {
	order := make(map[string]int)
	maxIdx := len(places)
	for i, name := range t.sortName {
		order[name] = maxIdx - i
	}
	sort.SliceStable(places, func(i, j int) bool {
		return order[places[i].Name] > order[places[j].Name]
	})
	return places
}

// Встановлює сортування точок обліку
func (t *tmpl) setSortReport(names ...string) string {
	t.sortName = names
	return ""
}

// Спожита потужність за місяць
func (t *tmpl) totalMonth() int {
	return t.stor.GetTotal(t.date, t.date)
}
