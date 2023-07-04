BEGIN TRANSACTION;
--
INSERT INTO places VALUES(1,208,'1234567890abcdef','Госпдвір');
INSERT INTO places VALUES(2,220,NULL,'АВМ');
INSERT INTO places VALUES(3,205,NULL,'Контора');
--
INSERT INTO meters VALUES(1,1,1,'НІК2301АП1',2020,'344848',4,40);
INSERT INTO meters VALUES(2,2,0,'НІК2102-02',2021,'475434',4,40);
INSERT INTO meters VALUES(3,3,1,NULL,NULL,'001930',5,1);
INSERT INTO meters VALUES(4,2,1,'НІК2102-02',2022,'E12345',4,40);
--
INSERT INTO readings VALUES('2021-10-01',1,1,9348,NULL);
INSERT INTO readings VALUES('2021-10-01',2,1,3371,NULL);
INSERT INTO readings VALUES('2021-10-01',3,1,10736,NULL);
INSERT INTO readings VALUES('2021-10-01',3,2,10000,NULL);
--
INSERT INTO readings VALUES('2021-11-01',1,1,9525,NULL);
INSERT INTO readings VALUES('2021-11-01',2,1,3400,NULL);
INSERT INTO readings VALUES('2021-11-01',4,1,7400,NULL);
INSERT INTO readings VALUES('2021-11-01',3,1,11577,NULL);
INSERT INTO readings VALUES('2021-11-01',3,2,11500,NULL);
--
INSERT INTO readings VALUES('2021-12-01',1,1,9721,NULL);
INSERT INTO readings VALUES('2021-12-01',2,1,3426,NULL);
INSERT INTO readings VALUES('2021-12-01',4,1,7426,NULL);
INSERT INTO readings VALUES('2021-12-01',3,1,12575,NULL);
INSERT INTO readings VALUES('2021-12-01',3,2,12500,NULL);
--
INSERT INTO readings VALUES('2022-01-01',1,1,9907,NULL);
INSERT INTO readings VALUES('2022-01-01',4,1,7455,NULL);
INSERT INTO readings VALUES('2022-01-01',3,1,13350,NULL);
INSERT INTO readings VALUES('2022-01-01',3,2,13300,NULL);
--
INSERT INTO readings VALUES('2022-02-01',1,1,0064,NULL);
INSERT INTO readings VALUES('2022-02-01',4,1,7481,NULL);
INSERT INTO readings VALUES('2022-02-01',3,1,13745,NULL);
INSERT INTO readings VALUES('2022-02-01',3,2,13700,NULL);
--
INSERT INTO readings VALUES('2022-03-01',4,1,7581,NULL);
UPDATE meters SET active = false WHERE meter_id = 4;
--
UPDATE service SET value = '2022-03-01' WHERE skey = 'next_date';
--
COMMIT;