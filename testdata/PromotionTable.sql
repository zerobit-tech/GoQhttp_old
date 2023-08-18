create or replace table sumitg1.promotiontable (operation char(1) , endpoint char(40),
 storedproc char(20), storedproclib char(10), httpmethod char(20), 
 usespecificname char(1), usewithoutauth char(1), paramalias varchar(200), 
 paramplacement varchar(200), namespace varchar(100), 
 Status char(1), StatusMessage varchar(100));






create table chains.promotiontable2 (operation char(1) , endpoint char(40),
 storedproc char(20), storedproclib char(10), httpmethod char(20), 
 usespecificname char(1), usewithoutauth char(1), paramalias varchar(200), 
 Status char(1), StatusMessage varchar(100));

insert into promotiontable 
 (operation,endpoint,storedproc,storedproclib,httpmethod,usespecificname,usewithoutauth,paramalias,paramplacement,namespace) 
 values('I','SELECTALLCUSTOMERS222','SELECTALLCUSTOMERS','TESTDB','POST','N','N','','','')