package main

import (
	"context"
	"log"
	"net"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedProcessingServiceServer
}

func (s *server) ProcessFile(ctx context.Context, req *pb.ProcessRequest) (*pb.ProcessResponse, error) {
	log.Printf("Получен файл: %s, размер: %d байт", req.Filename, len(req.Content))

	// Здесь должна быть логика обработки файла, поиска точки сбора и генерации изображения
	meetingPoint := "(10, 20)" // Заглушка
	imageData := []byte{}      // Пока пустой массив

	return &pb.ProcessResponse{
		ImageData:    imageData,
		MeetingPoint: meetingPoint,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProcessingServiceServer(grpcServer, &server{})

	log.Println("gRPC сервер запущен на порту 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка при работе сервера: %v", err)
	}
}
