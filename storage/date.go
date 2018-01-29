package storage

import (
	"errors"
	"fmt"
)

// Date представляє дату у вигляді року та місяця.
type Date int

// NewDate створює дату.
func NewDate(year, month int) (Date, error) {
	if month < 1 || month > 12 {
		date, _ := NewDate(year, 1)
		return date, errors.New("невірна дата")
	}
	return Date(year*12 + month - 1), nil
}

// ParseDate створює дату із рядка в форматі "YYYY-MM".
func ParseDate(dateStr string) (Date, error) {
	var year, month int
	_, err := fmt.Sscanf(dateStr+"\n", "%04d-%02d\n", &year, &month)
	if err != nil {
		date, _ := NewDate(year, 1)
		return date, errors.New("невірний формат дати")
	}
	return NewDate(year, month)
}

// Add повертає нову дату збільшену на вказану кількість місяців.
func (d Date) Add(months int) Date {
	sum := int(d) + months
	return Date(sum)
}

// After перевіряє чи d старша за u.
func (d Date) After(u Date) bool {
	return int(d) > int(u)
}

// Before перевіряє чи d молодша за u.
func (d Date) Before(u Date) bool {
	return int(d) < int(u)
}

// Equal порівнює дві дати на рівність.
func (d Date) Equal(u Date) bool {
	return int(d) == int(u)
}

// Sub повертає різницю d - u.
func (d Date) Sub(u Date) int {
	return int(d) - int(u)
}

// Month повертає номер місяця.
func (d Date) Month() int {
	return int(d)%12 + 1
}

// Year повертає рік.
func (d Date) Year() int {
	return int(d) / 12
}

// StartYear повертає дату початку року.
func (d Date) StartYear() Date {
	newDate, _ := NewDate(d.Year(), 1)
	return newDate
}

// String повертає дату у вигляді рядка, в форматі YYYY-MM.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d", d.Year(), d.Month())
}

// timestring повертає дату в форматі дати бази даних.
func (d Date) timestring() string {
	return fmt.Sprintf("%04d-%02d-01", d.Year(), d.Month())
}

// NameMonth повертає назву місяця.
func (d Date) MonthName() string {
	var months = [...]string{
		"січень",
		"лютий",
		"березень",
		"квітень",
		"травень",
		"червень",
		"липень",
		"серпень",
		"вересень",
		"жовтень",
		"листопад",
		"грудень",
	}
	return months[d.Month()-1]
}

// GenitivNameMonth повертає назву місяця в родовому відмінку.
func (d Date) MonthNameGenitiv() string {
	var months = [...]string{
		"січня",
		"лютого",
		"березня",
		"квітня",
		"травня",
		"червня",
		"липня",
		"серпня",
		"вересня",
		"жовтня",
		"листопада",
		"грудня",
	}
	return months[d.Month()-1]
}
