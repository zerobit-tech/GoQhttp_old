create table sumitg1.nulltable2 (f1 char(10), f2 decimal(5,0), f3 date , f4 timestamp);

insert into sumitg1.nulltable2 (f1) values('a');
insert into sumitg1.nulltable2 (f2) values(15);
insert into sumitg1.nulltable2 (f3) values(current_date);
insert into sumitg1.nulltable2 (f4) values(current_timestamp);

insert into sumitg1.nulltable2   values('c','250','2015-12-21', current_timestamp)

CREATE OR REPLACE PROCEDURE SUMITG1.NULLRS2 (
                IN CHARFIELD CHAR(10) DEFAULT '',
                 
                INOUT ioCHARFIELD CHAR(10) ,
                 inOUT io2 decimal(10) ,
                inOUT io3 date,

                inOUT io4 timestamp, 


                OUT oCHARFIELD CHAR(10) ,
                OUT o2 decimal(10) ,
                OUT o3 date,

                OUT o4 timestamp,
                                OUT o5 time
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.NULLRS2
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        DYNAMIC RESULT SETS 1
        CALLED ON NULL INPUT
PROCBODY : BEGIN
DECLARE C1 CURSOR WITH RETURN FOR SELECT * FROM SUMITG1 . nulltable2 ;

--SET OCHARFIELD = IOCHARFIELD ;
 

--SET IOCHARFIELD = CHARFIELD ;
 

OPEN C1 ;
 RETURN ;
END PROCBODY 