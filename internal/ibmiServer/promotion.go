package ibmiServer

import (
	"fmt"
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) buildPromotionSQL(sp *storedProc.StoredProc) {
	sp.Promotionsql = ""
	if strings.TrimSpace(s.ConfigFile) == "" || strings.TrimSpace(s.ConfigFileLib) == "" {
		return
	}
	paramAliasMap := make([]string, 0)

	paramPlacementMap := make([]string, 0)

	for _, p := range sp.Parameters {
		if p.Alias != "" {
			paramAliasMap = append(paramAliasMap, fmt.Sprintf("%s:%s", p.Name, p.Alias))
		}
		if p.Placement != "" {
			paramPlacementMap = append(paramPlacementMap, fmt.Sprintf("%s:%s", p.Name, p.Placement))
		}

	}
	paramList := strings.Join(paramAliasMap, ", ")
	placementlist := strings.Join(paramPlacementMap, ", ")

	allowWithoutAuth := "N"
	if sp.AllowWithoutAuth {
		allowWithoutAuth = "Y"
	}

	sqlToUse := fmt.Sprintf("insert into %s.%s \n (operation,endpoint,storedproc,storedproclib,httpmethod,usespecificname,usewithoutauth,paramalias, paramplacement, namespace)", s.ConfigFileLib, s.ConfigFile)
	sqlToUse = fmt.Sprintf("%s \n values('%s','%s','%s','%s','%s','%s','%s','%s' ,'%s','%s')", sqlToUse, "I", sp.EndPointName, sp.SpecificName, sp.SpecificLib, sp.HttpMethod, "Y", allowWithoutAuth, paramList, placementlist, sp.Namespace)
	sp.Promotionsql = sqlToUse
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ListPromotion(withupdate bool) ([]*storedProc.PromotionRecord, error) {

	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.ConfigFile) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(ifnull(operation,''))) , upper(trim(ifnull(endpoint,''))), trim(ifnull(storedproc,'')), trim(ifnull(storedproclib,'')), upper(trim(ifnull(httpmethod,''))), upper(trim(ifnull(usespecificname,''))), upper(trim(ifnull(usewithoutauth,''))) , upper(trim(ifnull(paramalias,''))) , upper(trim(ifnull(paramplacement,''))), upper(trim(ifnull(namespace,''))) from %s.%s a where ifnull(status,'')=''", s.ConfigFileLib, s.ConfigFile)

		conn, err := s.GetSingleConnection()
		if err != nil {

			return promotionRecords, err
		}
		defer conn.Close()

		rows, err := conn.Query(sqlToUse)

		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return promotionRecords, err
		}
		defer rows.Close()

		for rows.Next() {
			rcd := &storedProc.PromotionRecord{}
			err := rows.Scan(&rcd.Rowid,
				&rcd.Action,
				&rcd.Endpoint,
				&rcd.Storedproc,
				&rcd.Storedproclib,
				&rcd.Httpmethod,
				&rcd.UseSpecificName,
				&rcd.UseWithoutAuth,
				&rcd.ParamAlias,
				&rcd.ParamPlacement,
				&rcd.Namespace,
			)
			if err != nil {
				rcd.Status = "E"
				rcd.StatusMessage = err.Error()
			} else {
				rcd.CheckField(validator.MustBeFromList(rcd.Action, "D", "I", "R"), "ErrorMsg", "Action: Invalid value")
				rcd.CheckField(validator.NotBlank(rcd.Endpoint), "ErrorMsg", "Endpoint: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Storedproc), "ErrorMsg", "Storedproc: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Storedproclib), "ErrorMsg", "Storedproclib: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Httpmethod), "ErrorMsg", "Httpmethod: This field cannot be blank")
				rcd.CheckField(validator.MustBeFromList(rcd.Httpmethod, "GET", "POST", "PATCH", "PUT", "DELETE"), "ErrorMsg", "Httpmethod: Invalid value")

				if rcd.Valid() {
					rcd.Status = "P"
					rcd.StatusMessage = "processing"

				} else {
					// update table with error
					rcd.Status = "E"
					rcd.StatusMessage = rcd.Validator.FieldErrors["ErrorMsg"]
				}
			}

			rcd.BreakParamAlias()
			rcd.BreakParamPlacements()
			promotionRecords = append(promotionRecords, rcd)
		}
	}
	// if withupdate && updateSQL != "" {
	// 	_, err := conn.Exec(updateSQL)
	// 	if err != nil {
	// 		log.Println("Error updateing promotion status ", err.Error())
	// 	}
	// }

	autoP, err := s.listAutoPromotion()
	if err == nil && len(autoP) > 0 {
		promotionRecords = append(promotionRecords, autoP...)
	}
	return promotionRecords, nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc {
	sp := &storedProc.StoredProc{
		EndPointName: p.Endpoint,
		HttpMethod:   p.Httpmethod,
		Name:         p.Storedproc,
		Lib:          p.Storedproclib,
		Namespace:    p.Namespace,
	}
	if p.UseSpecificName == "Y" {
		sp.UseSpecificName = true
	}

	if p.UseWithoutAuth == "Y" {
		sp.AllowWithoutAuth = true
	}
	srcd := &storedProc.ServerRecord{
		ID:   s.ID,
		Name: s.Name,
	}
	sp.DefaultServer = srcd
	sp.SetNameSpace()

	return sp
}
