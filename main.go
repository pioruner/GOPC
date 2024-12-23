package main

import (
	"OPC/loger"
	"OPC/server"
)

func main() {
	logs := loger.Create("server") //Создание лога
	logs.Println("App started")
	go server.Start("localhost:30301") //Запуск сервера
	logs.Println("Server started")
	select {} // Бесконечный цикл
}
