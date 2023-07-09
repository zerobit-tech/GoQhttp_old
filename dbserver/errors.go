package dbserver

func (s *Server) ErrorToHttpStatus(inerr error) (int, string,string, bool) {
	return s.GetDbDriver().ErrorToHttpStatus(inerr)

}
