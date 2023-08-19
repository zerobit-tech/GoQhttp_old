package ibmiServer

type LibList struct {
	Pos     int
	Lib     string
	Libtype string
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetLibList() ([]*LibList, error) {

	libList := make([]*LibList, 0)

	sqlToUse := "SELECT ORDINAL_POSITION, SYSTEM_SCHEMA_NAME, TYPE FROM QSYS2.LIBRARY_LIST_INFO order by ORDINAL_POSITION"

	conn, err := s.GetSingleConnection()
	if err != nil {

		return libList, err
	}
	defer conn.Close()

	rows, err := conn.Query(sqlToUse)

	if err != nil {

		return libList, err
	}
	defer rows.Close()

	for rows.Next() {
		libE := &LibList{}

		err := rows.Scan(&libE.Pos, &libE.Lib, &libE.Libtype)
		if err != nil {
			return libList, err
		} else {
			libList = append(libList, libE)
		}

	}

	return libList, nil
}
