package storage

import (
	"os"
	"testing"
)

const (
	initDate = "2017-01"
	dbName   = "../example/db_test.ez.db"
)

var db *DB

////////////////////////////////// Date //////////////////////////////////////

func TestNewDate(t *testing.T) {
	if _, err := NewDate(2016, 1); err != nil {
		t.Error(err)
	}
	if _, err := NewDate(2016, 13); err == nil {
		t.Error(err)
	}
	if _, err := NewDate(2016, 0); err == nil {
		t.Error(err)
	}
}

func TestParseDate(t *testing.T) {
	if _, err := ParseDate("2016-1"); err != nil {
		t.Error(err)
	}
	if _, err := ParseDate("2016-01"); err != nil {
		t.Error(err)
	}
	if _, err := ParseDate("2016"); err == nil {
		t.Error(err)
	}
}

func TestAfterBeforeEqual(t *testing.T) {
	d1, _ := NewDate(2016, 7)
	d2, _ := NewDate(2000, 1)
	if !d1.After(d2) {
		t.Error()
	}
	if d1.Before(d2) {
		t.Error()
	}
	if !d1.Equal(d1) && d1.Equal(d2) {
		t.Error()
	}
}

func TestAdd(t *testing.T) {
	d, _ := NewDate(2016, 1)
	if u, _ := NewDate(2016, 2); !d.Add(1).Equal(u) {
		t.Error()
	}
	if u, _ := NewDate(2015, 12); !d.Add(-1).Equal(u) {
		t.Error()
	}
}

func TestSub(t *testing.T) {
	d, _ := NewDate(2015, 11)
	if u, _ := NewDate(2016, 7); u.Sub(d) != 8 {
		t.Error()
	}
}

func TestYearMonth(t *testing.T) {
	d, _ := NewDate(2016, 7)
	if d.Month() != 7 {
		t.Error()
	}
	if d.Year() != 2016 {
		t.Error()
	}
}

func TestNameMonth(t *testing.T) {
	d, _ := NewDate(2016, 7)
	if d.MonthName() != "липень" {
		t.Error()
	}
	if d.MonthNameGenitiv() != "липня" {
		t.Error()
	}
}

////////////////////////////////// Main //////////////////////////////////////

func TestCreateDB(t *testing.T) {
	os.Remove(dbName)
	var date Date
	var err error
	if date, err = ParseDate(initDate); err != nil {
		t.Error(err)
	}
	db = CreateDB(dbName, date)
}

func TestCloseDB(t *testing.T) {
	err := db.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestOpenDB(t *testing.T) {
	db = OpenDB(dbName)
}

//////////////////////////////////////////////////////////////////////////////

func TestLocationInsert(t *testing.T) {
	l1 := new(Location)
	l1.Substation = 222
	l1.Lname = "Location 1"
	err := db.LocationInsert(l1)
	if err != nil {
		t.Error(err)
	}
	l2 := new(Location)
	l2.Substation = 208
	l2.Lname = "Location 2"
	err = db.LocationInsert(l2)
	if err != nil {
		t.Error(err)
	}
}

func TestMeterInsert(t *testing.T) {
	m1 := new(Meter)
	m1.Lname = "Location 1"
	m1.Model = "НІК 2102"
	m1.Number = "12345"
	m1.Limval = 1000000
	m1.Ratio = 1
	m1.Numzones = 2
	err := db.MeterInsert(m1)
	if err != nil {
		t.Error(err)
	}
	m2 := new(Meter)
	m2.Lname = "Location 2"
	m2.Model = "НІК 2102"
	m2.Number = "98765"
	m2.Limval = 1000000
	m2.Ratio = 10
	m2.Numzones = 1
	err = db.MeterInsert(m2)
	if err != nil {
		t.Error(err)
	}
}

func TestMlogInsert(t *testing.T) {
	model := "НІК 2102"
	number := "12345"
	m1 := []int{5, 10}
	err := db.MlogInsert(model, number, m1, "")
	if err != nil {
		t.Error(err)
	}
	m1 = []int{1000, 2000}
	err = db.MlogInsert(model, number, m1, "Start")
	if err != nil {
		t.Error(err)
	}

	model = "НІК 2102"
	number = "98765"
	m2 := []int{999999}
	err = db.MlogInsert(model, number, m2, "")
	if err != nil {
		t.Error(err)
	}
	m2 = []int{100}
	err = db.MlogInsert(model, number, m2, "Start")
	if err != nil {
		t.Error(err)
	}

	err = db.MlogCommit()
	if err != nil {
		t.Error(err)
	}
}

func TestMlogUpdate(t *testing.T) {
	model := "НІК 2102"
	number := "12345"
	m1 := []int{3000, 4000}
	err := db.MlogUpdate(model, number, m1, "Updated")
	if err != nil {
		t.Error(err)
	}
}

func TestPartInsert(t *testing.T) {
	p1 := new(Part)
	p1.Lname = "Location 1"
	p1.Pname = "Part 1"
	err := db.PartInsert(p1)
	if err != nil {
		t.Error(err)
	}
	p2 := new(Part)
	p2.Lname = "Location 1"
	p2.Pname = "Part 2"
	err = db.PartInsert(p2)
	if err != nil {
		t.Error(err)
	}
}

func TestPlogInsert(t *testing.T) {
	err := db.PlogInsert("Part 1", 6000)
	if err != nil {
		t.Error(err)
	}
	err = db.PlogInsert("Part 2", 980)
	if err != nil {
		t.Error(err)
	}
	err = db.PlogCommit()
	if err != nil {
		t.Error(err)
	}
}

func TestPlogUpdate(t *testing.T) {
	err := db.PlogUpdate("Part 2", 985)
	if err != nil {
		t.Error(err)
	}
}

func TestLimitInsert(t *testing.T) {
	date, err := ParseDate(initDate)
	if err != nil {
		t.Error(err)
	}
	err = db.LimitInsert(222, 10000, date)
	if err != nil {
		t.Error(err)
	}
	err = db.LimitInsert(208, 555, date)
	if err != nil {
		t.Error(err)
	}
	date = date.Add(1)
	err = db.LimitInsert(222, 8000, date)
	if err != nil {
		t.Error(err)
	}
	err = db.LimitInsert(208, 1000, date)
	if err != nil {
		t.Error(err)
	}
}

func TestLimitUpdate(t *testing.T) {
	date, err := ParseDate(initDate)
	if err != nil {
		t.Error(err)
	}
	err = db.LimitUpdate(208, 500, date)
	if err != nil {
		t.Error(err)
	}
}
