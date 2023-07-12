CREATE OR REPLACE PROCEDURE SUMITG1.SPPARM (
                IN inputval CHAR(20) DEFAULT '',
               
                OUT QHTTP_STATUS_CODE int
               
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.SPPARM
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        
        CALLED ON NULL INPUT
PROCBODY : BEGIN
  set QHTTP_STATUS_CODE = 200;

   if inputval='a' 
   then set QHTTP_STATUS_CODE = 201 ;
   end if;
   
    if inputval='b' 
   then set QHTTP_STATUS_CODE = 301 ;
   end if;
   
   
      if inputval='c' 
   then set QHTTP_STATUS_CODE = 404 ;
   end if;
   
   
        if inputval='d' 
   then set QHTTP_STATUS_CODE = 500 ;
   end if;
   
   
RETURN ;
END PROCBODY 

call SUMITG1.SPPARM (inputval=>'b',QHTTP_STATUS_CODE=>?)