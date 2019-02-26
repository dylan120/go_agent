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
	getValue := reflect.ValueOf(obj)
	funcValue := getValue.MethodByName(funcName)
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
	funcMap map[string]interface{}
)

func LoadPlugins(opt *config.MinionOptions) map[string]interface{} {
	base := filepath.Join(opt.BaseDir, "oms_agent/modules")
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
				plug, err := plugin.Open(soFilePath)
				cmd := exec.Command(goRun,
					"build", "--buildmode=plugin", "-o",
					soFilePath, goFilePath)
				out, err := cmd.CombinedOutput()
				if !CheckError(err) {
					for _, fname := range opt.RegisterFunc[name] {
						symbol, err := plug.Lookup(fname)
						if !CheckError(err) {
							val := reflect.ValueOf(symbol).Elem()
							funcs := val.Interface().([]string)
							for _, f := range funcs {
								function, _ := plug.Lookup(f)
								fn := strings.Split(pluginFile, ".")
								funcMap[fn[0]+"."+strings.ToLower(f)] = function
							}
						}
					}
				} else {
					log.Error(string(out))
				}
			}
		}
	}
	return funcMap
}
