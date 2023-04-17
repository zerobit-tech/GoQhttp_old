CREATE OR REPLACE table SUMITG1.SQLDTTABLE 
 (                                             
    TIMEFIELD TIME    ,                           
    TIMSTAMPFIELD TIMESTAMP,                   
    DATEFIELD DATE,                            
                                               
    CHARFIELD CHAR(10),                        
      
    VARCHARFIELD VARCHAR(20),                  
                                               
                                               
    CLOBFIELD CLOB,                            
                                               
    GRAPHICFIELD        GRAPHIC(10)         ,  
    VARGRAPHICFIELD     VARGRAPHIC(20)      ,  
DBCLOBFIELD         DBCLOB              ,  
                                           
                                           
SMALLINTFIELD       SMALLINT            ,  
INTERGERFIELD       Integer            ,  
INTFIELD            INT                 ,  
BININTFIELD         BIGINT              ,  
                                           
                                           
DECIMALFIELD        DECIMAL(9,2)       ,   
DECFIELD            DEC(9)             ,   
NUMERICFIELD        NUMERIC(11,2)      ,   
NUMFIELD            NUM(11)            ,   
BINARYFIELD         BINARY(100)         ,   
VARBINARYFIELD       VARBINARY(100)    ,    
BLOBFIELD            BLOB               ,   
XMLFIELD             XML                ,   
                                            
                                            
DECFLOATFIELD        DECFLOAT           ,   
                                            
                                            
FLOATFIELD           FLOAT              ,   
REALFIELD            REAL               ,   
DOUBLEFIELD          DOUBLE             ,   
                                            
                                            
                                           
ROWIDFIELD           ROWID              ,   
DATALINKFIELD        DATALINK               );