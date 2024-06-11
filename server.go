package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"strings"
	"time"
)

// структура сообщения
type Message struct {
	Sender   string
	Receiver string
	Text     string
}

// зарезервированное спец.имя клиента
const clientName string = "client"

// зарезервированное спец.имя сервера
const serverName string = "server"

func main() {
	fmt.Println("Hello, I'am Server!")
	//открытие порта сервера по протоколу tcp
	listener, err := net.Listen("tcp", ":56565")
	//првоерка на ошибку при открытии
	if err != nil {
		fmt.Println(err)
		return
	}
	//закрытие порта в конце выполнения программы
	defer listener.Close()
	fmt.Println("Server is listening...")

	//создание хэш-таблицы ключ - никнейм, значение - подключение
	connMap := make(map[string]net.Conn)

	for i := 0; ; i++ {
		//подтверждение подключения
		conn, err := listener.Accept()
		//проверка на ошибку
		if err != nil {
			fmt.Println(err)
			continue
		}
		connMap[string(i)] = conn
		//вызываем функцию в отдельной горутине
		go connProc(conn, connMap)

	}
}

func connProc(conn net.Conn, connMap map[string]net.Conn) {
	defer conn.Close()
	enc := gob.NewEncoder(conn)
	enc.Encode(formSpecialMessages("001", keysToString(connMap)))
	for {
		time.Sleep(time.Second * 5)
	}
}

// функция копирования ключей хэш-мапы в строку
func keysToString(connMap map[string]net.Conn) string {

	//cоздание слайса для хранения ключей
	keys := make([]string, 0, len(connMap))

	//заполнение слайса ключами мапы
	for key := range connMap {
		keys = append(keys, key)
	}
	//объединение ключей в одну строку
	result := strings.Join(keys, "\n")
	//возвращаем полученную строку
	return result
}

// функция формирования специального сообщения
func formSpecialMessages(code string, text string) Message {
	messText := (code + "\n" + text)
	return Message{serverName, clientName, messText}
}
