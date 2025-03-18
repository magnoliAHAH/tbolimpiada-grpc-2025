package main

import (
	"context"
	"core/imagegen"
	"core/pathfinder"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedProcessingServiceServer
}

func (s *server) ProcessFile(ctx context.Context, req *pb.ProcessRequest) (*pb.ProcessResponse, error) {
	log.Printf("Получен файл: %s, размер: %d байт", req.Filename, len(req.Content))

	// Сохранение файла
	tmpDir := "/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка при создании директории: %v", err)
	}
	filePath := tmpDir + "/" + req.Filename
	err := os.WriteFile(filePath, req.Content, 0644)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	// Читаем лабиринт
	maze, heroes := pathfinder.ReadMaze(filePath)

	// Ищем оптимальную точку встречи
	meetingPoint := pathfinder.FindOptimalMeetingPoint(maze, heroes)

	// Находим пути героев до точки встречи
	var paths [][]pathfinder.Point
	for _, hero := range heroes {
		path := pathfinder.FindPath(maze, hero, meetingPoint)
		if path != nil {
			paths = append(paths, path)
		}
	}

	// Генерация изображения
	imagePath := tmpDir + "/maze.png"
	meetingPointImg := imagegen.Point{X: meetingPoint.X, Y: meetingPoint.Y}
	var pathsImg [][]imagegen.Point
	for _, path := range paths {
		var imgPath []imagegen.Point
		for _, p := range path {
			imgPath = append(imgPath, imagegen.Point{X: p.X, Y: p.Y})
		}
		pathsImg = append(pathsImg, imgPath)
	}

	err = imagegen.GenerateMazeImage(maze, meetingPointImg, pathsImg, imagePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации изображения: %v", err)
	}

	// Читаем изображение в []byte
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения изображения: %v", err)
	}

	// Возвращаем gRPC-ответ
	return &pb.ProcessResponse{
		ImageData:    imageData,
		MeetingPoint: fmt.Sprintf("(%d, %d)", meetingPoint.X, meetingPoint.Y),
	}, nil
}

// Читает лабиринт, [][]rune — сам лабиринт, каждый символ представляет тип местности, []Point — координаты героев

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
