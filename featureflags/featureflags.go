package featureflags

type Features struct {
	AllowPromotion      bool // enable promotion logic
	AllowTokenSync      bool // enable token sync logic
	Dashboard           bool // enable dashboard
	AllowParameterAlias bool // enable parameter alias
	MaxAllowedServers   int
	MaxAllowedEndPoints int
	AllowedServerTypes  []string
	LoginMessages       []string
	AdminEmail          string
	AdminPassword       string
	AllowHtmlTemplates  bool
	AllowParamPlacement bool
	AllowLibList        bool
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
		LoginMessages:       make([]string, 0),
		AdminEmail:          "",
		AdminPassword:       "",
		AllowParamPlacement: true,
		AllowLibList:        true,
	}
}

// --------------------------------------------------------------
//
// --------------------------------------------------------------
func Demo() *Features {

	fset := AllowALL()

	fset.AllowPromotion = false
	fset.AllowTokenSync = false
	fset.AllowParameterAlias = false
	fset.MaxAllowedServers = 5
	fset.MaxAllowedEndPoints = 20
	fset.AllowedServerTypes = []string{"IBM I"}

	return fset
}

// --------------------------------------------------------------
//
// --------------------------------------------------------------
func Pub400() *Features {

	return &Features{
		AllowParamPlacement: true,
		AllowLibList:        false,

		Dashboard:           true,
		AllowPromotion:      false,
		AllowTokenSync:      false,
		AllowParameterAlias: true,
		MaxAllowedServers:   1,
		MaxAllowedEndPoints: -1,
		AllowedServerTypes:  []string{"IBM I"},
		AdminEmail:          "admin2@example.com",
		AdminPassword:       "SaveAdmin#2023",

		LoginMessages: []string{
			"Thanks to Pub400.com",
			"Please be respectful to other users.",
			"This is a Demo app \nand depends on Pub400.com \navailbility and speed.",

			"\n\nLogin Email: demo@example.com",
			"Password: demopass",

			"\n\nFor support email to: support@zerobit.tech",
		},
	}
}
