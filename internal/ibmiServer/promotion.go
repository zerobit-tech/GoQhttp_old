package ibmiServer

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
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

	sqlToUse := fmt.Sprintf("insert into %s.%s \n (operation,endpoint,storedproc,storedproclib,httpmethod,usespecificname,usewithoutauth,paramalias, paramplacement)", s.ConfigFileLib, s.ConfigFile)
	sqlToUse = fmt.Sprintf("%s \n values('%s','%s','%s','%s','%s','%s','%s','%s' ,'%s')", sqlToUse, "I", sp.EndPointName, sp.SpecificName, sp.SpecificLib, sp.HttpMethod, "Y", allowWithoutAuth, paramList, placementlist)
	sp.Promotionsql = sqlToUse
}
