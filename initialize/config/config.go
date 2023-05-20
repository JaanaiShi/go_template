package config

import (
	"fmt"
	"os"
	"path"

	"github.com/JaanaiShi/flint/common"
	"github.com/spf13/viper"
)

func Init() {
	var (
		err      error
		fileName string
		config   common.Config
	)
	v := viper.New()
	mode := os.Getenv("ServerMode")
	fmt.Println(os.Getwd())
	if mode == "" {
		fileName = "config.yml"
	} else {
		fileName = "config-" + mode + ".yml"
	}

	// 文件路径如何设置
	v.SetConfigFile(path.Join("config", fileName))
	if err = v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err = v.Unmarshal(&config); err != nil {
		panic("unmarshal err: " + err.Error())
	}

	common.Conf = &config

}
