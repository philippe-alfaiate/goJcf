package goJcf

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type GoJCFConfig struct {
	Path          string
	EraseOnFail   bool
	ConfigDefault interface{}
}

type GoJCFErrorType int

const (
	ErrorOpen GoJCFErrorType = iota
	ErrorRead
	ErrorDefaultNil
	ErrorDefaultFaileMarshall
	ErrorReset
)

type GoJCFError struct {
	Type    GoJCFErrorType
	Error   error
	Details map[string]interface{}
}

func createError(jcfType GoJCFErrorType, err error) *GoJCFError {
	var e GoJCFError

	e.Type = jcfType
	e.Error = err
	e.Details = make(map[string]interface{})

	return &e
}

func (e *GoJCFError) changeError(err error) {
	e.Error = err
}

func (e *GoJCFError) addDetail(key string, value interface{}) {
	e.Details[key] = value
}

func (e *GoJCFError) Equal(jcfType GoJCFErrorType) bool {
	return e.Type == jcfType
}

func GetConfig(jcfConfig *GoJCFConfig, configOutput interface{}) *GoJCFError {

	//Default value of config
	jcf := &GoJCFConfig{
		Path:          "config.json",
		EraseOnFail:   true,
		ConfigDefault: nil,
	}

	if jcfConfig != nil {
		jcf = jcfConfig
	}

	configFile, err := os.OpenFile(jcf.Path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		e := createError(ErrorOpen, errors.New("JsonConfigFile: Config file could not be created or open"))
		e.addDetail("path", jcf.Path)
		e.addDetail("error", err)
		return e
	}
	defer configFile.Close()

	configByte, err := ioutil.ReadAll(configFile)
	if err != nil {
		e := createError(ErrorRead, errors.New("JsonConfigFile: Fail to read file already opened, weird!! "))
		e.addDetail("error", err)
		return e
	}

	err = json.Unmarshal(configByte, &configOutput)
	if err != nil {

		if jcf.ConfigDefault == nil && !jcf.EraseOnFail {
			e := createError(ErrorDefaultNil, nil)
			e.changeError(errors.New("JsonConfigFile: No default value set"))
			return e
		} else if jcf.ConfigDefault == nil {
			jcf.ConfigDefault = configOutput
		}

		configNew, err := json.Marshal(jcf.ConfigDefault)
		if err != nil {
			e := createError(ErrorDefaultFaileMarshall, errors.New("JsonConfigFile: Fail to marshal Json"))
			e.addDetail("error", err)
			return e
		} else if jcf.EraseOnFail {
			configFile.Truncate(0)
			configFile.Seek(0, 0)
			configFile.WriteString(string(configNew))
			configOutput = jcf.ConfigDefault
			e := createError(ErrorReset, errors.New("JsonConfigFile: Config json reset to default "))
			return e
		}
	}

	return nil
}
