package resolvers

import (
	"booking"
	"booking/util"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type configResolver struct{ *Resolver }

func (r *queryResolver) Config(ctx context.Context) (resp booking.ClientConfig, err error) {
	viper.SetConfigFile("./conf/client_config.yaml") // 如果指定了配置文件，则解析指定的配置文件
	viper.SetConfigType("yaml")                      // 设置配置文件格式为YAML
	if err := viper.ReadInConfig(); err != nil {     // viper解析配置文件
		return resp, err
	}

	prompt := viper.GetString("prompt")
	appId := viper.GetString("wxAppID")
	appSecret := viper.GetString("wxSecret")
	resp.Prompt = &prompt
	resp.WxAppID = &appId
	resp.WxSecret = &appSecret

	return resp, nil
}

func (r *mutationResolver) Config(ctx context.Context, input booking.ConfigInput) (resp booking.ClientConfig, err error) {

	filename := "conf/client_config.yaml"
	if util.Exists(filename) {
		err := util.MoveFile(filename, "conf/old/client_config-"+time.Now().Format("20060102-150405")+".yaml")
		if err != nil {
			fmt.Println("cannot move file to new directory" + err.Error())
			return resp, errors.New("更新配置文件时出现错误 :" + err.Error())
		}
	}

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return resp, errors.New("更新配置文件时出现错误 :" + err.Error())
	}

	config := &booking.ClientConfig{}

	if input.WxAppID != nil && len(*input.WxAppID) != 0 {
		config.WxAppID = input.WxAppID
	}

	if input.WxSecret != nil && len(*input.WxSecret) != 0 {
		config.WxSecret = input.WxSecret
	}

	if input.Prompt != nil && len(*input.Prompt) != 0 {
		config.Prompt = input.Prompt
	}

	d, err := yaml.Marshal(config)
	if err != nil {
		return resp, errors.New("更新配置文件时出现错误 :" + err.Error())
	} else {
		_, err := f.Write(d)
		if err != nil {
			return resp, errors.New("更新配置文件时出现错误 :" + err.Error())
		}
	}

	return *config, nil
}
