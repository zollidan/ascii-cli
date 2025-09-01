package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	filePath = flag.String("file", "", "Путь к изображению (png/jpg/gif/webp)")
	width    = flag.Int("width", 100, "Ширина ASCII в символах")
	color    = flag.Bool("color", true, "Цветной вывод в терминал (ANSI 24-bit)")
	bw       = flag.Bool("bw", false, "Черно-белый вывод (# и пробел)")
	markdown = flag.Bool("markdown", false, "Вывод в формате Markdown (без цветов)")
	outFile  = flag.String("out", "", "Сохранить вывод в файл (например, README.md)")
	fps      = flag.Int("fps", 0, "FPS для GIF (0 = использовать задержки из GIF)")
	loop     = flag.Bool("loop", true, "Крутить GIF бесконечно в терминале")
)

const grayRamp = " .:-=+*#%@"

func clamp[T ~int | ~float64](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func toASCII(img image.Image, targetWidth int, useColor bool, useBW bool) []string {
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w == 0 || h == 0 {
		return []string{"(empty image)"}
	}

	ratio := float64(w) / float64(h)
	targetHeight := int(float64(targetWidth) / ratio / 2)
	targetHeight = clamp(targetHeight, 1, 2000)

	xStep := float64(w) / float64(targetWidth)
	yStep := float64(h) / float64(targetHeight)

	lines := make([]string, 0, targetHeight)
	var sb strings.Builder

	for y := 0; y < targetHeight; y++ {
		sb.Reset()
		for x := 0; x < targetWidth; x++ {
			srcX := int(float64(x) * xStep)
			srcY := int(float64(y) * yStep)
			r, g, b, _ := img.At(b.Min.X+srcX, b.Min.Y+srcY).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			gray := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
			var ch byte
			if useBW {
				if gray > 127 {
					ch = ' '
				} else {
					ch = '#'
				}
			} else {
				idx := int((gray / 255.0) * float64(len(grayRamp)-1))
				ch = grayRamp[idx]
			}

			if useColor {
				// 24-bit цвет в терминале
				sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", r8, g8, b8, ch))
			} else {
				sb.WriteByte(ch)
			}
		}
		if useColor {
			sb.WriteString("\033[0m")
		}
		lines = append(lines, sb.String())
	}
	return lines
}

func printFrame(lines []string, useAnsi bool) {
	if useAnsi {
		// Переместиться в левый верхний угол и очистить экран (без мерцания)
		fmt.Print("\033[H")
	}
	for _, line := range lines {
		fmt.Println(line)
	}
}

func writeOutput(lines []string, asMarkdown bool, path string) error {
	var b strings.Builder
	if asMarkdown {
		b.WriteString("```\n")
	}
	for _, l := range lines {
		// Без ANSI-цветов в markdown
		if asMarkdown {
			// удалим возможные ANSI коды на всякий случай
			l = stripANSI(l)
		}
		b.WriteString(l)
		b.WriteByte('\n')
	}
	if asMarkdown {
		b.WriteString("```\n")
	}
	if path == "" {
		_, err := os.Stdout.WriteString(b.String())
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func stripANSI(s string) string {
	// Простейшая зачистка ESC-последовательностей
	var out strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x1b { // ESC
			inEsc = true
			continue
		}
		if inEsc {
			// завершаем CSI на 'm' или любую букву
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				inEsc = false
			}
			continue
		}
		out.WriteByte(c)
	}
	return out.String()
}

func main() {
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Укажи путь к файлу: -file path/to/image.(png|jpg|gif)")
		os.Exit(1)
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		os.Exit(1)
	}

	// Обработка SIGINT, чтобы вернуть курсор
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer func() {
		if *color || !*markdown {
			fmt.Print("\033[0m\033[?25h") // reset + показать курсор
		}
	}()
	go func() {
		<-c
		fmt.Print("\033[0m\033[?25h\n")
		os.Exit(0)
	}()

	ext := strings.ToLower(filepath.Ext(*filePath))

	// Попробуем как GIF сначала, если расширение .gif — обязательно
	isGIF := ext == ".gif"
	var g *gif.GIF
	if isGIF {
		g, err = gif.DecodeAll(bytes.NewReader(data))
		if err != nil {
			fmt.Println("Ошибка декодирования GIF:", err)
			os.Exit(1)
		}
	}

	// Режим Markdown всегда без цвета
	useColor := *color && !*markdown && !*bw

	if isGIF && len(g.Image) > 1 {
		// Анимированный GIF
		framesASCII := make([][]string, len(g.Image))
		for i, palImg := range g.Image {
			framesASCII[i] = toASCII(palImg, *width, useColor, *bw)
		}

		if *markdown || *outFile != "" {
			// GitHub не анимирует ASCII; выводим первый кадр как статичный блок
			first := framesASCII[0]
			if *outFile != "" {
				if err := writeOutput(first, true, *outFile); err != nil {
					fmt.Println("Ошибка записи:", err)
					os.Exit(1)
				}
				fmt.Println("Сохранено:", *outFile, "(статичный кадр для Markdown)")
				return
			}
			_ = writeOutput(first, true, "")
			return
		}

		// Терминальная анимация
		// Скрыть курсор и очистить экран
		fmt.Print("\033[?25l\033[2J\033[H")

		var frameDurations []time.Duration
		for _, d := range g.Delay {
			// Delay указывается в сотых долях секунды
			ms := time.Duration(d*10) * time.Millisecond
			if ms <= 0 {
				ms = 50 * time.Millisecond
			}
			frameDurations = append(frameDurations, ms)
		}

		tick := func(i int) time.Duration {
			if *fps > 0 {
				return time.Second / time.Duration(*fps)
			}
			return frameDurations[i%len(frameDurations)]
		}

		i := 0
		for {
			printFrame(framesASCII[i%len(framesASCII)], true)
			time.Sleep(tick(i))
			i++
			if !*loop && i >= len(framesASCII) {
				break
			}
		}
		return
	}

	// Статичное изображение (или одно кадро́вый GIF)
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Println("Ошибка декодирования изображения:", err)
		os.Exit(1)
	}
	lines := toASCII(img, *width, useColor, *bw)

	if *outFile != "" {
		if err := writeOutput(lines, *markdown, *outFile); err != nil {
			fmt.Println("Ошибка записи:", err)
			os.Exit(1)
		}
		fmt.Println("Сохранено:", *outFile)
		return
	}

	if *markdown {
		_ = writeOutput(lines, true, "")
		return
	}

	fmt.Print("\033[2J\033[H")
	printFrame(lines, false)
}
