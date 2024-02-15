package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Projector struct {
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Link         string `json:"link"`
}

const magicNumber uint32 = 0xFA123456
const url string = "https://raw.githubusercontent.com/WiLuX-Source/SWFTOOL/main/projectors.json"
const version string = "2.0.0"

var fileBytes = map[string][]byte{
	"exe":   {0x4D, 0x5A, 0x90},
	"swf":   {0x43, 0x57, 0x53},
	"linux": {0x7F, 0x45, 0x4C},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You need to provide a command.")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "download":
		showLogo()
		goos := runtime.GOOS
		bits := strconv.IntSize
		fmt.Println(goos, "-", bits, "bit", "detected!")

		projectors, err := parseProjectors(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		projector := projectorFilter(projectors, goos, bits)
		err = downloadFile(projector.Link, "Projector.exe")
		fmt.Println("Projector downloaded successfully.")
		check(err)
	case "merge":
		showLogo()
		filePath1, filePath2 := os.Args[2], os.Args[3]
		name1, ext1 := nameext(filePath1)
		name2, ext2 := nameext(filePath2)

		if validate(filePath1, ext1) && validate(filePath2, ext2) {
			var mergedBuffer []byte
			var err error

			if ext1 == ".swf" {
				mergedBuffer, err = mergeMovie(filePath2, filePath1)
				saveMovie(name1+".exe", mergedBuffer)
			} else if ext2 == ".swf" {
				mergedBuffer, err = mergeMovie(filePath1, filePath2)
				saveMovie(name2+".exe", mergedBuffer)
			}

			check(err)
			fmt.Println("Movies merged successfully.")
		} else {
			fmt.Println("Invalid file(s) provided.")
		}
	case "extract":
		showLogo()
		filePath := os.Args[2]
		name, ext := nameext(filePath)

		if validate(filePath, ext) {
			movie, err := extractMovie(filePath)
			check(err)
			err = saveMovie(name+".swf", movie)
			check(err)
			fmt.Println("Movie extracted successfully.")
		}
	case "help":
		showLogo()
		fmt.Printf("SWFTOOL v%s\n", version)
		fmt.Println("Available commands:")
		fmt.Println("download - Downloads the latest projector for your OS")
		fmt.Println("merge - Merges a projector with a movie")
		fmt.Println("extract - Extracts a movie from a projector")
		fmt.Println("help - Shows this help message")
	default:
		fmt.Println("Command does not exist!")
		os.Exit(1)
	}
}

func downloadFile(url string, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("Downloading...")

	startTime := time.Now()
	var lastPrintTime time.Time
	var downloadedBytes int64

	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			file.Write(buf[:n])
			downloadedBytes += int64(n)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if time.Since(lastPrintTime) >= time.Second {
			elapsedTime := time.Since(startTime).Seconds()
			speed := float64(downloadedBytes) / elapsedTime
			percentage := float64(downloadedBytes) / float64(resp.ContentLength) * 100

			fmt.Printf("\r%.2f%% - %.2f KB/s", percentage, speed/1024)
			lastPrintTime = time.Now()
		}
	}

	fmt.Println()
	return nil
}

func parseProjectors(url string) ([]Projector, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching JSON data: %v", err)
	}
	defer response.Body.Close()

	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON data: %v", err)
	}

	var projectors []Projector

	err = json.Unmarshal(jsonData, &projectors)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return projectors, nil
}

func projectorFilter(projectors []Projector, goos string, bits int) Projector {
	for _, p := range projectors {
		if p.Architecture == strconv.Itoa(bits) && p.Name == (goos+"-latest") {
			return p
		}
	}

	return Projector{}
}

func mergeMovie(projector, movie string) ([]byte, error) {
	buffer1, err := os.ReadFile(projector)
	if err != nil {
		return nil, err
	}

	buffer2, err := os.ReadFile(movie)
	if err != nil {
		return nil, err
	}

	// Calculate the total length of the merged buffer
	totalLength := len(buffer1) + len(buffer2) + 8

	// Create the merged buffer with pre-allocated capacity
	mergedBuffer := make([]byte, 0, totalLength)

	// Append buffer1 to the merged buffer
	mergedBuffer = append(mergedBuffer, buffer1...)

	// Append buffer2 to the merged buffer
	mergedBuffer = append(mergedBuffer, buffer2...)

	// Append the magic number
	magicBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(magicBytes, magicNumber)
	mergedBuffer = append(mergedBuffer, magicBytes...)

	// Append the length of buffer2 as little-endian 4-byte value
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, uint32(len(buffer2)))
	mergedBuffer = append(mergedBuffer, lengthBytes...)

	return mergedBuffer, nil
}

func extractMovie(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()

	// Read the last 8 bytes
	buffer := make([]byte, 8)
	_, err = file.ReadAt(buffer, fileSize-8)
	if err != nil {
		return nil, err
	}

	// Check if the last 8 bytes start with the magic number
	if binary.LittleEndian.Uint32(buffer[:4]) != magicNumber {
		return nil, fmt.Errorf("magic number not found, aborting")
	}

	// Get the length of the movie from the last 4 bytes
	movieSize := int64(binary.LittleEndian.Uint32(buffer[4:]))

	// Read the movie data
	movieBuffer := make([]byte, movieSize)
	_, err = file.ReadAt(movieBuffer, fileSize-8-movieSize)
	if err != nil {
		return nil, err
	}

	return movieBuffer, nil
}

func saveMovie(name string, movieData []byte) error {
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	outputPath := filepath.Join(directory, name)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.Write(movieData)
	if err != nil {
		return err
	}

	return nil
}

func validate(filePath string, ext string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 3)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	switch ext {
	case ".exe":
		return compareBytes(buffer, fileBytes["exe"])
	case ".swf":
		return compareBytes(buffer, fileBytes["swf"])
	case "":
		return compareBytes(buffer, fileBytes["linux"])
	default:
		return false
	}
}

func nameext(filePath string) (string, string) {
	filename := strings.Split(filepath.Base(filePath), ".")[0]
	extension := filepath.Ext(filePath)
	return filename, extension
}

func compareBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func showLogo() {
	fmt.Println(" ________  ___       __   ________     __________  ________  ________ ___")
	fmt.Println(`|\   ____\|\  \     |\  \|\  _____\   |\___   ___\\   __  \|\   __  \|\  \`)
	fmt.Println(`\ \  \___|\ \  \    \ \  \ \  \__/     \|___\  \_\ \  \|\  \ \  \|\  \ \  \`)
	fmt.Println(` \ \_____  \ \  \  __\ \  \ \   __\        \ \  \ \ \  \\\  \ \  \\\  \ \  \`)
	fmt.Println(`  \|____|\  \ \  \|\__\_\  \ \  \_|         \ \  \ \ \  \\\  \ \  \\\  \ \  \____`)
	fmt.Println(`    ____\_\  \ \____________\ \__\           \ \__\ \ \_______\ \_______\ \_______\`)
	fmt.Println(`   |\_________\|____________|\|__|            \|__|  \|_______|\|_______|\|_______|`)
	fmt.Println(`   \|_________|`)
}

func check(e error) {
	if e != nil {
		_ = fmt.Errorf("error: %v", e)
	}
}
