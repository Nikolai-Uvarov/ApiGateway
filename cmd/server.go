package main

import (
	"log"
	"net/http"
	"APIGateway/pkg/api"
)

func main() {


	// Создание объекта API
	api := api.New()

	// Запуск сетевой службы и HTTP-сервера на всех локальных IP-адресах на порту 8080.
	err := http.ListenAndServe(":9092", api.Router())
	if err != nil {
		log.Fatal(err)
	}

}
