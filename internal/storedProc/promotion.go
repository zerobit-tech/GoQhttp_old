package storedProc

import (
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/validator"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type ParamAliasRcd struct {
	Name  string
	Alias string
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type ParamPlacementRcd struct {
	Name      string
	Placement string
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type PromotionRecord struct {
	Rowid           string
	Action          string // D: Delete   R:Refresh   I:Insert
	Endpoint        string
	Storedproc      string
	Storedproclib   string
	Httpmethod      string
	UseSpecificName string
	UseWithoutAuth  string
	ParamAlias      string
	ParamAliasRcds  []*ParamAliasRcd

	ParamPlacement     string
	ParamPlacementRcds []*ParamPlacementRcd

	Namespace string

	Status              string
	StatusMessage       string
	validator.Validator `json:"-" db:"-" form:"-"`
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (p *PromotionRecord) BreakParamAlias() {
	paramALiasRcds := make([]*ParamAliasRcd, 0)
	byComa := strings.Split(p.ParamAlias, ",")

	for _, oneMap := range byComa {
		byColon := strings.Split(oneMap, ":")
		if len(byColon) == 2 {
			paramALiasRcd := &ParamAliasRcd{
				Name:  strings.ToUpper(strings.TrimSpace(byColon[0])),
				Alias: strings.ToUpper(strings.TrimSpace(byColon[1])),
			}
			paramALiasRcds = append(paramALiasRcds, paramALiasRcd)
		}
	}

	p.ParamAliasRcds = paramALiasRcds
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

func (p *PromotionRecord) BreakParamPlacements() {
	paramPlacementRcds := make([]*ParamPlacementRcd, 0)
	byComa := strings.Split(p.ParamPlacement, ",")

	for _, oneMap := range byComa {
		byColon := strings.Split(oneMap, ":")
		if len(byColon) == 2 {
			rcd := &ParamPlacementRcd{
				Name:      strings.ToUpper(strings.TrimSpace(byColon[0])),
				Placement: strings.ToUpper(strings.TrimSpace(byColon[1])),
			}
			paramPlacementRcds = append(paramPlacementRcds, rcd)
		}
	}

	p.ParamPlacementRcds = paramPlacementRcds
}
