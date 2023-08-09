CREATE OR REPLACE PROCEDURE SUMITG1.SPTABLELIST (
                IN PageNumber numeric(5,0) DEFAULT 1,
                IN PageSize numeric(5,0) DEFAULT 10,
                    OUT NextPageNumber numeric(5,0) 
            )
        LANGUAGE SQL
        SPECIFIC SUMITG1.SPTABLELIST
        NOT DETERMINISTIC
        MODIFIES SQL DATA
        DYNAMIC RESULT SETS 1
        CALLED ON NULL INPUT
PROCBODY : BEGIN
DECLARE C1 CURSOR WITH RETURN FOR select char(table_name) as table_name, table_owner, table_type from qsys2.systables limit pagesize offset (PageNumber-1)* PageSize;

set NextPageNumber = PageNumber+1;
 
OPEN C1 ;
RETURN ;
END PROCBODY ;


call SUMITG1.SPTABLELIST(2,15); 