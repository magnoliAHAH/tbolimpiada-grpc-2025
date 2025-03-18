package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen" // Должно совпадать с go_package в .proto
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	client := pb.NewProcessingServiceClient(conn)

	resp, err := client.ProcessFile(context.Background(), &pb.ProcessRequest{
		Filename: "test.txt",
		Content:  []byte("Hello, world!"),
	})
	if err != nil {
		log.Fatalf("Ошибка при вызове ProcessFile: %v", err)
	}

	fmt.Printf("Ответ сервера: %s\n", resp.MeetingPoint)
}
