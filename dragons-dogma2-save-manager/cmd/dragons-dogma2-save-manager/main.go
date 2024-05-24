package main

import (
	"bufio"
	"dragons-dogma2-save-manager/app/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func InitSaver() {
	conf := config.NewConfig()

	dir := askInput("Enter directory path for store DD2 save-files: ")
	if dir == "" {
		fmt.Println("Please enter directory path")
		os.Exit(1)
	}
	_, err := os.Stat(dir)
	if err != nil {
		fmt.Println("Directory not found")
		os.Exit(1)
	}

	charName := askInput("Enter character name or live empty: ")
	fullDestPath := dir
	if len(charName) > 0 {
		conf.Character = charName
		fullDestPath = filepath.Join(dir, charName)
	}
	conf.SavesDir = fullDestPath

	// Find Steam directory
	steamDir, err := findSubdirectory("C:\\Program Files (x86)\\Steam\\userdata", "2054970\\remote\\win64_save")
	if err != nil {
		fmt.Println("Steam directory not found")
		os.Exit(1)
	}
	if len(steamDir) > 1 {
		fmt.Println("Multiple Steam directories found")
		for i, dir := range steamDir {
			fmt.Printf("%d) %s\n", i, dir)
		}

		steamDirNum := askInput("Enter number of Steam directory: ")
		num := 0
		fmt.Sscanf(steamDirNum, "%d", &num)
		if num < 0 || num >= len(steamDir) {
			fmt.Println("Invalid number")
			os.Exit(1)
		}
		conf.SteamDir = steamDir[num]
	} else {
		conf.SteamDir = steamDir[0]
	}
	config.SaveConfig(conf)
}

func getDtFolderName() string {
	currentTime := time.Now()
	return currentTime.Format("20060102-150405")
}

func getDD2SavePath(dstDir string, message string) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return filepath.Join(dstDir, getDtFolderName()+" "+message)
	}
	return filepath.Join(dstDir, getDtFolderName())
}

func askInput(question string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// func makeDir(path string) error {
// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		err := os.MkdirAll(path, 0755)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func findSubdirectory(root, subtree string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasSuffix(path, subtree) {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirs, nil
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			dstPath := filepath.Join(dst, relPath)
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return err
			}
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()
			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return err
			}
		}
		return nil
	})
}

func saveSaves(srcDir string, dstDir string, message string) {
	dstDir = getDD2SavePath(dstDir, message)
	if err := copyDir(srcDir, dstDir); err != nil {
		fmt.Println("Error copying save files")
		os.Exit(1)
	}
}

func loadSaves(srcDir string, dstDir string) {
	if err := copyDir(srcDir, dstDir); err != nil {
		fmt.Println("Error copying save files")
		os.Exit(1)
	}
}

func listSaves(srcDir string) {
	dirs, err := os.ReadDir(srcDir)
	if err != nil {
		fmt.Println("Error reading directory")
		os.Exit(1)
	}
	for _, file := range dirs {
		if file.IsDir() {
			fmt.Println(file.Name())
		}
	}

}

func autoSave(srcDir string, dstDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating file watcher")
		os.Exit(1)
	}
	defer watcher.Close()

	err = watcher.Add(srcDir)
	if err != nil {
		fmt.Println("Error adding directory to watcher")
		os.Exit(1)
	}

	var lastEventTime time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				currentTime := time.Now()
				if currentTime.Sub(lastEventTime) > 5*time.Second {
					fileInfo, err := os.Stat(event.Name)
					if err != nil {
						fmt.Println("Error getting file info:", err)
						continue
					}
					dirNew := filepath.Join(dstDir, fileInfo.ModTime().Format("20060102-150405"))

					if err := copyDir(srcDir, dirNew); err != nil {
						fmt.Println("Error copying save files")
						os.Exit(1)
					}
					lastEventTime = currentTime
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error:", err)
		}
	}
}

func showHelp() {
	fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Println("This program save and restore save files for the Dragon's Dogma 2")
	fmt.Println("Options:")
	fmt.Println("  --help - Show this help message")
	fmt.Println("  --save [\"message\"] - Save current save files")
	fmt.Println("  --load [src dir] - Load save files from src dir")
	fmt.Println("  --list - List all save files")
	fmt.Println("  --autosave - Save files automatically")

}
func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	if !config.IsExist() {
		InitSaver()
	}
	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--help":
		showHelp()
	case "--save":
		if len(os.Args) < 3 {
			saveSaves(conf.SteamDir, conf.SavesDir, "")
		} else {
			message := strings.TrimSpace(strings.Join(os.Args[2:], " "))
			saveSaves(conf.SteamDir, conf.SavesDir, message)
		}
	case "--load":
		if len(os.Args) < 3 {
			fmt.Println("Please enter source directory")
			os.Exit(1)
		}
		srcDir := os.Args[2]
		if len(srcDir) > 3 && srcDir[1:2] != ":\\" {
			srcDir = filepath.Join(conf.SavesDir, srcDir)
		}
		loadSaves(srcDir, conf.SteamDir)
	case "--list":
		listSaves(conf.SavesDir)
	case "--autosave":
		autoSave(conf.SteamDir, conf.SavesDir)
	}

}
