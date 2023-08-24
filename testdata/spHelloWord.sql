CREATE OR REPLACE PROCEDURE SUMITG1.HelloWorld (
                IN Name CHAR(10) DEFAULT '',
            
         
                OUT Message char(200)
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.HelloWorld
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        DYNAMIC RESULT SETS 0
        CALLED ON NULL INPUT
PROCBODY : BEGIN
 
SET Message = 'Hello ' || trim(Name) || '.' || 'Welcome to QHTTP.';
 

 
RETURN ;
END PROCBODY 