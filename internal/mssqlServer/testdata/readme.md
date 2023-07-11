MS SQL SERVER

https://github.com/microsoft/go-mssqldb


deb [arch=amd64,armhf,arm64] https://packages.microsoft.com/ubuntu/22.04/mssql-server-2022 jammy main


https://learn.microsoft.com/en-us/sql/linux/quickstart-install-connect-docker?view=sql-server-ver15&pivots=cs1-bash

sudo docker run -e "ACCEPT_EULA=Y" -e "MSSQL_SA_PASSWORD=SaveP0wer#2" \
   -p 1433:1433 --name sql1 --hostname sql1 \
   -d \
   mcr.microsoft.com/mssql/server:2019-latest
   
   
   >> localhost:1433
   >> user SA
   >> password SaveP0wer#2
   
   
   
  docker exec -it sql1 /opt/mssql-tools/bin/sqlcmd -S localhost -U SA -P "SaveP0wer#2"
   
   ----------------
   use TestDB
   GO
   ------------
   
   1> SELECT Name from sys.databases;

GO




================ tables =======================

create table testtb1 (id int, c1 char(20), c2 varchar(20), c3 nchar(20), c4 nvarchar(20), d1 date , d2 DATETIME2 , d3 DATETIME, d4 DATETIMEOFFSET, d5  SMALLDATETIME, t1 time , i1 int  ,i2 bigint, i3 smallint, i4 tinyint, c5 text, f1 decimal, f2 numeric, f3 float , f4 real,  f5 money, f7 smallmoney);








   -------------------
   CREATE PROCEDURE SelectAllCustomers @name nvarchar(30) AS SELECT * FROM Inventory WHERE name= @name
GO;

1> EXEC SelectAllCustomers @name='banana';

-----------



USE TestDB

CREATE PROCEDURE SimpleInOutProcedure (@p_InputInt  INT, @p_OutputInt INT OUTPUT) AS BEGIN
SELECT * FROM Inventory ;
SELECT @p_OutputInt = @p_OutputInt + @p_InputInt;
END

==
1> DECLARE @p_OutputInt int = 4
2> EXEC SimpleInOutProcedure @p_InputInt = 1, @p_OutputInt = @p_OutputInt OUTPUT
3> SELECT @p_OutputInt
4> GO







SimpleInOutProcedure23 >> multiple result sets




===
catalog views::: 

select * from sys.objects where [type]='P';


>> sp parameters
SELECT SCHEMA_NAME(schema_id) AS schema_name  
    ,o.name AS object_name  
    ,o.type_desc  
    ,p.parameter_id  
    ,p.name AS parameter_name  
    ,TYPE_NAME(p.user_type_id) AS parameter_type  
    ,p.max_length  
    ,p.precision  
    ,p.scale  
    ,p.is_output  
FROM sys.objects AS o  
INNER JOIN sys.parameters AS p ON o.object_id = p.object_id  
WHERE o.object_id = OBJECT_ID('SimpleInOutProcedure')  
ORDER BY schema_name, object_name, p.parameter_id;  


>>>> qualiedied 
SELECT *
FROM sys.objects AS o  
INNER JOIN sys.parameters AS p ON o.object_id = p.object_id  
WHERE o.object_id = OBJECT_ID('chains.p001')  
ORDER BY  p.parameter_id;  


>>> fully qualified name

SELECT * FROM testdb.sys.objects AS o   INNER JOIN testdb.sys.parameters AS p ON o.object_id = p.object_id   WHERE o.object_id = OBJECT_ID('testdb.chains.p001')   ORDER BY  p.parameter_id;     






>> sp info
 SELECT * FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_TYPE = N'PROCEDURE' and ROUTINE_SCHEMA = N'dbo' 



================== user =======================
-- Creates the login AbolrousHazem with password '340$Uuxwp7Mcxo7Khy'.  
CREATE LOGIN sumit with password = 'A2Password#2';
Go

-- Creates a database user for the login created above.  
CREATE USER sumit FOR LOGIN sumit;  


give user permission to create view

1> GRANT CREATE VIEW TO [sumit];
2> GRANT CREATE PROCEDURE to [sumit];
3> GRANT ALTER ON SCHEMA::[dbo] to [sumit];


   CREATE PROCEDURE p001 @name nvarchar(30) AS SELECT * FROM Inventory WHERE name= @name



==================== SCHEMA ===================

one db can have multiple schema

default is dbo

CREATE SCHEMA Chains;
