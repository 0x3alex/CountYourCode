package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type lines struct {
	linesEachFile map[string]int
	total         int
}

func (lines *lines) readFile(path string,
	c chan int,
	scannerBuffer *int,
	wg sync.WaitGroup) {
	file, err := os.Open(path)
	defer file.Close() //close when done
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer wg.Done() //tell waitgroup that routine is done
	total := 0
	scanner := bufio.NewScanner(file)
	buf := make([]byte, *scannerBuffer)
	scanner.Buffer(buf, *scannerBuffer)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 { //skip if it's just an empty line
			total++
		}
	}
	lines.total += total
	c <- total //send total of the file back through channel
}

func (lines *lines) iterateOverDir(path string,
	excludeDirs []string,
	excludeFileTypes []string,
	scannerBuffer *int,
	wg sync.WaitGroup) {
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		skip := false
		if err != nil {
			log.Fatalf(err.Error())
		}
		if f.IsDir() {
			return nil
		}
		for _, e := range excludeDirs {
			if strings.Contains(path, e) {
				skip = true
			}
		}
		if skip {
			return nil
		}
		extension := strings.Split(f.Name(), ".")
		if len(extension) > 1 {
			for _, e := range excludeFileTypes {
				if e == extension[1] {
					skip = true
					break
				}
			}
			c := make(chan int)
			if !skip {
				wg.Add(1)
				go lines.readFile(path, c, scannerBuffer, wg)
				lines.linesEachFile[path] = <-c
			}
			skip = false

		}
		return nil
	})
}

func (lines *lines) printResult() {
	if len(lines.linesEachFile) == 0 || lines.total == 0 {
		fmt.Println("No files were read!")
		return
	}
	for k, e := range lines.linesEachFile {
		fmt.Printf("File %s contained %d lines\n", k, e)
	}
	fmt.Printf("This project has a total lenght of %d lines\n", lines.total)
}

func main() {

	dirFlag := flag.String(
		"dir",
		"None",
		"Set the directory to iterate over",
	)
	excludeFileFlag := flag.String(
		"excludeFiles",
		"None",
		"Add file extensions to exclude.\n-> Split with ;\n-> no . needed",
	)
	excludeDirsFlag := flag.String(
		"excludeDirs",
		"None",
		"Add directories that shall be excluded.\n-> Split with ;",
	)
	scannerBuffer := flag.Int(
		"scannerBuffer",
		64000,
		"Adjust the size of the scanner's buffer, when reading a file (in lines).",
	)
	flag.Parse()
	if *dirFlag == "None" {
		log.Fatalf("No directory was given!\nRun -help to see all options")
	}
	lines := lines{
		linesEachFile: make(map[string]int),
		total:         0,
	}
	wg := sync.WaitGroup{}
	excludeFileTypes := strings.Split(*excludeFileFlag, ";")
	excludeDirs := strings.Split(*excludeDirsFlag, ";")
	lines.iterateOverDir(*dirFlag, excludeDirs, excludeFileTypes, scannerBuffer, wg)
	wg.Wait() //wait for all routines to finish
	lines.printResult()
}
