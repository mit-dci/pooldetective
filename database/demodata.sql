INSERT INTO algorithms(name) VALUES ('SHA256');
INSERT INTO algorithms(name) VALUES ('Scrypt');

INSERT INTO coins(ticker, name) VALUES ('BTC', 'Bitcoin');
INSERT INTO coins(ticker, name) VALUES ('LTC', 'Litecoin');

INSERT INTO coin_algorithm(coin_id, algorithm_id) VALUES (1,1);
INSERT INTO coin_algorithm(coin_id, algorithm_id) VALUES (2,2);

INSERT INTO locations(name) VALUES ('Default location');

INSERT INTO pools(name) VALUES ('BSOD.PW');

INSERT INTO pool_observers(pool_id, coin_id, algorithm_id, location_id,  stratum_host, stratum_port, stratum_username, stratum_password)
VALUES (1,1,1,1,'pool.bsod.pw',3333,'16TSsrrQW883buCUpsSTDz81gcEkTo4Tkr','x');

INSERT INTO pool_observers(pool_id, coin_id, algorithm_id, location_id,  stratum_host, stratum_port, stratum_username, stratum_password)
VALUES (1,2,2,1, 'pool.bsod.pw',2155,'MLWpQqBn53sg4RbaM6KfjU9pKiaNXdVk5A','x');

INSERT INTO stratum_servers(location_id, algorithm_id, port)
VALUES (1,1,42651);
INSERT INTO stratum_servers(location_id, algorithm_id, port)
VALUES (1,2,42652);
