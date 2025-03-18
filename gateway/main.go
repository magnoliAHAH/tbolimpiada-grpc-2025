package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Подключаемся к gRPC серверу
var client pb.ProcessingServiceClient

func main() {
	conn, err := grpc.NewClient("loadbalancer:80", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024*1024*50)))
	if err != nil {
		log.Fatalf("Ошибка подключения к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client = pb.NewProcessingServiceClient(conn)

	// HTTP обработчик для загрузки файлов
	http.HandleFunc("/upload", uploadHandler)

	log.Println("HTTP Gateway запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// uploadHandler обрабатывает загрузку файлов
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST-запросы", http.StatusMethodNotAllowed)
		return
	}

	// Читаем файл из запроса
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Читаем содержимое файла в буфер
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		http.Error(w, "Ошибка при чтении файла", http.StatusInternalServerError)
		return
	}

	// Отправляем файл на gRPC-сервер
	resp, err := client.ProcessFile(context.Background(), &pb.ProcessRequest{
		Filename: header.Filename,
		Content:  buf.Bytes(),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка вызова gRPC: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат клиенту
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Файл обработан, точка сбора: %s\n", resp.MeetingPoint)
	fmt.Fprintf(w, "Изображение путей: %s\n", resp.ImageUrl)

}
