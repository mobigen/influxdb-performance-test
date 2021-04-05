package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// GetAbsPath 프로세스 실행위치를 기반으로 절대 경로 정보를 반환
func GetAbsPath(input string) string {
	/* 프로세스가 실행된 위치의 절대경로 정보 */
	if input == "." {
		path, _ := os.Getwd()
		return path
	}

	if strings.Index(input, "./") == 0 || strings.Index(input, "/") != 0 {
		if strings.Index(input, "./") == 0 {
			input = strings.TrimLeft(input, "./")
		}
		path, _ := os.Getwd()
		if path == "/" {
			path = fmt.Sprintf("/%s", input)
		} else {
			if input == "" {
				path = fmt.Sprintf("%s", path)
			} else {
				path = fmt.Sprintf("%s/%s", path, input)
			}
		}
		return path
	}

	if len(input) > 1 {
		if (len(input) - 1) == strings.LastIndex(input, "/") {
			input = strings.TrimRight(input, "/")
		}
	}
	return input
}

// ReadYaml read yaml file to struct
func ReadYaml(filePath string, obj interface{}) error {
	readFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "fail to read( config file )")
	}
	err = yaml.Unmarshal(readFile, obj)
	if err != nil {
		return err
	}
	return nil
}

// PrintYaml print yaml format
func PrintYaml(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	// encoder.SetIndent("", "    ")
	return encoder.Encode(data)
}
