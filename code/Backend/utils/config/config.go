package config

var (
	MaxSheetCache	int64
)


func LoadConfig(path string) {
	MaxSheetCache = 256 * (1 << 20)			// 256M
}