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
	pluginMap = make(map[string]*plugin.Plugin)
	funcMap   = make(map[string]interface{})
)

func InitPlugins(opt *config.MinionOptions) {
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
				plug, err := plugin.Open(soFilePath)
				cmd := exec.Command(goRun,
					"build", "--buildmode=plugin", "-o",
					soFilePath, goFilePath)
				out, err := cmd.CombinedOutput()
				if !CheckError(err) {
					pluginMap[fileName] = plug
				} else {
					log.Error(string(out))
				}
			}
		}
	}
	//return funcMap
}

func LoadFunc(funcName string) interface{} {
	a := strings.Split(funcName, ".")
	if _, exist := funcMap[funcName]; !exist {
		for _, plugin := range pluginMap {
			function, err := plugin.Lookup(a[1])
			if !CheckError(err) {
				funcMap[funcName] = function
				break
			}
		}
	}
	return funcMap[funcName]
}
