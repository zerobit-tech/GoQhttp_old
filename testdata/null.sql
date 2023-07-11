create table sumitg1.nulltable (f1 char(10), f2 char(20));

insert into sumitg1.nulltable (f1) values('a');
insert into sumitg1.nulltable (f2) values('b');
insert into sumitg1.nulltable (f1,f2) values('c','d')


CREATE OR REPLACE PROCEDURE SUMITG1.NULLRS (
                IN CHARFIELD CHAR(10) DEFAULT '',
                 
                INOUT ioCHARFIELD CHAR(10) ,
                
                OUT oCHARFIELD CHAR(10) 
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.NULLRS
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        DYNAMIC RESULT SETS 1
        CALLED ON NULL INPUT
PROCBODY : BEGIN
DECLARE C1 CURSOR WITH RETURN FOR SELECT * FROM SUMITG1 . nulltable ;

--SET OCHARFIELD = IOCHARFIELD ;
 

--SET IOCHARFIELD = CHARFIELD ;
 

OPEN C1 ;
 RETURN ;
END PROCBODY 