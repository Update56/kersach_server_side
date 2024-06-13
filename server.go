package main

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// структура сообщения
type Message struct {
	Sender   string
	Receiver string
	Text     string
}

// переменная файла логирования

// зарезервированное спец.имя клиента
const clientName string = "client"

// зарезервированное спец.имя сервера
const serverName string = "server"

func main() {
	//загружаем пару сертификат/ключ
	cer, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	fmt.Println("Hello, I'am Server!")
	//открытие порта сервера по протоколу tcp с tls
	listener, err := tls.Listen("tcp", ":56565", config)
	//првоерка на ошибку при открытии
	if err != nil {
		fmt.Println(err)
		return
	}
	//закрытие порта в конце выполнения программы
	defer listener.Close()
	//открываем файл для записи
	file, err := os.OpenFile((time.Now().Format("02.01.06_15.04.05_") + "server.log"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	//создаём логер
	infoLog := log.New(file, "INFO ", log.Ldate|log.Ltime)
	infoLog.Println("Server is listening...")

	//создание хэш-таблицы ключ - никнейм, значение - подключение
	connMap := make(map[string]net.Conn)

	for {
		//подтверждение подключения
		conn, err := listener.Accept()
		//проверка на ошибку
		if err != nil {
			fmt.Println(err)
			continue
		}
		//вызываем функцию в отдельной горутине
		go connProc(conn, connMap, infoLog)
	}
}

func connProc(conn net.Conn, connMap map[string]net.Conn, infoLog *log.Logger) {
	//3.Отправка изменённого списка
	defer sendCurrentList(connMap)
	//2.Закрытие соединения
	defer conn.Close()
	//объявляем входной декодер с потоком в виде подключения
	dec := gob.NewDecoder(conn)
	//пустая структура для сообщения
	recMess := Message{}
	//десериализация полученного "нулевого" сообщения
	dec.Decode(&recMess)
	//объявляем новую запись в хэш-мапе
	connMap[recMess.Sender] = conn
	//лог в консоль
	fmt.Println("User:", recMess.Sender, "connected")
	//лог в файл
	infoLog.Println("User:", recMess.Sender, "connected")

	sendCurrentList(connMap)
	//1.Удаление записи
	defer delete(connMap, recMess.Sender)

	for {
		//объявляем входной декодер с потоком
		dec := gob.NewDecoder(conn)
		//пустая структура для сообщения
		recMess := Message{}
		//десериализация полученного сообщения
		dec.Decode(&recMess)
		//обработка отключения от сервера
		if recMess.Receiver == serverName && len(recMess.Text) >= len("Disconnect") {
			if recMess.Text[:len("Disconnect")] == "Disconnect" {
				//логирование
				infoLog.Println("User:", recMess.Sender, "disconnected")
				fmt.Println("User:", recMess.Sender, "disconnected")
				return
			}
		}
		//создаём выходной енкодер с соединение соотвествующим отправителю
		enc := gob.NewEncoder(connMap[recMess.Receiver])
		//лог сообщения
		infoLog.Println(recMess.Receiver, "to", recMess.Receiver, ": ", recMess.Text)
		//передача сообщения
		enc.Encode(recMess)
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

// функция отправки текущего списка пользователей
func sendCurrentList(connMap map[string]net.Conn) {
	//формируем сообщение со текущим списком пользователей
	currentList := formSpecialMessages("001", keysToString(connMap))
	//проходимся по хэш-таблице и всем отправляем актуалный список пользователей
	for _, conn := range connMap {
		enc := gob.NewEncoder(conn)
		enc.Encode(currentList)
	}
}
