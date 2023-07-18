package featureflags

type Features struct {
	AllowPromotion      bool // enable promotion logic
	AllowTokenSync      bool // enable token sync logic
	Dashboard           bool // enable dashboard
	AllowParameterAlias bool // enable parameter alias
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
		AllowPromotion:      true,
		AllowTokenSync:      true,
		AllowParameterAlias: true,
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
		AllowPromotion:      false,
		AllowTokenSync:      false,
		AllowParameterAlias: false,
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
		AllowPromotion:      false,
		AllowTokenSync:      false,
		AllowParameterAlias: false,
		MaxAllowedServers:   1,
		MaxAllowedEndPoints: 0,
		AllowedServerTypes:  []string{"IBM I"},
	}
}
