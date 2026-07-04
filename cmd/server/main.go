	package main

	import (
		"errors"
		"fmt"
		"net/http"
		"os"
		"os/signal"
		"strconv"
		"syscall"

		"github.com/joho/godotenv"
		"github.com/rezect/go-interview/internal/server"
	)

	func main() {
		godotenv.Load()

		var addr string
		port := os.Getenv("PORT")
		if _, err := strconv.Atoi(port); err != nil {
			addr = ":8000"
		} else {
			addr = ":" + port
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		srv := server.NewServer(addr)

		go func() {
			fmt.Printf("Запуск сервера на порту %v\n", addr)
			if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Ошибка запуска сервера: %v\n", err)
			}
		}()

		// Мягкое завершение работы сервера
		<- sigCh

		fmt.Printf("Завершение работы сервера\n")
		
		if err := srv.Shutdown(); err != nil {
			fmt.Printf("Ошибка завершения сервера: %v\n", err)
		}

		fmt.Printf("Сервер успешно остановлен\n")
	}