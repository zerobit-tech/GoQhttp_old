package models

import "github.com/onlysumitg/GoQhttp/internal/validator"

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type ConditionAndGroupMap struct {
	ID          string `json:"id" db:"id" form:"id"`
	GroupID     string `json:"groupid" db:"groupid" form:"groupid"`
	ConditionId string `json:"conditionid" db:"conditionid" form:"conditionid"`
	validator.Validator
}
