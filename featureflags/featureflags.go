package featureflags

type Features struct {
	Promotion           bool // enable promotion logic
	TokenSync           bool // enable token sync logic
	Dashboard           bool // enable dashboard
	ParameterAlias      bool // enable parameter alias
	MaxAllowedServers   int
	MaxAllowedEndPoints int
	AllowedServerTypes  []string
}

// --------------------------------------------------------------
//
// --------------------------------------------------------------
var FeatureSetMap map[string]*Features = map[string]*Features{
	"ALL":    AllowALL(),
	"DEMO":   Demo(),
	"PUB400": Pub400(),
}

// --------------------------------------------------------------
//
// --------------------------------------------------------------
func AllowALL() *Features {

	return &Features{
		Dashboard:           true,
		Promotion:           true,
		TokenSync:           true,
		ParameterAlias:      true,
		MaxAllowedServers:   0,
		MaxAllowedEndPoints: 0,
		AllowedServerTypes:  make([]string, 0),
	}
}

// --------------------------------------------------------------
// 
// --------------------------------------------------------------
func Demo() *Features {

	return &Features{
		Dashboard:           true,
		Promotion:           false,
		TokenSync:           false,
		ParameterAlias:      false,
		MaxAllowedServers:   5,
		MaxAllowedEndPoints: 20,
		AllowedServerTypes:  []string{"IBM I"},
	}
}

// --------------------------------------------------------------
//
// --------------------------------------------------------------
func Pub400() *Features {

	return &Features{
		Dashboard:           true,
		Promotion:           false,
		TokenSync:           false,
		ParameterAlias:      false,
		MaxAllowedServers:   1,
		MaxAllowedEndPoints: 0,
		AllowedServerTypes:  []string{"IBM I"},
	}
}
