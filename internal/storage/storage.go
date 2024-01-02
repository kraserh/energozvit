package storage

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// Версія бази даних яку підтримує ця програма.
const DBVERSION = 1

//go:embed schema.sql
var schema string

type Storage struct {
	*sql.DB
	filepath string
}

// Create створює нову базу даних.
func Create(filepath string, firstDate time.Time) error {
	// Перевірка існування файла бази даних
	var err error
	_, err = os.Stat(filepath)
	if err == nil {
		return errors.New("база даних вже існує")
	}

	// Створення бази даних
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	// Створення схеми бази даних
	_, err = db.Exec(schema)
	if err != nil {
		return (err)
	}

	// Початкова дата
	date := dateToString(firstDate)
	stmtService := "INSERT OR REPLACE INTO service VALUES (?, ?)"
	_, err = db.Exec(stmtService, "next_date", date)
	if err != nil {
		return (err)
	}

	// Закриття бази даних
	err = db.Close()
	if err != nil {
		return (err)
	}
	return nil
}

// Open відкриває базу даних.
func Open(filepath string) (*Storage, error) {
	// Перевірка існування файла бази даних
	var err error
	_, err = os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// Відкриття бази даних
	stor := new(Storage)
	stor.DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	stor.filepath = filepath

	// Set foreign keys
	_, err = stor.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Формат дати в базі даних
	sqlite3.SQLiteTimestampFormats = []string{DateLayout}

	// Перевірка версії
	if stor.GetVersion() != DBVERSION {
		err := errors.New("не підтримувана версія бази даних")
		return nil, err
	}
	return stor, nil
}

// Close закриває базу даних.
func (stor *Storage) Close() {
	stor.DB.Close()
}

// GetFilepath повертає шлях до бази даних.
func (stor *Storage) GetFilepath() string {
	return stor.filepath
}

// GetVersion повертає версію бази даних.
func (stor *Storage) GetVersion() int {
	row := stor.QueryRow("PRAGMA user_version")
	var version int
	err := row.Scan(&version)
	if err != nil {
		panic(err)
	}
	return version
}

//-------------------------- METER FUNCTIONS ---------------------------

type Meter struct {
	id         int64
	Substation int
	Eic        string
	Name       string
	Model      string
	Year       int
	Serial     string
	Digits     int
	Ratio      int
}

