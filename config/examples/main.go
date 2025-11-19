package main

import (
	"fmt"

	"trellis.tech/trellis/common.v3/config"
)

/*
include: "test.yaml"
app:
  name: "my-app"
  version: "1.0.0"
database:
  host: "localhost"
  port: 5432
users:
  - name: "user1"
    age: 30
  - name: "user2"
    age: 25
  - name: "${test.name}"
    age: "${test.age}"
*/

type Config struct {
	App      App    `yaml:"app"`
	Database User   `yaml:"database"`
	Users    []User `yaml:"users"`
}
type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Database struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type User struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func main() {
	cfg, err := config.NewConfig("config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)

	cnf := &Config{}
	cfg.Object(cnf)
	fmt.Println(cnf)
}
