package config

import (
	"backend/utils/logger"
	"github.com/jinzhu/configor"
)

type config struct {
	MaxSheetCache	int64		`default:"2"`
	UnitCache		int64		`default:"1"`
	ZKRoot			string		`default:"/backend"`
	ZKAddr			string		`required:"true"`
	Addr			string		`required:"true"`
	MySqlAddr		string		`required:"true"`
	JWTSharedKey	string		`required:"true"`
}

var cfg config

func LoadConfig() {
	if err := configor.New(&configor.Config{ENVPrefix: "GDOC"}).Load(&cfg); err != nil {
		panic(err)
	}

	cfg.MaxSheetCache <<= 20
	cfg.UnitCache <<= 20

	logger.Infof("Config: %+v", cfg)
}

func Get() *config {
	return &cfg
}