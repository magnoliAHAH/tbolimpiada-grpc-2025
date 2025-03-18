package main

import (
	"bytes"
	"context"
	"core/imagegen"
	"core/pathfinder"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedProcessingServiceServer
}

func (s *server) ProcessFile(ctx context.Context, req *pb.ProcessRequest) (*pb.ProcessResponse, error) {
	log.Printf("Получен файл: %s, размер: %d байт", req.Filename, len(req.Content))

	// Создание временной директории для хранения файлов
	tmpDir := "/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка при создании директории: %v", err)
	}

	// Сохранение загруженного файла
	filePath := fmt.Sprintf("%s/%s", tmpDir, req.Filename)
	if err := os.WriteFile(filePath, req.Content, 0644); err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	// Читаем лабиринт
	maze, heroes := pathfinder.ReadMaze(filePath)

	// Ищем оптимальную точку встречи
	meetingPoint := pathfinder.FindOptimalMeetingPoint(maze, heroes)

	// Находим пути героев до точки встречи
	var paths [][]imagegen.Point
	for _, hero := range heroes {
		path := pathfinder.FindPath(maze, hero, meetingPoint)
		if path != nil {
			var imgPath []imagegen.Point
			for _, p := range path {
				imgPath = append(imgPath, imagegen.Point{X: p.X, Y: p.Y})
			}
			paths = append(paths, imgPath)
		}
	}

	// Генерация изображения
	imagePath := fmt.Sprintf("%s/maze.png", tmpDir)
	meetingPointImg := imagegen.Point{X: meetingPoint.X, Y: meetingPoint.Y}

	if err := imagegen.GenerateMazeImage(maze, meetingPointImg, paths, imagePath); err != nil {
		return nil, fmt.Errorf("ошибка генерации изображения: %v", err)
	}

	// Формирование имени файла без расширения
	dotIndex := strings.LastIndex(req.Filename, ".")
	if dotIndex != -1 {
		req.Filename = req.Filename[:dotIndex] // Обрезаем расширение
	}

	// Отправка изображения на файловый сервер
	downloadURL, err := uploadImageToServer(imagePath, req.Filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки изображения на файловый сервер: %v", err)
	}

	// Возвращаем gRPC-ответ
	return &pb.ProcessResponse{
		ImageUrl:     downloadURL,
		MeetingPoint: fmt.Sprintf("(%d, %d)", meetingPoint.X, meetingPoint.Y),
	}, nil
}

func uploadImageToServer(imagePath, filename string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения файла: %w", err)
	}
	fileServerURL := fmt.Sprintf("http://fileserver/upload/%s.png", filename)

	req, err := http.NewRequest("PUT", fileServerURL, bytes.NewBuffer(fileData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "image/png")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("сервер вернул ошибку: %d", resp.StatusCode)
	}

	return fmt.Sprintf("http://localhost:8085/images/upload/%s.png", filename), nil
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
