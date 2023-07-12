MY SQL


docker run --name=mysql1 --restart on-failure  -e MYSQL_ROOT_HOST=%  -d  -p3306:3306 container-registry.oracle.com/mysql/community-server:latest


 -e MYSQL_ROOT_HOST=%   ==> allow root user to login fromany where

docker logs mysql1 2>&1 | grep GENERATED
[Entrypoint] GENERATED ROOT PASSWORD: FUN:ZR7_P32X1h3G_#WAP=7:Cr2U?D.9


Connecting to MySQL Server from within the Container
docker exec -it mysql1 mysql -uroot -p


>> have to reset the password

mysql> ALTER USER 'root'@'localhost' IDENTIFIED BY 'SaveP0wer#2';



mysql> CREATE USER 'sumit'@'172.17.0.1' IDENTIFIED BY 'SaveP0wer#2';
mysql> GRANT ALL PRIVILEGES ON *.* TO 'sumit'@'172.17.0.1' WITH GRANT OPTION;
mysql> GRANT ALL PRIVILEGES ON *.* TO 'sumit'@'172.17.0.1' WITH GRANT OPTION;


show databases;

mysql> create database testdb;

use testdb;
create table testtb1 (name char(20), email char(200));
insert into testtb1 (name,email) values('a','a@example.com');
insert into testtb1 (name,email) values('b','b@example.com');
insert into testtb1 (name) values('b');
insert into testtb1 (email) values('x@example');


=== catalog view

select * from information_schema.tables where table_name='testtb1';

===================================================================

DELIMITER //

CREATE PROCEDURE GetAllProducts()
BEGIN
	SELECT *  FROM testtb1;
END // 
DELIMITER ;


call GetAllProducts()

============================MULTI RS======================================

DELIMITER //

CREATE PROCEDURE MULTIRS()
BEGIN
	SELECT *  FROM testtb1;
	SELECT *  FROM testtb1 where name='a';
END // 
DELIMITER ;


call GetAllProducts()

================================ OUT parms===========================


DELIMITER //
CREATE PROCEDURE spoutparm(
    IN input1 int,
    OUT output1 int
     )
BEGIN
     set output1 = 10;
END //
DELIMITER ;




DELIMITER //
CREATE PROCEDURE spinoutparm(
    IN input1 int,
	inout inout1 int,
    OUT output1 int
     )
BEGIN
     
     set output1 = inout1+10;
	 set inout1 = input1;
END //
DELIMITER ;

set @inoa = 99

call spinoutparm(1,@inoa, @ob);

select @ob, @inoa;





=====================================================================


https://dev.mysql.com/doc/refman/8.0/en/information-schema-routines-table.html
mysql> select * from information_schema.routines where SPECIFIC_NAME='GetAllProducts';


ROUTINE_TYPE ==> PROCEDURE   PROCEDURE for stored procedures, FUNCTION for stored functions.



SPECIFIC_NAME  | ROUTINE_CATALOG | ROUTINE_SCHEMA | ROUTINE_NAME   | ROUTINE_TYPE | DATA_TYPE | CHARACTER_MAXIMUM_LENGTH | CHARACTER_OCTET_LENGTH | NUMERIC_PRECISION | NUMERIC_SCALE | DATETIME_PRECISION | CHARACTER_SET_NAME | COLLATION_NAME | DTD_IDENTIFIER | ROUTINE_BODY | ROUTINE_DEFINITION                | EXTERNAL_NAME | EXTERNAL_LANGUAGE | PARAMETER_STYLE | IS_DETERMINISTIC | SQL_DATA_ACCESS | SQL_PATH | SECURITY_TYPE | CREATED             | LAST_ALTERED        | SQL_MODE                                                                                                              | ROUTINE_COMMENT | DEFINER        | CHARACTER_SET_CLIENT | COLLATION_CONNECTION | DATABASE_COLLATION 




 
catalog vs schema

Schemas are supported and interpreted as MySQL database names, specifying catalog triggers an error. Both catalogs and schemas are supported but it is an error if both are specified at the same time. If only catalog or only schema is specified, it is interpreted as a MySQL database name.

ROUTINE_CATALOG  ==> The name of the catalog to which the routine belongs. This value is always def.




Paramters 
https://dev.mysql.com/doc/refman/8.0/en/information-schema-parameters-table.html