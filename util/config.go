package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	GatewayServerAddress string        `mapstructure:"GATEWAY_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	// ExpenseImageStorageDir is where receipt photos uploaded through the
	// one-time QR flow (api/expense_upload.go) are saved on local disk.
	ExpenseImageStorageDir string `mapstructure:"EXPENSE_IMAGE_STORAGE_DIR"`
	// AppLogPath is the bounded, rotating backend log file. It defaults to
	// logs/kashi.log when APP_LOG_PATH is not set.
	AppLogPath string `mapstructure:"APP_LOG_PATH"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
