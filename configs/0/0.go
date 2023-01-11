package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Password string `json:"password"`
	} `json:"database"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func main() {
	// Открываем файл конфигурации на чтение
	file, err := os.Open("config.json")
	// Конфиг очень важная часть приложения,
	// поэтому если не можем его прочитать,
	// то прерываем выполнение программы
	if err != nil {
		panic(err)
	}
	// Закрываем файл после выполнения функции
	defer file.Close()

	// Создаем переменную типа Config
	config := Config{}
	// создаем декодер JSON
	decoder := json.NewDecoder(file)
	// Декодируем содержимое файла в переменную config
	err = decoder.Decode(&config)
	// Аналогично, паника если не можем декодировать config
	if err != nil {
		panic(err)
	}
	// выводим значение переменной config
	fmt.Println(config.Database.Host)
	fmt.Println(config.Database.Password)
	fmt.Println(config.Host)
	fmt.Println(config.Port)
}
