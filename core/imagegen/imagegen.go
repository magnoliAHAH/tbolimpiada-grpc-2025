package imagegen

import (
	"image/color"
	"image/png"
	"os"

	gg "github.com/fogleman/gg"
)

type Point struct {
	X, Y int
}

func GenerateMazeImage(maze [][]rune, meetingPoint Point, paths [][]Point, outputPath string) error {
	width, height := len(maze[0]), len(maze)
	dc := gg.NewContext(width*10, height*10)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Рисуем карту
	for y, row := range maze {
		for x, cell := range row {
			var col color.Color
			switch cell {
			case 'R':
				col = color.RGBA{255, 255, 0, 255} // Дорога
			case 'G':
				col = color.RGBA{0, 255, 0, 255} // Поле
			case 'S':
				col = color.RGBA{139, 69, 19, 255} // Болото
			case 'H':
				col = color.RGBA{34, 139, 34, 255} // Холмы
			case 'F':
				col = color.RGBA{0, 100, 0, 255} // Лес
			case 'M':
				col = color.RGBA{128, 128, 128, 255} // Горы
			case 'W':
				col = color.RGBA{0, 255, 255, 255} // Вода
			default:
				col = color.White
			}
			dc.SetColor(col)
			dc.DrawRectangle(float64(x*10), float64(y*10), 10, 10)
			dc.Fill()
		}
	}

	// Рисуем пути
	dc.SetRGB(1, 0, 0) // Красный цвет для путей
	dc.SetLineWidth(2)
	for _, path := range paths {
		for i := 0; i < len(path)-1; i++ {
			p1, p2 := path[i], path[i+1]
			dc.DrawLine(float64(p1.X*10+5), float64(p1.Y*10+5), float64(p2.X*10+5), float64(p2.Y*10+5))
			dc.Stroke()
		}
	}

	// Рисуем точку встречи
	dc.SetRGB(0, 0, 1) // Синий цвет для точки встречи
	dc.DrawCircle(float64(meetingPoint.X*10+5), float64(meetingPoint.Y*10+5), 5)
	dc.Fill()

	// Сохраняем изображение
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, dc.Image())
}

/*
func main() {
	inputFile := "maze.txt"
	outputImage := "maze.png"

	maze := readMaze(inputFile)
	meetingPoint := Point{5, 5} // Пример точки встречи
	paths := [][]Point{
		{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},
		{{7, 2}, {6, 3}, {5, 4}, {5, 5}},
	} // Пример путей

	err := generateMazeImage(maze, meetingPoint, paths, outputImage)
	if err != nil {
		log.Fatalf("Ошибка генерации изображения: %v", err)
	}
	fmt.Println("Карта сохранена в", outputImage)
}
*/