// GetActiveMeters повертає діючі лічильники.
func (stor *Storage) GetActiveMeters() []*Meter {
	queryActiveMeters := ` 
	SELECT meter_id,
	       ifnull(substation, 0),
	       ifnull(eic, ''),
	       name,
	       ifnull(model, ''),
	       ifnull(year, 0),
	       serial,
	       digits,
	       ratio
	  FROM meters JOIN places USING(place_id)
	 WHERE active = true
	 ORDER BY name, meter_id
	`
	rows, err := stor.Query(queryActiveMeters)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	meters := make([]*Meter, 0)
	for rows.Next() {
		meter := new(Meter)
		err := rows.Scan(&meter.id, &meter.Substation,
			&meter.Eic, &meter.Name, &meter.Model,
			&meter.Year, &meter.Serial, &meter.Digits,
			&meter.Ratio)
		if err != nil {
			panic(err)
		}
		meters = append(meters, meter)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return meters
}

// AddMeter додає лічильник з початковими показниками
func (stor *Storage) AddMeter(meter *Meter, kwh []int) error {
	// Початок транзакції
	tx, err := stor.Begin()
	if err != nil {
		panic(err)
	}

	// Додати точку обліку, якщо такої нема
	err = addPlaceIfNotExists(tx, meter)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Додати лічильник
	err = addMeter(tx, meter)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Додати початкові показники
	err = addFirstKwh(tx, meter, kwh)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Кінець транзакції
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	return nil
}

// addPlaceIfNotExists додає точку обліку, якщо таке імʼя відсутнє.
func addPlaceIfNotExists(tx *sql.Tx, meter *Meter) error {
	// Перевірка існування точки обліку
	queryPlaceExists := `
	SELECT EXISTS (
	       SELECT place_id
	         FROM places
	        WHERE name = ?)
	`
	var placeExists bool
	row := tx.QueryRow(queryPlaceExists, meter.Name)
	err := row.Scan(&placeExists)
	if err != nil {
		panic(err)
	}
	if placeExists {
		return nil
	}

	// Додавання точки обліку
	stmtAddPlace := `
	INSERT INTO places (
		substation,
		eic,
		name)
	VALUES (nullif(?, 0), nullif(?, ''), ?)
	`
	_, err = tx.Exec(stmtAddPlace, meter.Substation, meter.Eic,
		meter.Name)
	return err
}

// addMeter додає лічильник.
func addMeter(tx *sql.Tx, meter *Meter) error {
	stmtAddMeter := `
	INSERT INTO meters (
	       place_id,
	       active,
	       model,
	       year,
	       serial,
	       digits,
	       ratio)
	VALUES ((SELECT place_id FROM places WHERE name = ?),
	       true,
	       nullif(?, ''), nullif(?, 0), ?, ?, ?)
	`
	result, err := tx.Exec(stmtAddMeter, meter.Name,
		meter.Model, meter.Year, meter.Serial,
		meter.Digits, meter.Ratio)
	if err != nil {
		return err
	}
	meter.id, err = result.LastInsertId()
	if err != nil {
		panic(err)
	}
	return nil
}

// addFirstKwh додає початкові показники лічильника.
func addFirstKwh(tx *sql.Tx, meter *Meter, kwh []int) error {
	stmtAddKwh := `
	INSERT OR REPLACE INTO readings (rdate, meter_id, zone, kwh)
	VALUES (date((SELECT value
                        FROM service
                       WHERE skey == 'next_date'),
		'start of month', '-1 month'), ?, ?, ?)
	`
	stmt, err := tx.Prepare(stmtAddKwh)
	if err != nil {
		panic(err)
	}
	if len(kwh) == 0 {
		return errors.New("Не вказано початкові показники")
	}
	for i, v := range kwh {
		_, err = stmt.Exec(meter.id, i+1, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateMeter оновлює лічильник і точку обліку.
func (stor *Storage) UpdateMeter(em *Meter) error {
	//TODO
	return errors.New("Функціонал поки не реалізований")
}

var ErrMissingMeter = errors.New("missing meter")

// RemoveMeter видаляє лічильник.
func (stor *Storage) RemoveMeter(meter *Meter) error {
	var err error
	if meter == nil || meter.id == 0 {
		return ErrMissingMeter
	}
	stmtRemoveMeter := `
	UPDATE meters
	   SET active = false
	 WHERE meter_id = ?
	`
	_, err = stor.Exec(stmtRemoveMeter, meter.id)
	if err == nil {
		meter.id = 0
	}
	return err
}

//-------------------------- REPORT FUNCTIONS --------------------------

type Report struct {
	*Meter
	Zone       int
	CurKwh     int
	PreKwh     int
	Diff       int
	Energy     int
	Annotation string
}

// GetReports повертає звіт за вказану дату.
func (stor *Storage) GetReports(date time.Time) []*Report {
	queryReports := `
	SELECT meter_id,
               ifnull(substation, 0),
	       ifnull(eic, ''),
	       name,
	       ifnull(model, ''),
	       ifnull(year, 0),
	       serial,
	       digits,
	       ratio,
	       zone,
	       cur_kwh,
	       pre_kwh,
	       diff,
	       energy,
	       ifnull(annotation, '')
	  FROM reports
	 WHERE rdate = ?
	 ORDER BY name, meter_id, zone
	`
	rows, err := stor.Query(queryReports, date)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	return scanToReports(rows)
}

// GetNextReports повертає форму для введення показників. Дату можна
// прочитати функцією GetNextDate. В кожному рядку потрібно заповнити
// поле CurKwh після чого викликати функцію SaveReports.
func (stor *Storage) GetNextReports() []*Report {
	queryNextReports := ` 
	SELECT meter_id,
               ifnull(substation, 0),
	       ifnull(eic, ''),
               name,
	       ifnull(model, ''),
	       ifnull(year, 0),
	       serial,
	       digits,
	       ratio,
	       zone,
	       ifnull(cur_kwh, pre_kwh),
	       pre_kwh,
	       ifnull(diff, 0),
	       ifnull(energy, 0),
	       ifnull(annotation, '')
	  FROM next_reports
	 ORDER BY name, meter_id, zone
	`
	rows, err := stor.Query(queryNextReports)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	return scanToReports(rows)
}

// scanToReports сканує рядки бази даних в масив Report.
func scanToReports(rows *sql.Rows) []*Report {
	reports := make([]*Report, 0)
	for rows.Next() {
		report := new(Report)
		report.Meter = new(Meter)
		err := rows.Scan(&report.id, &report.Substation,
			&report.Eic, &report.Name, &report.Model,
			&report.Year, &report.Serial, &report.Digits,
			&report.Ratio, &report.Zone, &report.CurKwh,
			&report.PreKwh, &report.Diff, &report.Energy,
			&report.Annotation)
		if err != nil {
			panic(err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return reports
}

// Calculate робе підрахунки в звіті.
func (report *Report) Calculate() {
	diff := report.CurKwh - report.PreKwh
	if diff < 0 {
		diff = diff + int(math.Pow10(report.Digits))
	}
	report.Diff = diff
	report.Energy = report.Diff * report.Ratio
}

// SaveReports зберігає звіт до бази даних.
func (stor *Storage) SaveReports(reports []*Report) error {
	stmtUpdateNextReports := `
	UPDATE next_reports
	   SET cur_kwh = ?,
	       annotation = ?
	 WHERE meter_id = ? AND zone = ?
	`
	stmt, err := stor.Prepare(stmtUpdateNextReports)
	if err != nil {
		panic(err)
	}
	for _, report := range reports {
		report.Calculate()
		_, err := stmt.Exec(report.CurKwh, report.Annotation,
			report.id, report.Zone)
		if err != nil {
			return err
		}
	}
	return stor.gotoNextDate()
}

// gotoNextDate підтверджує що всі показники введені і можна переходити
// до слідуючої дати.
func (stor *Storage) gotoNextDate() error {
	stmtgotoNextDate := `
	UPDATE service
	   SET value = 1
	 WHERE skey = 'goto_next_date'
	`
	_, err := stor.Exec(stmtgotoNextDate)
	return err
}

// GetTotal повертає суму витраченої енергії за вказану дату, по вказаним
// точкам обліку.
func (stor *Storage) GetTotal(from, to time.Time, name ...string) int {
	queryTotal := `
	SELECT total(energy)
	  FROM reports
	 WHERE (rdate BETWEEN ? AND ?)`
	var nameStr string
	if len(name) > 0 {
		nameStr = fmt.Sprintf(" AND (name IN ('%s'))",
			strings.Join(name, "', '"))
	}
	var total int
	err := stor.QueryRow(queryTotal+nameStr, from, to).
		Scan(&total)
	if err != nil {
		panic(err)
	}
	return total
}

// GetNextTotal повертає суму витраченої енергії заданого звіту, плюс
// сума енргії видалених лічильників за поточну дату.
func (stor *Storage) GetNextTotal(reports []*Report) int {
	var total int
	for _, row := range reports {
		row.Calculate()
		total = total + row.Energy
	}

	// вибираємо видалені лічильники
	queryTotal := `
	SELECT total(energy)
	  FROM reports JOIN meters USING (meter_id)
	 WHERE rdate = (SELECT value FROM service WHERE skey = 'next_date')
	   AND active = false`
	var totalForNotActive int
	err := stor.QueryRow(queryTotal).
		Scan(&totalForNotActive)
	if err != nil {
		panic(err)
	}
	return total + totalForNotActive
}

//-------------------------- DATE  FUNCTIONS ---------------------------

const DateLayout = "2006-01-02"

// MakeDate створює дату при вказанні року та місяця.
func MakeDate(year, month int) time.Time {
	date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0,
		time.UTC)
	return date
}

// DateParse створює дату з рядка формату "2006-01".
func DateParse(value string) (time.Time, error) {
	date, err := time.Parse(DateLayout, fmt.Sprintf("%s-01", value))
	return date, err
}

// GetNextDate повертає дату наступного звіту.
func (stor *Storage) GetNextDate() time.Time {
	queryNextDate := `
	SELECT value
	  FROM service
	 WHERE skey = 'next_date'`
	row := stor.QueryRow(queryNextDate)
	var nextDate string
	err := row.Scan(&nextDate)
	if err != nil {
		panic(err)
	}

	date, err := stringToDate(nextDate)
	if err != nil {
		panic(err)
	}
	return date
}

// Дату перетворити в рядок формату "2006-01-02"
func dateToString(date time.Time) string {
	return date.Format(DateLayout)
}

// Рядок формату "2006-01-02" перетворити а дату
func stringToDate(str string) (time.Time, error) {
	return time.Parse(DateLayout, str)
}

//-------------------------- QUERY FUNCTIONS ---------------------------

// QueryLines виконує запит до бази даних. Повертає масив рядків.
func (stor *Storage) QueryLines(query string, args ...any) ([][]string, error) {
	return stor.query(false, query, args...)
}

// QueryLine виконує запит до бази даних. Повертає рядок.
func (stor *Storage) QueryLine(query string, args ...any) ([]string, error) {
	var row []string
	rows, err := stor.query(true, query, args...)
	if err == nil && len(rows) > 0 {
		row = rows[0]
	}
	return row, err
}

// query виконує запит до бази даних. Повертає масив рядків. Якщо
// встановлено oneLine, то результатом є тільки перший рядок.
func (stor *Storage) query(oneLine bool, query string, args ...any) ([][]string, error) {
	// Створення запиту.
	rows, err := stor.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обчислення кількості колонок.
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	col := len(columns)

	// Підготовка масивів
	result := make([][]string, 0)
	rowPtr := make([]any, col)
	row := make([]sql.NullString, col)
	for i := 0; i < col; i++ {
		rowPtr[i] = &row[i]
	}

	// Читання результату.
	for rows.Next() {
		err := rows.Scan(rowPtr...)
		if err != nil {
			return nil, err
		}
		result = append(result, nullStringsToStrings(row))
		if oneLine {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// nullStringsToStrings перетворює масив NullString в масив string.
func nullStringsToStrings(nullStrs []sql.NullString) []string {
	strs := make([]string, len(nullStrs))
	for i, nullStr := range nullStrs {
		if nullStr.Valid {
			strs[i] = nullStr.String
		}
	}
	return strs
}
