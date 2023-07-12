DELIMITER //
CREATE PROCEDURE spinoutparm(
    IN input1 int,
	inout inout1 int,
    OUT output1 int
     )
BEGIN
     
     set output1 = inout1+10;
	 set inout1 = input1;
END //
DELIMITER ;