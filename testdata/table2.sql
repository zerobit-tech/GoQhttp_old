CREATE OR REPLACE table SUMITG1.SQLDTTABLE2 
 (                                             
    TIMEFIELD TIME  not null with default  ,                           
    TIMSTAMPFIELD TIMESTAMP not null with default ,                   
    DATEFIELD DATE not null with default ,                            
                                               
    CHARFIELD CHAR(10) not null with default ,                        
      
    VARCHARFIELD VARCHAR(20)  not null with default ,                  
                                               
                                               
    CLOBFIELD CLOB not null with default ,                            
                                               
    GRAPHICFIELD        GRAPHIC(10)      not null with default     ,  
    VARGRAPHICFIELD     VARGRAPHIC(20)    not null with default    ,  
DBCLOBFIELD         DBCLOB             not null with default   ,  
                                           
                                           
SMALLINTFIELD       SMALLINT      not null with default        ,  
INTERGERFIELD       Integer      not null with default        ,  
INTFIELD            INT         not null with default          ,  
BININTFIELD         BIGINT      not null with default          ,  
                                           
                                           
DECIMALFIELD        DECIMAL(9,2)    not null with default     ,   
DECFIELD            DEC(9)          not null with default     ,   
NUMERICFIELD        NUMERIC(11,2)   not null with default     ,   
NUMFIELD            NUM(11)            not null with default  ,   
BINARYFIELD         BINARY(100)    not null with default       ,   
VARBINARYFIELD       VARBINARY(100)   not null with default   ,    
BLOBFIELD            BLOB             not null with default    ,   
XMLFIELD             XML            not null with default      ,   
                                            
                                            
DECFLOATFIELD        DECFLOAT        not null with default     ,   
                                            
                                            
FLOATFIELD           FLOAT           not null with default     ,   
REALFIELD            REAL             not null with default    ,   
DOUBLEFIELD          DOUBLE           not null with default    ,   
                                            
                                            
                                           
ROWIDFIELD           ROWID               ,   
DATALINKFIELD        DATALINK              not null with default   );