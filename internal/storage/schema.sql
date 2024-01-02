-- EnergoZvit
-- Sqlite database schema
--
PRAGMA user_version = 1;
PRAGMA foreign_keys = ON;
--
-------------------------------- TABLES --------------------------------
--
-- Таблиця точок обліку, де встановлено лічильник
CREATE TABLE IF NOT EXISTS places (
    place_id   -- Первинний ключ
               INTEGER PRIMARY KEY ASC NOT NULL,      
    substation -- Номер підстанції
               INTEGER,
    eic        -- Energy Identification Code
               CHAR(16)                      
               CONSTRAINT eic_not_valid
               CHECK(length(eic) == 16),
    name       -- Назва точки обліку
               VARCHAR(24) UNIQUE NOT NULL   
               CONSTRAINT name_empty
               CHECK(length(trim(name)) != 0)
               CONSTRAINT name_too_long
               CHECK(length(name) <= 24)
);
--
-- Таблиця лічильників
CREATE TABLE IF NOT EXISTS meters (
    meter_id   -- Первинний ключ
               INTEGER PRIMARY KEY ASC NOT NULL,
    place_id   -- Посилання на площадку вимірювання
               INTEGER NOT NULL
               REFERENCES places
               ON DELETE RESTRICT 
               ON UPDATE RESTRICT,
    active     -- Чи діючий лічильник
               BOOLEAN DEFAULT true NOT NULL
               CONSTRAINT active_not_valid
               CHECK(active IN (false, true)),
    model      -- Модель лічильника
               VARCHAR(24)
               CONSTRAINT model_too_long
               CHECK(length(model) <= 24),
    year       -- Рік виготовлення лічильника
               INTEGER
               CONSTRAINT year_not_valid
               CHECK(year BETWEEN 1000 AND 9999),
    serial     -- Серійний номер лічильника
               VARCHAR(24) NOT NULL
               CONSTRAINT serial_too_long
               CHECK(length(serial) <= 24),
    digits     -- Кількість значущих розрядів (див. reports.energy)
               INTEGER NOT NULL
               CONSTRAINT digits_not_valid
               CHECK(digits BETWEEN 1 AND 8),
    ratio      -- Коефіцієнт трансформації
               INTEGER NOT NULL
               CONSTRAINT ratio_not_valid
               CHECK(ratio > 0)
);
--
-- Показники лічильників
CREATE TABLE IF NOT EXISTS readings (
    rdate      -- Дата в форматі РРРР-ММ-ДД, день завжди 01
               CHAR(10) NOT NULL
               CONSTRAINT wrong_date_format 
               CHECK(date(rdate) NOT NULL)
               CONSTRAINT wrong_day_in_date
               CHECK(rdate == date(rdate, 'start of month')),
    meter_id   -- Посилання на лічильник
               INTEGER NOT NULL
               REFERENCES meters
               ON DELETE RESTRICT 
               ON UPDATE RESTRICT,
    zone       -- Номер тарифної зони
               INTEGER NOT NULL
               CONSTRAINT zone_not_valid
               CHECK(zone BETWEEN 1 AND 3),
    kwh        -- Показники лічильника
               INTEGER NOT NULL
               CONSTRAINT kwh_not_valid
               CHECK(kwh >= 0),
    annotation -- Примітка
               VARCHAR(32)
               CONSTRAINT annotation_too_long
               CHECK(length(annotation) <= 32),
    -- Унікальний ключ рядка
    PRIMARY KEY (rdate, meter_id, zone)
);
--
-- Сервісна таблиця для внутрішнього використання
CREATE TABLE IF NOT EXISTS service ( 
    skey       -- Ключ
               VARCHAR(16) PRIMARY KEY NOT NULL,
    value      -- Значення
               VARCHAR(16) NOT NULL
);
--
-- Запис початкових даних
INSERT OR IGNORE INTO service VALUES
    ('goto_next_date', '0'),
    ('next_date', date('now', 'start of month'));
