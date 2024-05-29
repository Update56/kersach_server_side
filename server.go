package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Hello Server!")
	//открытие порта сервера по протоколу tcp
	listener, err := net.Listen("tcp", ":56565")
	//првоерка на ошибку при открытии
	if err != nil {
		fmt.Println(err)
		return
	}
	//закрытие в конце выполнения программы
	defer listener.Close()
	fmt.Println("Server is listening...")

	for {
		//создание потока при подтверждении подключении
		conn, err := listener.Accept()
		//проверка на ошибку
		if err != nil {
			fmt.Println(err)
			continue
		}
		connProc(conn)
	}
}

func connProc(conn net.Conn) {
	defer conn.Close()
	for {
		line := make([]byte, 1024)
		buf := bufio.NewReader(conn)
		n, err := buf.Read(line)
		//проверка на считывание и ошибку
		if n == 0 || err != nil {
			fmt.Println("Read error:", err)
			break
		}
		conn.Write([]byte("Hello " + string(line[:n]) + "! I'am Server"))
	}
}
