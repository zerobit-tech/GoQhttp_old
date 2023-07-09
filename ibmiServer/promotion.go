package ibmiServer

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *IBMiServer) BuildPromotionSQL(sp *storedProc.StoredProc) {
	sp.Promotionsql = ""
	if strings.TrimSpace(s.ConfigFile) == "" || strings.TrimSpace(s.ConfigFileLib) == "" {
		return
	}
	paramAliasMap := make([]string, 0)

	for _, p := range sp.Parameters {
		if p.Alias != "" {
			paramAliasMap = append(paramAliasMap, fmt.Sprintf("%s:%s", p.Name, p.Alias))
		}
	}
	paramList := strings.Join(paramAliasMap, ", ")

	allowWithoutAuth := "N"
	if sp.AllowWithoutAuth {
		allowWithoutAuth = "Y"
	}

	sqlToUse := fmt.Sprintf("insert into %s.%s \n (action,,endpoint,storedproc,storedproclib,httpmethod,usespecificname,usewithoutauth,paramalias)", s.ConfigFileLib, s.ConfigFile)
	sqlToUse = fmt.Sprintf("%s \n values('%s','%s','%s','%s','%s','%s','%s','%s')", sqlToUse, "I", sp.EndPointName, sp.SpecificName, sp.SpecificLib, sp.HttpMethod, "Y", allowWithoutAuth, paramList)
	sp.Promotionsql = sqlToUse
}
