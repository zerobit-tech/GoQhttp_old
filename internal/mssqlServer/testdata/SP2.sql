CREATE PROCEDURE testsp02 (@p_InputInt  INT, @p_OutputInt INT OUTPUT , @P_CHAR CHAR(20) OUT , @P_DATE DATE OUT) AS BEGIN
SELECT * FROM testtb1 ;
SELECT @p_OutputInt = @p_OutputInt + @p_InputInt;
SELECT @P_DATE = NULL;
END
