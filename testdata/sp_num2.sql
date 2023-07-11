CREATE OR REPLACE PROCEDURE SUMITG1.SPNUM2 (
              
              IN NUMFIELD num DEFAULT 17 ,

             
              INOUT ioNUMFIELD num DEFAULT 24 ,

              OUT oNUMERICFIELD numeric  ,
              OUT oNUMFIELD num   
 
                
            )
         
        LANGUAGE SQL
        SPECIFIC SUMITG1.SPNUM2
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        CALLED ON NULL INPUT

PROCBODY : BEGIN
 
SET ONUMFIELD = 10 ;

 
SET IONUMFIELD = 20 ;
set oNUMERICFIELD = 30;

RETURN ;
END PROCBODY 