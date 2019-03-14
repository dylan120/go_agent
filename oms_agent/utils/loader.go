package utils

import (
	"../config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
)

func RunReflectArgsFunc(obj interface{}, funcName string, args ...interface{}) []reflect.Value {
	value := reflect.ValueOf(obj)
	funcValue := value.MethodByName(funcName)
	if funcValue.Kind() != reflect.Func {
		panic(funcName + " is not a func type")
	} else {
		input := make([]reflect.Value, len(args))
		for i, arg := range args {
			input[i] = reflect.ValueOf(arg)
		}
		return funcValue.Call(input)
	}

}

var (
	funcMap = make(map[string]interface{})
)

func LoadPlugins(input interface{}) map[string]interface{} {
	var (
		baseDir      = "/tmp"
		registerFunc map[string][]string
	)
	val := reflect.ValueOf(input)
	log.Info(val.Type().Name())
	switch val.Type().Name() {
	case "*config.MinionOptions":
		opt := input.(*config.MinionOptions)
		baseDir = opt.BaseDir
		registerFunc = opt.RegisterFunc

	case "*config.MasterOptions":
		opt := input.(*config.MasterOptions)
		baseDir = opt.BaseDir
		registerFunc = opt.RegisterFunc
	default:
		log.Panicln("LoadPlugins input type invalid")
	}
	base := filepath.Join(baseDir, "oms_agent/modules")
	files, err := ioutil.ReadDir(base)
	if !CheckError(err) {
		goRoot := os.Getenv("GOROOT")
		if goRoot == "" {
			goRoot = "/usr/local/go"
		}
		goRun := filepath.Join(goRoot, "bin/go")

		for _, f := range files {
			fileName := f.Name()
			if strings.HasSuffix(fileName, ".go") {
				name := strings.Split(fileName, ".")[0]
				pluginFile := name + ".so"

				goFilePath := filepath.Join(base, fileName)
				soFilePath := filepath.Join(base, pluginFile)
				if !CheckError(err) {
					cmd := exec.Command(goRun,
						"build", "--buildmode=plugin", "-o",
						soFilePath, goFilePath)
					cmd.Start()
					if err := cmd.Wait(); err == nil {
						plug, err := plugin.Open(soFilePath)
						for _, fname := range registerFunc[name] {
							if !CheckError(err) {
								function, _ := plug.Lookup(fname)
								fn := strings.Split(pluginFile, ".")
								funcMap[fn[0]+"."+strings.ToLower(fname)] = function
							}
						}
					} else {
						log.Error(err)
					}
				}

			}
		}
	}
	return funcMap
}
