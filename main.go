package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/windows/registry"
)

func main() {

	checkLog := "check.log"
	_, err := os.Stat(checkLog)
	if os.IsNotExist(err) {
		_, err := os.Create(checkLog)
		if err != nil {
			fmt.Println(err)
		}
	}
	checkLogAllert := "checkAllert.log"
	_, err = os.Stat(checkLogAllert)
	if os.IsNotExist(err) {
		_, err := os.Create(checkLogAllert)
		if err != nil {
			fmt.Println(err)
		}
	}
	var gameName string
	steamPath := regGet(`SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath") // получаем папку стима
	vdfSteam := steamPath + `\steamapps\libraryfolders.vdf`                // файл со списком библиотек стима

	steamLibrary, err := parseSteamLibrary(vdfSteam)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(5 * time.Second)

	for _, path := range steamLibrary {
		manifest, err := filepath.Glob(path + `\appmanifest_*.acf`)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("\n>>>>>>>>>> Библиотека - %s <<<<<<<<<<\n", path)
		for _, file_manifest := range manifest {

			gameName, err = getInfo(file_manifest, "\"name\"")
			if err != nil {
				fmt.Println(err)
			}

			changeUpdate(file_manifest, gameName)

		}
	}

	fmt.Printf("\nПроверка завершена, все игры настроены на обновление с высоким приоритетом. Окно можно закрывать")
	g := ""
	fmt.Scan(&g)
}

func changeUpdate(filename, gameName string) {
	var check bool
	tempFilePath := filename + ".temp"

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer tempFile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "AutoUpdateBehavior") && strings.Contains(line, "0") {
			line = strings.Replace(line, "0", "2", -1)
			check = true
		}
		tempFile.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()
	tempFile.Close()

	if check {
		if err := os.Rename(tempFilePath, filename); err != nil {
			log.Fatal(err)
		}
		fmt.Println(filename, "обновлен. Игра -", gameName)
	} else {
		err := os.Remove(tempFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}

// получаем данные из реестра
func regGet(regFolder, keys string) string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, regFolder, registry.QUERY_VALUE)
	if err != nil {
		log.Printf("Ошибка открытия ветки реестра: %v. %s\n", err, getLine())
	}
	defer key.Close()

	value, _, err := key.GetStringValue(keys)
	if err != nil {
		log.Printf("Ошибка чтения папки стима: %v. %s\n", err, getLine())
	}
	return value
}

// получение строки кода где возникла ошибка
func getLine() string {
	_, _, line, _ := runtime.Caller(1)
	lineErr := fmt.Sprintf("\nСтрока: %d", line)
	return lineErr
}

func parseSteamLibrary(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var data []string
	var currentPath string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "\"path\"") {
			parts := strings.SplitN(line, "\"path\"", 2)
			path := strings.TrimSpace(parts[1])
			path = strings.Trim(path, "\"\t")
			path = strings.ReplaceAll(path, "\\\\", "\\")
			currentPath = path + "\\steamapps"
			data = append(data, currentPath)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ошибка при сканировании файла:", err)
	}
	return data, err
}

func getInfo(fileName, trimString string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var data string = ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, trimString) {
			parts := strings.SplitN(line, trimString, 2)
			stringManifest := strings.TrimSpace(parts[1])
			stringManifest = strings.Trim(stringManifest, "\"")
			data = stringManifest
		}
	}
	return data, err
}
