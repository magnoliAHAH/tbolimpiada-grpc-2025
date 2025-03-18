package main

import (
	"bufio"
	"container/heap"
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"os"

	pb "github.com/magnoliAHAH/protos-tbolimpiada/gen"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedProcessingServiceServer
}
type Point struct {
	X, Y int
}

type Item struct {
	point Point
	cost  float64
	path  []Point
	index int
}

var terrainCost = map[rune]float64{
	'R': 0.5,         // Дорога
	'G': 1,           // Поле
	'S': 5,           // Болото
	'H': 4,           // Холмы
	'F': 3,           // Лес
	'W': math.Inf(1), // Вода (непроходимо)
	'M': math.Inf(1), // Горы (непроходимо)
}

func (s *server) ProcessFile(ctx context.Context, req *pb.ProcessRequest) (*pb.ProcessResponse, error) {
	log.Printf("Получен файл: %s, размер: %d байт", req.Filename, len(req.Content))

	// 1. Проверяем и создаем директорию /tmp/, если её нет
	tmpDir := "/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка при создании директории: %v", err)
	}

	// 2. Сохраняем файл на диск
	filePath := tmpDir + "/" + req.Filename
	err := os.WriteFile(filePath, req.Content, 0644)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %v", err)
	}
	log.Printf("Файл сохранён: %s", filePath)
	maze, heroes := readMaze(filePath)

	// 3. Здесь можно выполнить обработку файла (например, анализ содержимого)
	meetingPoint := findOptimalMeetingPoint(maze, heroes)
	// 4. Можно создать картинку, например, как превью или анализ данных
	imageData := []byte{} // Пока пусто

	// 5. Возвращаем ответ
	return &pb.ProcessResponse{
		ImageData:    imageData,
		MeetingPoint: fmt.Sprintf("(%d, %d)", meetingPoint.X, meetingPoint.Y),
	}, nil
}

// Читает лабиринт, [][]rune — сам лабиринт, каждый символ представляет тип местности, []Point — координаты героев
func readMaze(filename string) ([][]rune, []Point) {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Ошибка при открытии файла: %v", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var width, height int
	scanner.Scan()
	fmt.Sscanf(scanner.Text(), "%d", &width)

	scanner.Scan()
	fmt.Sscanf(scanner.Text(), "%d", &height)

	maze := make([][]rune, height)
	var heroes []Point

	for y := 0; y < height; y++ {
		scanner.Scan()
		line := scanner.Text()
		maze[y] = []rune(line)
		for x, cell := range line {
			if cell >= '1' && cell <= '9' {
				heroes = append(heroes, Point{x, y})
			}
		}
	}

	return maze, heroes
}

func findOptimalMeetingPoint(maze [][]rune, heroes []Point) Point {
	minCost := math.Inf(1)
	bestPoint := Point{0, 0}

	for y := 0; y < len(maze); y++ {
		for x := 0; x < len(maze[0]); x++ {
			if terrainCost[maze[y][x]] == math.Inf(1) {
				continue
			}

			totalCost := 0.0
			for _, hero := range heroes {
				path := findPath(maze, hero, Point{x, y})
				if path == nil {
					totalCost = math.Inf(1)
					break
				}
				for _, p := range path {
					totalCost += terrainCost[maze[p.Y][p.X]]
				}
			}

			if totalCost < minCost {
				minCost = totalCost
				bestPoint = Point{x, y}
			}
		}
	}

	return bestPoint
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].cost < pq[j].cost }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index, pq[j].index = i, j }
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func findPath(maze [][]rune, start, end Point) []Point {
	dx := []int{0, 0, -1, 1}
	dy := []int{-1, 1, 0, 0}

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Item{start, 0, []Point{start}, 0})

	visited := make(map[Point]bool)

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*Item)

		if current.point == end {
			return current.path
		}

		if visited[current.point] {
			continue
		}
		visited[current.point] = true

		for i := 0; i < 4; i++ {
			nx, ny := current.point.X+dx[i], current.point.Y+dy[i]
			neighbor := Point{nx, ny}

			if isValid(neighbor, maze) && !visited[neighbor] {
				newCost := current.cost + terrainCost[maze[ny][nx]]
				newPath := append([]Point(nil), current.path...)
				newPath = append(newPath, neighbor)
				heap.Push(pq, &Item{neighbor, newCost, newPath, 0})
			}
		}
	}

	return nil
}

func isValid(p Point, maze [][]rune) bool {
	return p.Y >= 0 && p.Y < len(maze) && p.X >= 0 && p.X < len(maze[0]) && terrainCost[maze[p.Y][p.X]] != math.Inf(1)
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
