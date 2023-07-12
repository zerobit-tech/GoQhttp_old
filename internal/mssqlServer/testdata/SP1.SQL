CREATE PROCEDURE testsp01 (@p_InputInt  INT, @p_OutputInt INT OUTPUT , @P_CHAR CHAR(20) OUT , @P_DATE DATE OUT) AS BEGIN
SELECT * FROM testtb1 ;
SELECT @p_OutputInt = @p_OutputInt + @p_InputInt;
END


EXEC testsp01 @p_InputInt='2' ,@p_OutputInt=3, @P_CHAR=NULL, @P_DATE=NULL