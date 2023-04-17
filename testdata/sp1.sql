CREATE OR REPLACE procedure SUMITG1.DTPROC1 
 (                                             
   in TIMEFIELD TIME   default current time  ,                           
   in TIMSTAMPFIELD TIMESTAMP default CURRENT TIMESTAMP ,                   
   in DATEFIELD DATE default  CURRENT Date,                            
                                               
   in CHARFIELD CHAR(10) default '',                        
      
   in VARCHARFIELD VARCHAR(20)    default '',                  
                                               
                                               
   in CLOBFIELD CLOB  default '',                            
                                               
   in GRAPHICFIELD        GRAPHIC(10)       default  ''   ,  
   in VARGRAPHICFIELD     VARGRAPHIC(20)      default  ''  ,  
in DBCLOBFIELD         DBCLOB              default  '' ,  
                                           
                                           
in SMALLINTFIELD       SMALLINT      default     1   ,  
in INTERGERFIELD       Integer       default     2   ,  
in INTFIELD            INT         default       3   ,  
in BININTFIELD         BIGINT       default      4    ,  
                                           
                                           
in DECIMALFIELD        DECIMAL(9,2)     default   12.15  ,   
in DECFIELD            DEC(9)           default   13  ,   
in NUMERICFIELD        NUMERIC(11,2)    default   14.16  ,   
in NUMFIELD            NUM(11)             default  16,   

in BINARYFIELD         BINARY(100)     default      'a' ,   
in VARBINARYFIELD       VARBINARY(100)    default   'b',    
in BLOBFIELD            BLOB              default   'c' ,   
in XMLFIELD             XML             default null ,   
                                            
                                            
in DECFLOATFIELD        DECFLOAT         default   10  ,   
                                            
                                            
in FLOATFIELD           FLOAT            default   10.12  ,   
in REALFIELD            REAL             default   11.13 ,   
in DOUBLEFIELD          DOUBLE            default   12.14 ,   
                                            
                                            
                                           
in ROWIDFIELD           ROWID               ,   
in DATALINKFIELD        DATALINK              default  null )
DYNAMIC RESULT SETS 2      
LANGUAGE SQL               
SPECIFIC SUMITG1.DTPROC1  
NOT DETERMINISTIC          
MODIFIES SQL DATA          
CALLED ON NULL INPUT       
 PROCBODY : BEGIN          
 DECLARE C1 CURSOR WITH RETURN FOR                    
    SELECT * FROM SUMITG1.SQLDTTABLE ;
                                                     
DECLARE C2 CURSOR WITH RETURN  FOR                   
    SELECT * FROM SUMITG1.SQLDTTABLE2 ;          
    
     OPEN C1;  
      OPEN C2;
      RETURN; 
  END PROCBODY