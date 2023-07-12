 
-- testdb.mysqlMain definition

CREATE TABLE testdb.mysqlMain (
	i1 BIGINT NULL,
	i2 BIGINT UNSIGNED NULL,
	Column1 BINARY NULL,
	Column2 BIT NULL,
	Column3 BLOB NULL,
	Column4 BOOL NULL,
	Column5 CHAR NULL,
	Column6 DATE NULL,
	Column7 DATETIME NULL,
	Column8 DECIMAL NULL,
	Column9 DOUBLE NULL,
	Column10 DOUBLE PRECISION NULL,
	Column11 ENUM ('x-small', 'small', 'medium', 'large', 'x-large'),
	Column12 FLOAT NULL,
	Column13 INT NULL,
	Column14 INT UNSIGNED NULL,
	Column15 LONG VARBINARY NULL,
	Column16 LONG VARCHAR NULL,
	Column17 LONGBLOB NULL,
	Column18 LONGTEXT NULL,
	Column19 NUMERIC NULL,
	Column20 REAL NULL,
	Column21 SET('one', 'two') NULL,
	Column22 TIME NULL,
	Column23 TIMESTAMP NULL,
	Column24 VARCHAR(100) NULL,
	Column25 VARBINARY(100) NULL,
	Column26 YEAR NULL,
	Column27 json NULL
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;