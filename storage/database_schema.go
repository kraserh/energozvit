package storage

// Схема бази даних.
const dbSchema = `
BEGIN TRANSACTION;

CREATE TABLE location ( 
	id         INTEGER PRIMARY KEY,
	substation INTEGER NOT NULL,
	lname      TEXT    NOT NULL CHECK(lname != ""),
	sdate      DATE    NOT NULL CHECK(sdate LIKE '____-__-__'),
	edate      DATE             CHECK(edate LIKE '____-__-__')
);


CREATE TABLE meters (
	id         INTEGER PRIMARY KEY,
	idlocation REFERENCES location(id),
	model      TEXT    NOT NULL CHECK(model != ""),
	number     TEXT    NOT NULL CHECK(number != ""),
	limval     INTEGER NOT NULL CHECK(limval >= 10),
	ratio      INTEGER NOT NULL CHECK(ratio > 0), 
	numzones   INTEGER NOT NULL CHECK(numzones > 0), 
	sdate      DATE    NOT NULL CHECK(sdate LIKE '____-__-__'),
	edate      DATE             CHECK(edate LIKE '____-__-__')
);


CREATE TABLE mlog (
	idmeter    REFERENCES meters(id),
	zone       INTEGER NOT NULL CHECK(zone > 0),
	value      INTEGER NOT NULL CHECK(value >= 0),
	init       BOOL    NOT NULL DEFAULT 0,
	comment    TEXT    NOT NULL DEFAULT "",
	date       DATE    NOT NULL CHECK(date LIKE '____-__-__'),
	PRIMARY KEY (idmeter, zone, date)
);

CREATE TABLE parts (
	id         INTEGER PRIMARY KEY,
	idlocation REFERENCES location(id),
	pname      TEXT    NOT NULL CHECK(pname != ""),
	sdate      DATE    NOT NULL CHECK(sdate LIKE '____-__-__'),
	edate      DATE             CHECK(edate LIKE '____-__-__')
);

CREATE TABLE plog (
	idpart     REFERENCES parts(id),
	energy     INTEGER NOT NULL CHECK(energy >= 0),
	date       TEXT    NOT NULL CHECK(date LIKE '____-__-__'),
	PRIMARY KEY (idpart, date)
);


CREATE TABLE limits (
	substation INTEGER NOT NULL CHECK(substation != ""),
	energy     INTEGER NOT NULL CHECK(energy >= 0),
	date       DATE    NOT NULL CHECK(date LIKE '____-__-__'),
	PRIMARY KEY (substation, date)
);


CREATE TABLE stat ( 
	key     TEXT    PRIMARY KEY,
	value   TEXT    NOT NULL
);


INSERT INTO stat VALUES
	('version', '0.2'),
	('mDate', '2000-01-01'),
	('pDate', '2000-01-01');


---------------------------------------------------------------------

CREATE VIEW mreport AS
	WITH curlog AS (
		SELECT idmeter, zone, value, comment, date 
			FROM mlog 
			WHERE init = 0
	)
	SELECT substation, lname, model, number, ratio, curlog.zone AS zone,
		curlog.value AS cur, prevlog.value AS prev, 
		CASE WHEN curlog.value >= prevlog.value 
			THEN curlog.value - prevlog.value
			ELSE curlog.value - prevlog.value + limval
		END AS diff,
		CASE WHEN curlog.value >= prevlog.value 
			THEN (curlog.value - prevlog.value) * ratio
			ELSE (curlog.value - prevlog.value + limval) * ratio
		END AS energy,
		curlog.comment AS comment, curlog.date AS date
	FROM curlog, mlog AS prevlog 
	JOIN meters  ON curlog.idmeter = meters.id
	JOIN location ON idlocation = location.id
	WHERE curlog.idmeter = prevlog.idmeter 
		AND curlog.zone = prevlog.zone 
		AND curlog.date = date(prevlog.date, '+1 month');


CREATE VIEW form_mreport AS
	SELECT substation, lname, model, number, limval, ratio, zone, value as prev
	FROM mlog
	JOIN meters ON idmeter = meters.id
	JOIN location ON idlocation = location.id
	WHERE date = date(
		(SELECT value FROM stat WHERE key = 'mDate'), '-1 month')
		AND meters.edate IS NULL;


CREATE VIEW preport AS
	SELECT substation, lname, pname, energy, date
	FROM plog
	JOIN parts ON idpart = parts.id
	JOIN location ON idlocation = location.id;


CREATE VIEW mpower AS
	SELECT substation, lname, sum(energy) AS energy, date
	FROM mreport
	GROUP BY lname, date;


CREATE VIEW ppower AS
	SELECT lname, pname, energy, date 
	FROM preport
	JOIN parts ON idpart = parts.id
	JOIN location ON idlocation = location.id;

---------------------------------------------------------------------

CREATE TRIGGER check_lname
	BEFORE INSERT ON location
	WHEN 
	EXISTS (SELECT lname FROM location WHERE edate IS NULL AND lname = NEW.lname)
	OR EXISTS (SELECT pname FROM parts WHERE edate IS NULL AND pname = NEW.lname)
	BEGIN
		SELECT RAISE(ROLLBACK, 'location name is exists');
	END;


CREATE TRIGGER check_pname
	BEFORE INSERT ON parts
	WHEN 
	EXISTS (SELECT lname FROM location WHERE edate IS NULL AND lname = NEW.pname)
	OR EXISTS (SELECT pname FROM parts WHERE edate IS NULL AND pname = NEW.pname)
	BEGIN
		SELECT RAISE(ROLLBACK, 'parts name is exists');
	END;


CREATE TRIGGER check_pDate
	BEFORE UPDATE OF value ON stat
	WHEN NEW.key = 'pDate'
		AND NEW.value > (SELECT value FROM stat WHERE key = 'mDate')
	BEGIN
		SELECT RAISE(IGNORE);
	END;

CREATE TRIGGER check_limval_insert
	BEFORE INSERT ON mlog
	WHEN NEW.value >= (SELECT limval FROM meters WHERE id = NEW.idmeter)
	BEGIN
		SELECT RAISE(ROLLBACK, 'limval out range');
	END;


CREATE TRIGGER check_limval_update
	BEFORE UPDATE OF value ON mlog
	WHEN NEW.value >= (SELECT limval FROM meters, mlog WHERE id = idmeter)
	BEGIN
		SELECT RAISE(ROLLBACK, 'limval out range');
	END;


CREATE TRIGGER delete_location
	AFTER UPDATE OF edate ON location
	BEGIN
		UPDATE meter
			SET edate = NEW.edate
			WHERE edate IS NULL AND idlocation = OLD.id;
		UPDATE parts
			SET edate = NEW.edate
			WHERE edate IS NULL AND idlocation = OLD.id;
		DELETE
			FROM location
			WHERE id = OLD.id AND NEW.edate < OLD.sdate;
	END;

CREATE TRIGGER delete_meter
	AFTER UPDATE OF edate ON meters
	BEGIN
		DELETE
			FROM meter
			WHERE id = OLD.id AND NEW.edate < OLD.sdate;
	END;

CREATE TRIGGER delete_parts
	AFTER UPDATE OF edate ON parts
	BEGIN
		DELETE
			FROM parts
			WHERE id = OLD.id AND NEW.edate < OLD.sdate;
	END;

COMMIT;`
