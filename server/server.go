package server

import (
	"OPC/loger"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"
)

const (
	read  byte = 0
	write byte = 1
)

type Message struct {
	CMD   byte    `json:"cmd"`
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

func Start(address string) {
	logs := loger.Create("server") //Создание лога
	logs.Println("Server started")

	//Создание map для хранения данных
	var data sync.Map
	data.Store("test", -99.99)
	//value, ok := data.Load("test")
	//data.Delete("test")

	logs.Println("Starting TCP server on", address)
	// Запускаем сервер
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logs.Println("Error starting server:", err)
		return
	}
	defer func() {
		if err := listener.Close(); err != nil {
			logs.Printf("Error closing listener: %v", err)
		}
	}()

	for {
		conn, err := listener.Accept() // Принимаем новое соединение
		if err != nil {
			logs.Println("Error accepting connection:", err)
			continue
		}
		logs.Println("New connection from", conn.RemoteAddr())
		go handleConnection(conn, &data, logs) // Обрабатываем соединение в отдельной go
	}
}

func handleConnection(conn net.Conn, data *sync.Map, logs *log.Logger) {
	defer func() { // закрытие соединения
		if err := conn.Close(); err != nil {
			logs.Printf("Error closing connection: %v", err)
		}
	}()
	buffer := make([]byte, 1024) // Буфер для чтения данных
	var msg Message              // Наша структура пакета
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			logs.Println("Connection closed by", conn.RemoteAddr())
			break
		}
		if err != nil {
			logs.Println("Error reading from connection:", err)
			break
		}

		// Выводим полученные данные в hex формате
		logs.Printf("Received from %s: %s\n", conn.RemoteAddr(), hex.EncodeToString(buffer[:n]))

		err = json.Unmarshal(buffer[:n], &msg) // Декодируем данные из JSON
		if err != nil {
			logs.Println("Error decoding JSON:", err)
			continue
		} else {
			logs.Printf("Received from %s: %+v\n", conn.RemoteAddr(), msg)
			answer(msg, data, conn, logs)
		} // Выводим результат

	}
}

func answer(msg Message, data *sync.Map, conn net.Conn, logs *log.Logger) {
	switch msg.CMD {
	case read:
		value, ok := data.Load(msg.Key)
		if !ok {
			logs.Printf("Key %s not found\n", msg.Key)
			return
		}

		floatValue, ok := value.(float64)
		if !ok {
			logs.Printf("Key %s has non-float64 value: %v\n", msg.Key, value)
			return
		}

		ans := Message{CMD: read, Key: msg.Key, Value: floatValue}
		buff, err := json.Marshal(&ans) // Кодируем ответ в JSON
		if err != nil {
			logs.Println("Error encoding JSON:", err)
			return
		}

		_, err = conn.Write(buff) // Отправляем ответ клиенту
		if err != nil {
			logs.Printf("Error sending response to %s: %v\n", conn.RemoteAddr(), err)
			return
		}

		logs.Printf("Sent to %s: %+v\n", conn.RemoteAddr(), ans)
		// Выводим результат

	case write:
		data.Store(msg.Key, msg.Value)
		logs.Printf("Key %s set to %v\n", msg.Key, msg.Value)

		ans := Message{CMD: write, Key: msg.Key, Value: msg.Value}
		buff, err := json.Marshal(&ans) // Кодируем подтверждение
		if err != nil {
			logs.Println("Error encoding JSON:", err)
			return
		}

		_, err = conn.Write(buff) // Отправляем подтверждение клиенту
		if err != nil {
			logs.Printf("Error sending response to %s: %v\n", conn.RemoteAddr(), err)
			return
		}

		logs.Printf("Sent to %s: %+v\n", conn.RemoteAddr(), ans)
	}
}
