package zap_log

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var zapFakeLog *ZapLoggerFake

func init() {
	zapFakeLog = initLogger(nil)
}

// InitZapLog 默认是
func InitZapLog(configFilePath string) error {

	if configFilePath == "" {
		zapFakeLog = initLogger(nil)
		return nil
	}

	//configFilePath := "../conf/log.yaml"
	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return err
	}

	getObjFromYaml(configFilePath)

	return nil
}

func GetFakeLogger() *ZapLoggerFake {
	if zapFakeLog == nil {
		panic(errors.New("please init zap,use .zap_log.InitZapLog"))
	}

	return zapFakeLog
}

func getObjFromYaml(configFilePath string) {

	var configObj *Config

	bytes, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		panic(err)
	}

	var info *DuInfo

	if err = yaml.Unmarshal(bytes, &info); err != nil {
		panic(err)
	}

	configObj = info.Config

	zapFakeLog = initLogger(configObj)

	return

}
