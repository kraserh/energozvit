package storage

import (
	_ "embed"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

//go:embed data_test.sql
var data_test string

//-------------------------- Initial database --------------------------

func createDatabase(t *testing.T) *Storage {
	tempDir := t.TempDir()
	dbName := t.Name() + ".sqlite"
	dbPath := path.Join(tempDir, dbName)

	startDate := time.Now()
	err := Create(dbPath, startDate)
	if err != nil {
		t.Fatalf("create database: %s", err)
	}

	stor, err := Open(dbPath)
	if err != nil {
		t.Fatalf("create database: %s", err)
	}

	diff := cmp.Diff(DBVERSION, stor.GetVersion())
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	_, err = stor.Exec(data_test)
	if err != nil {
		t.Fatalf("create database: %s", err)
	}
	return stor
}

//------------------------ Meter Function Tests ------------------------

func TestGetActiveMeters(t *testing.T) {
	stor := createDatabase(t)
	meters := stor.GetActiveMeters()
	want := []*Meter{
		{1, 208, "1234567890abcdef", "Госпдвір",
			"НІК2301АП1", 2020, "344848", 4, 40},
		{3, 205, "", "Контора",
			"", 0, "001930", 5, 1},
	}

	if len(want) != len(meters) {
		t.Errorf("the number of meters want %d, got %d",
			len(want), len(meters))
	}

	for i := 0; i < len(want); i++ {
		diff := cmp.Diff(want[i], meters[i],
			cmp.AllowUnexported(Meter{}))
		if diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestAddMeter(t *testing.T) {
	// Додавання лічильника.
	stor := createDatabase(t)
	var meter = &Meter{
		Name:   "Office",
		Serial: "12345678",
		Digits: 4,
		Ratio:  10,
	}
	err := stor.AddMeter(meter, []int{9999, 1111})
	if err != nil {
		t.Fatalf("meter not added: %s", err)
	}

	// Перевірка початкових показників.
	query := `
	SELECT kwh
	  FROM readings
	 WHERE rdate = ? AND meter_id = ? AND zone = ?`
	date := "2022-02-01"
	zone := 2
	got := "1111"
	row, err := stor.QueryLine(query, date, meter.id, zone)
	if err != nil || len(row) != 1 {
		t.Error("first reading error")
	} else if row[0] != got {
		t.Errorf("first reading wont %s, got %s", got, row[0])
	}

	// Перевірка лічильника.
	meters := stor.GetActiveMeters()
	for _, gotMeter := range meters {
		if gotMeter.Name == meter.Name {
			diff := cmp.Diff(meter, gotMeter,
				cmp.AllowUnexported(Meter{}))
			if diff != "" {
				t.Errorf("mismatch (-want +got):\n%s",
					diff)
			}
			return
		}
	}
	t.Error("meter not found in active meters")
}

func TestRemoveMeter(t *testing.T) {
	stor := createDatabase(t)

	// Отримання списку активних лічильників
	meters := stor.GetActiveMeters()
	if len(meters) != 2 {
		t.Fatal("error getting list of meters")
	}

	// Видалення лічильника
	err := stor.RemoveMeter(meters[0])
	if err != nil {
		t.Fatalf("remove meter error: %s", err)
	} else if meters[0].id != 0 {
		t.Error("meter_id after remove must be zero")
	}

	// Видалення видаленого лічильника
	err = stor.RemoveMeter(meters[0])
	if err != ErrMissingMeter {
		t.Errorf("remove a removed meter error: %s", err)
	}

	// Отримання списку активних лічильників
	meters = stor.GetActiveMeters()
	if len(meters) != 1 {
		t.Fatal("could not remove the meter")
	}

	serialMust := []string{"001930"}
	for i, meter := range meters {
		if meter.Serial != serialMust[i] {
			t.Errorf("meter %s expected serial %s, got %s",
				meter.Name, serialMust[i], meter.Serial)
		}
	}
}

//----------------------- Reports Function Tests -----------------------

func TestGetReports(t *testing.T) {
	stor := createDatabase(t)
	date, err := stringToDate("2022-02-01")
	if err != nil {
		t.Fatal(err)
	}
	reports := stor.GetReports(date)
	want := &Report{
		&Meter{4, 220, "", "АВМ", "НІК2102-02", 2022, "E12345",
			4, 40}, 1, 7481, 7455, 26, 1040, ""}

	if len(reports) == 0 {
		t.Error("reports empty")
	}

	diff := cmp.Diff(want, reports[0],
		cmp.AllowUnexported(Meter{}))
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestNextReports(t *testing.T) {
	stor := createDatabase(t)

	// Отримуєм звіти на наступний період і заповнюєм CurKwh.
	reports := stor.GetNextReports()
	for _, report := range reports {
		report.CurKwh += 10
	}

	// Зберігаєм звіти до бази даних.
	err := stor.SaveReports(reports)
	if err != nil {
		t.Error("save reports error")
	}

	// Знову отримуєм звіти на наступний період.
	reports = stor.GetNextReports()
	if len(reports) == 0 {
		t.Error("next reports empty")
	}

	// Перевірка звіту.
	want := &Report{
		&Meter{1, 208, "1234567890abcdef", "Госпдвір", "НІК2301АП1",
			2020, "344848", 4, 40}, 1, 74, 74, 0, 0, ""}
	diff := cmp.Diff(want, reports[0],
		cmp.AllowUnexported(Meter{}))
	if diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	// Перевірка наступної дати
	date := stor.GetNextDate()
	dateWant, err := stringToDate("2022-04-01")
	if err != nil {
		t.Error(err)
	}
	if !date.Equal(dateWant) {
		t.Errorf("next date want %s, got %s",
			dateWant.Format(DateLayout),
			date.Format(DateLayout))
	}
}

func TestCalculate(t *testing.T) {
	report := new(Report)
	report.Meter = new(Meter)
	report.CurKwh = 1
	report.PreKwh = 9999
	report.Digits = 4
	report.Ratio = 10
	report.Calculate()
	if report.Diff != 2 && report.Energy != 20 {
		t.Error("calculate report error")
	}
}

func TestGetTotal(t *testing.T) {
	stor := createDatabase(t)

	// Спожита енергія за вказану дату.
	from, err1 := stringToDate("2021-12-01")
	to, err2 := stringToDate("2022-02-01")
	if err1 != nil || err2 != nil {
		t.Fatal(err1, err2)
	}
	total := stor.GetTotal(from, to)
	want := 30208
	if total != want {
		t.Errorf("GetTotal() want %d, got %d", want, total)
	}

	// Спожита енергія за вказану дату по вказаній точкі обліку.
	total = stor.GetTotal(from, to, "АВМ")
	want = 4280
	if total != want {
		t.Errorf("GetTotal() want %d, got %d", want, total)
	}
}

func TestGetNextTotal(t *testing.T) {
	stor := createDatabase(t)

	//Спожита енергія по видаленим лічильникам.
	next := stor.GetNextTotal(nil)
	want := 4000
	if next != want {
		t.Errorf("GetNextTotal() want %d, got %d", want, next)
	}

	// Енергія по видаленим лічильникам разом із вказним звітом.
	report := new(Report)
	report.Meter = new(Meter)
	report.CurKwh = 1100
	report.PreKwh = 100
	report.Digits = 4
	report.Ratio = 1
	reports := []*Report{report}
	next = stor.GetNextTotal(reports)
	want = 5000
	if next != want {
		t.Errorf("GetNextTotal(r) want %d, got %d", want, next)
	}
}

//------------------------ Query Function Tests ------------------------

func TestQueryLines(t *testing.T) {
	stor := createDatabase(t)
	query := `SELECT substation, eic, name FROM places ORDER BY name`
	want := [][]string{
		{"220", "", "АВМ"},
		{"208", "1234567890abcdef", "Госпдвір"},
		{"205", "", "Контора"},
	}

	rows, err := stor.QueryLines(query)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, rows); diff != "" {
		t.Errorf("QueryLines() mismatch (-want +got):\n%s", diff)
	}
}

func TestQueryLine(t *testing.T) {
	stor := createDatabase(t)
	query := `SELECT substation, eic, name FROM places WHERE name = ?`
	want := []string{"205", "", "Контора"}

	rows, err := stor.QueryLine(query, "Контора")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, rows); diff != "" {
		t.Errorf("QueryLine() mismatch (-want +got):\n%s", diff)
	}
}
