package main

import (
	"io/ioutil"
	"log"
	"zaptest/logger"

	"gopkg.in/yaml.v3"
)

func ReadConfig(path string) (*logger.Config, error) {
	yamlF, err := ioutil.ReadFile(path)
	if err != nil {
		return &logger.Config{}, err
	}

	var config logger.Config
	err = yaml.Unmarshal(yamlF, &config)
	return &config, err

}

func main() {
	config, err := ReadConfig("logger-conf.yaml")
	if err != nil {
		log.Println(err)
	}

	logger, err := logger.NewLogger("testProject", "testService", "testBranch", "testVersion", config)
	if err != nil {
		log.Println(err)
	}

	logger.Info("Works perfectly")
	logger.Fatal("Hello")
	//fmt.Printf("%+v", config)
}
