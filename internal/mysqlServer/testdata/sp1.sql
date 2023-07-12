CREATE PROCEDURE spinoutparm3(
    IN input1 int,
	inout inout1 char,
    OUT output1 DATE
     )
BEGIN
    SELECT *  FROM mysqlMain;
	SELECT *  FROM mysqlMain where Column5='a';
     set output1 = NULL;
	 set inout1 = NULL;
END  