package storage

import (
	"database/sql"
	"errors"
)

///////////////////////////////////////////////////////////////////////////

type Location struct {
	Substation int
	Lname      string
	OldLname   string
}

func (db *DB) LocationInsert(l *Location) error {
	query := `
		INSERT INTO location (substation, lname, sdate)
		VALUES (?, ?, (SELECT value FROM stat WHERE key = 'mDate'))`
	_, err := db.Exec(query, l.Substation, l.Lname)
	return err
}

func (db *DB) LocationUpdate(l *Location) error {
	query := `
		UPDATE location
		SET substation = ?, lname = ?
		WHERE lname = ? AND edate IS NULL`
	_, err := db.Exec(query, l.Substation, l.Lname, l.OldLname)
	return err
}

func (db *DB) LocationDelete(lname string) error {
	query := `
		UPDATE location
		SET edate = date((SELECT value FROM stat WHERE key = 'mDate'),
			'-1 month', 'start of month')
		WHERE lname = ?`
	_, err := db.Exec(query, lname)
	return err
}

///////////////////////////////////////////////////////////////////////////

type Meter struct {
	Lname     string
	Model     string
	Number    string
	Limval    int
	Ratio     int
	Numzones  int
	OldModel  string
	OldNumber string
}

func (db *DB) MeterInsert(m *Meter) error {
	query := `
		INSERT INTO meters(idlocation, model, number, limval, 
			ratio, numzones, sdate)
		VALUES ((SELECT id FROM location
				WHERE lname = ? AND edate IS NULL),
			?, ?, ?, ?, ?, 
			(SELECT value FROM stat WHERE key = 'mDate'))`
	_, err := db.Exec(query, m.Lname, m.Model, m.Number, m.Limval,
		m.Ratio, m.Numzones)
	return err
}

func (db *DB) MeterUpdate(m *Meter) error {
	query := `
		UPDATE meters
		SET model = ?, number = ?, limval = ?, ratio = ?, numzones = ?
		WHERE model = ? AND number = ? AND edate IS NULL`
	_, err := db.Exec(query, m.Model, m.Number, m.Limval,
		m.Ratio, m.Numzones, m.OldModel, m.OldNumber)
	return err

}

func (db *DB) MeterDelete(model, number string) error {
	query := `
		UPDATE meters
		SET edate = date(
			(SELECT value FROM stat WHERE key = 'mDate'),
			'-1 month', 'start of month')
		WHERE model = ? AND number = ? AND edate IS NULL`
	_, err := db.Exec(query, model, number)
	return err
}

///////////////////////////////////////////////////////////////////////////

func (db *DB) MlogInsert(model, number string,
	values []int, comment string) error {
	//
	queryMeter := `
		SELECT id, numzones
		FROM meters
		WHERE model = ? AND number =? AND edate IS NULL`
	var id, numzones int
	err := db.QueryRow(queryMeter, model, number).Scan(&id, &numzones)
	if err == sql.ErrNoRows {
		return errors.New("відсутній лічильник")
	} else {
		check(err)
	}
	if numzones != len(values) {
		return errors.New("кількість показників і зон не збігається")
	}
	//
	queryInit := `
		SELECT init
		FROM mlog
		WHERE idmeter = ?`
	var init int
	var queryDate string
	err = db.QueryRow(queryInit, id).Scan(&init)
	if err == sql.ErrNoRows {
		init = 1
		queryDate = `date((SELECT value FROM stat WHERE key = 'mDate'), 
			'-1 month')`
	} else {
		check(err)
		init = 0
		queryDate = `(SELECT value FROM stat WHERE key = 'mDate')`
	}
	//
	query := `
		INSERT OR REPLACE
		INTO mlog (idmeter, zone, value, init, comment, date)
		VALUES ((SELECT id FROM meters
			WHERE model = ? AND number = ? AND edate IS NULL),
			?, ?, ?, ?, ` + queryDate + `)`
	tx, err := db.Begin()
	check(err)
	defer tx.Rollback()
	stmt, err := tx.Prepare(query)
	check(err)
	for zone, value := range values {
		_, err := stmt.Exec(model, number, zone+1, value, init, comment)
		if err != nil {
			stmt.Close()
			return err
		}
		comment = ""
	}
	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) MlogCommit() error {
	query := `
		UPDATE stat
		SET value = date((SELECT value FROM stat WHERE key = 'mDate'),
			'+1 month', 'start of month')
		WHERE key = 'mDate';`
	_, err := db.Exec(query)
	return err
}

