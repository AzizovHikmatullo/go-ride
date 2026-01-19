ALTER TABLE rides DROP CONSTRAINT rides_user_id_fkey;

ALTER TABLE rides DROP CONSTRAINT rides_driver_id_fkey;

DROP TABLE users;

DROP TABLE rides;