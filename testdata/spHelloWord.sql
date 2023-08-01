CREATE OR REPLACE PROCEDURE SUMITG1.HelloWorld (
                IN Name CHAR(10) DEFAULT '',
              
         
                OUT Message CLOB
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.HelloWorld
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        DYNAMIC RESULT SETS 0
        CALLED ON NULL INPUT
PROCBODY : BEGIN
 
SET Message = 'Hello ' || Name || ' ' || 'Welcome to QHTTP';
 

 
RETURN ;
END PROCBODY 