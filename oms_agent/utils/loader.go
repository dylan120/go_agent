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

var funcMap = make(map[string]interface{})

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
				pluginFile := strings.Replace(fileName, ".go", ".so", 1)

				goFilePath := filepath.Join(base, fileName)
				soFilePath := filepath.Join(base, pluginFile)
				log.Debug(goRun)
				cmd := exec.Command(goRun,
					"build", "--buildmode=plugin", "-o",
					soFilePath, goFilePath)
				out, err := cmd.CombinedOutput()
				if !CheckError(err) {
					plug, err := plugin.Open(soFilePath)
					//if plug == nil {
					//	plug, err = plugin.Open(soFilePath)
					//}
					if !CheckError(err) {
						symbol, err := plug.Lookup("RegisterFunc")
						if !CheckError(err) {
							val := reflect.ValueOf(symbol).Elem()
							funcs := val.Interface().([]string)
							for _, f := range funcs {
								function, _ := plug.Lookup(f)
								fn := strings.Split(pluginFile, ".")
								funcMap[fn[0]+"."+strings.ToLower(f)] = function
							}
						}
						//*(**int)(unsafe.Pointer(plug)) = nil
					}
				} else {
					log.Error(string(out))
				}
			}

			if strings.HasSuffix(fileName, ".so") {

			}
		}
	}
	return funcMap
}