--
--
CREATE TRIGGER IF NOT EXISTS goto_next_date_update
AFTER UPDATE ON service
WHEN NEW.skey = 'goto_next_date'
BEGIN
    -- Перевірка чи дані всіх лічильників введені
    VALUES(
    CASE
        WHEN (SELECT count(*)
                FROM next_reports
               WHERE cur_kwh IS NULL) > 0
        THEN RAISE(ABORT, 'missing_readings')
    END);
    -- Оновлення дати
    UPDATE service
       SET value = date(
           (SELECT value
              FROM service
             WHERE skey = 'next_date'),
              'start of month', '+1 months')
     WHERE skey = 'next_date';
END;
--
-------------------------------- VIEWS ---------------------------------
--
-- Представлення звітів.
CREATE VIEW IF NOT EXISTS reports AS
SELECT cur.rdate      AS rdate,     -- Дата
       cur.meter_id  AS meter_id,   -- ID лічильника
       substation,                  -- Номер підстанції
       eic,                         -- EIC код
       name,                        -- Назва площадки вимірювання
       model,                       -- Модель лічильника
       year,                        -- Рік виготовлення лічильника
       serial,                      -- Серійний номер лічильника
       digits,                      -- Кількість значущих розрядів
       ratio,                       -- Коефіцієнт трансформації
       cur.zone       AS zone,      -- Номер тарифної зони
       cur.kwh        AS cur_kwh,   -- Поточні показники лічильника
       pre.kwh        AS pre_kwh,   -- Попередні показники лічильника
       mod(cur.kwh - pre.kwh + power(10, digits), power(10, digits))
                      AS diff,      -- Різниця показників
       mod(cur.kwh - pre.kwh + power(10, digits), power(10, digits)) * ratio
                      AS energy,    -- Спожита електроенергія
       cur.annotation AS annotation -- Примітка
  FROM readings AS pre, readings AS cur
  JOIN meters USING(meter_id)
  JOIN places  USING(place_id)
 WHERE cur.meter_id = pre.meter_id 
   AND cur.zone = pre.zone
   AND pre.kwh NOT NULL
   AND pre.rdate = date(cur.rdate, '-1 month');
--
--
-- Форма для вводу показників. Ввід показників вводити в такій
-- послідовності:
-- 1. Читається ця таблиця;
-- 2. Оновлюється (update) cur_kwh для кожного запису.
--    Ключами є meter_id та zone;
-- 3. Оновлюється (update) ключ goto_next_date в таблиці service.
--    Повертається помилка якщо введені не всі показники. В іншому
--    випадку збільшується next_date в таблиці service.
CREATE VIEW IF NOT EXISTS next_reports AS
SELECT pre.meter_id  AS meter_id,   -- ID лічильника
       substation,                  -- Номер підстанції
       eic,                         -- EIC код
       name,                        -- Назва площадки вимірювання
       model,                       -- Модель лічильника
       year,                        -- Рік виготовлення лічильника
       serial,                      -- Серійний номер лічильника
       digits,                      -- Кількість значущих розрядів
       ratio,                       -- Коефіцієнт трансформації
       pre.zone       AS zone,      -- Номер тарифної зони
       cur.kwh        AS cur_kwh,   -- Теперішні показники лічильника
       pre.kwh        AS pre_kwh,   -- Попередні показники лічильника
       mod(cur.kwh - pre.kwh + power(10, digits), power(10, digits))
                      AS diff,      -- Різниця показників
       mod(cur.kwh - pre.kwh + power(10, digits), power(10, digits)) * ratio
                      AS energy,    -- Спожита електроенергія
       cur.annotation AS annotation -- Примітка
  FROM readings AS pre
  LEFT JOIN readings AS cur
    ON cur.rdate = date(pre.rdate, '+1 month')
   AND cur.meter_id = pre.meter_id
   AND cur.zone = pre.zone
  JOIN meters USING(meter_id)
  JOIN places  USING(place_id)
 WHERE meters.active = true
   AND pre.rdate = date(
       (SELECT value 
          FROM service
         WHERE skey = 'next_date'),
       '-1 month');
--
--
--
CREATE TRIGGER IF NOT EXISTS next_reports_update
INSTEAD OF UPDATE ON next_reports
FOR EACH ROW
BEGIN
    INSERT OR REPLACE INTO readings (
        rdate, meter_id, zone, kwh, annotation)
    VALUES (
        date((SELECT value FROM service WHERE skey = 'next_date'),
            'start of month'),
        NEW.meter_id,
        NEW.zone,
        NEW.cur_kwh,
        NEW.annotation
    );
END;
