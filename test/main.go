package main

import (
	"bufio"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/fogleman/gg"
)

type Point struct {
	X, Y int
}

func readMaze(filename string) [][]rune {
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

	for y := 0; y < height; y++ {
		scanner.Scan()
		maze[y] = []rune(scanner.Text())
	}

	return maze
}

func generateMazeImage(maze [][]rune, outputPath string) error {
	width, height := len(maze[0]), len(maze)
	dc := gg.NewContext(width*10, height*10)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

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

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, dc.Image())
}

func main() {
	inputFile := "maze.txt"
	outputImage := "maze.png"

	maze := readMaze(inputFile)
	err := generateMazeImage(maze, outputImage)
	if err != nil {
		log.Fatalf("Ошибка генерации изображения: %v", err)
	}
	fmt.Println("Карта сохранена в", outputImage)
}
