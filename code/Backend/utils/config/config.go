package config

var (
	MaxSheetCache	int64
	WriteThrough	bool
)


func LoadConfig(path string) {
	MaxSheetCache = 256 * (1 << 20)			// 256M
	WriteThrough = true
}