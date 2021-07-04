package config

import (
	"backend/utils/logger"
	"github.com/jinzhu/configor"
)

type config struct{
	MaxSheetCache	int64		`default:"3"`
	UnitCache		int64		`default:"1"`
	WriteThrough	bool		`default:"true"`
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

	logger.Debug(cfg)
}

func Get() *config {
	return &cfg
}