func (db *DB) MlogUpdate(model, number string,
	values []int, comment string) error {
	query := `
		UPDATE mlog
		SET value = ?, comment = ?
		WHERE idmeter = 
			(SELECT id FROM meters 
			WHERE model = ? 
			AND number = ? 
			AND date((SELECT value FROM stat WHERE key = 'mDate'),
				'-1 month')
			BETWEEN sdate AND ifnull(edate, '9999-12-31')
		AND zone = ?
		AND date = date((SELECT value FROM stat WHERE key = 'mDate'), 
			'-1 month'))`
	tx, err := db.Begin()
	check(err)
	defer tx.Rollback()
	stmt, err := tx.Prepare(query)
	check(err)
	for zone, value := range values {
		_, err := stmt.Exec(value, comment, model, number, zone+1)
		if err != nil {
			stmt.Close()
			return err
		}
		comment = ""
	}
	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

///////////////////////////////////////////////////////////////////////////

type Part struct {
	Lname    string
	Pname    string
	OldPname string
}

func (db *DB) PartInsert(p *Part) error {
	query := `
		INSERT INTO parts (idlocation, pname, sdate)
		VALUES ((SELECT id FROM location
				WHERE lname = ? AND edate IS NULL), 
			?,
			(SELECT value FROM stat WHERE key = 'mDate'))`
	_, err := db.Exec(query, p.Lname, p.Pname)
	return err
}

func (db *DB) PartUpdate(p *Part) error {
	query := `
		UPDATE parts
		SET pname = ?
		WHERE pname = ? AND edate IS NULL`
	_, err := db.Exec(query, p.Pname, p.OldPname)
	return err
}

func (db *DB) PartDelete(pname string) error {
	query := `
		UPDATE parts
		SET edate = date((SELECT value FROM stat WHERE key = 'mDate'),
			'-1 month', 'start of month')
		WHERE pname = ?`
	_, err := db.Exec(query, pname)
	return err
}

///////////////////////////////////////////////////////////////////////////

func (db *DB) PlogInsert(pname string, energy int) error {
	query := `
		INSERT OR REPLACE
		INTO plog (idpart, energy, date)
		VALUES ((SELECT id FROM parts
			WHERE pname = ? AND edate IS NULL),
			?,
			(SELECT value FROM stat WHERE key = 'pDate'))`
	_, err := db.Exec(query, pname, energy)
	return err
}

func (db *DB) PlogCommit() error {
	query := `
		UPDATE stat
		SET value = date((SELECT value FROM stat WHERE key = 'pDate'),
			'+1 month', 'start of month')
		WHERE key = 'pDate'`
	_, err := db.Exec(query)
	return err
}

func (db *DB) PlogUpdate(pname string, energy int) error {
	query := `
		UPDATE plog
		SET energy = ?
		WHERE idpart = 
			(SELECT id FROM parts 
			WHERE pname = ? 
			AND date((SELECT value FROM stat WHERE key = 'pDate'),
				'-1 month')
			BETWEEN sdate AND ifnull(edate, '9999-12-31')
		AND date = date((SELECT value FROM stat WHERE key = 'pDate'), 
			'-1 month'))`
	_, err := db.Exec(query, pname, energy)
	return err
}

///////////////////////////////////////////////////////////////////////////

func (db *DB) LimitInsert(substation, energy int, date Date) error {
	query := `
		INSERT INTO limits (substation, energy, date)
		VALUES (?, ?, ?)
		`
	_, err := db.Exec(query, substation, energy, date.timestring())
	return err
}

func (db *DB) LimitUpdate(substation, energy int, date Date) error {
	query := `
		UPDATE limits SET energy = ?
		WHERE substation = ? AND date = ?
		`
	_, err := db.Exec(query, energy, substation, date.timestring())
	return err
}

///////////////////////////////////////////////////////////////////////////

func (db *DB) StatUpdate(key, value string) error {
	query := `UPDATE stat SET value = ? WHERE key = ?`
	_, err := db.Exec(query, value, key)
	return err
}
