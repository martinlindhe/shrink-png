package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	inFile  = kingpin.Arg("in", "Input file").Required().String()
	outFile = kingpin.Flag("out", "Output file (dont overwrite input)").Short('o').String()
	verbose = kingpin.Flag("verbose", "Verbose").Short('v').Bool()
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	// support -h for --help
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if !exists(*inFile) {
		log.Fatal("File", *inFile, " not found")
	}

	inFileSize := fileSize(*inFile)
	if err := pngcrushCompress(*inFile); err != nil {
		log.Println(err)
	}
	if err := optipngCompress(*inFile); err != nil {
		log.Println(err)
	}
	outFileSize := fileSize(*inFile)

	diffSize := inFileSize - outFileSize
	pctShrunk := 100 - ((float64(outFileSize) / float64(inFileSize)) * 100)
	fmt.Printf("%s: %d -> %d (shrunk by %d bytes, %0.1f%%)",
		*inFile, inFileSize, outFileSize, diffSize, pctShrunk)
	fmt.Println()
}

func pngcrushCompress(file string) error {
	inFileSize := fileSize(file)
	pngcrushFile := findFreeOutFileName(file)
	pngcrushPath, err := lookPath("pngcrush")
	if err != nil {
		return err
	}

	err = runCommand(pngcrushPath, "-s", "-brute", "-rem", "alla", file, pngcrushFile)
	if err != nil {
		return fmt.Errorf("pngcrush: error occured while processing", file, ":", err)
	}
	outSize := fileSize(pngcrushFile)
	diffSize := inFileSize - outSize
	pctShrunk := 100 - ((float64(outSize) / float64(inFileSize)) * 100)
	if *verbose {
		fmt.Printf("pngcrush: %d -> %d (shrunk by %d bytes, %0.1f%%)",
			inFileSize, outSize, diffSize, pctShrunk)
		fmt.Println()
	}

	if diffSize < 0 {
		if *verbose {
			fmt.Println("pngcrush: throwing away non-shrunk file", pngcrushFile)
		}
		os.Remove(pngcrushFile)
		return nil
	}

	err = os.Remove(file)
	if err != nil {
		return err
	}
	return os.Rename(pngcrushFile, file)
}

func optipngCompress(file string) error {
	inFileSize := fileSize(file)
	optipngPath, err := lookPath("optipng")
	if err != nil {
		return err
	}

	optipngFile := findFreeOutFileName(file)
	err = runCommand(optipngPath, "-o7", "-out", optipngFile, file)
	if err != nil {
		return fmt.Errorf("optipng: error occured while processing %s: %v", file, err)
	}

	outSize := fileSize(optipngFile)
	diffSize := inFileSize - outSize
	pctShrunk := 100 - ((float64(outSize) / float64(inFileSize)) * 100)
	if *verbose {
		fmt.Printf("optipng: %d -> %d (shrunk by %d bytes, %0.1f%%)",
			inFileSize, outSize, diffSize, pctShrunk)
		fmt.Println()
	}

	if diffSize < 0 {
		if *verbose {
			fmt.Println("optipng: throwing away non-shrunk file", optipngFile)
		}
		os.Remove(optipngFile)
		return nil
	}

	err = os.Remove(file)
	if err != nil {
		return err
	}
	return os.Rename(optipngFile, file)
}

// exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func baseNameWithoutExt(filename string) string {

	s := filepath.Base(filename)
	n := strings.LastIndexByte(s, '.')
	if n >= 0 {
		return s[:n]
	}
	return s
}

func findFreeOutFileName(file string) string {

	cnt := 0
	res := ""
	ext := path.Ext(file)

	for {
		res = path.Join(filepath.Dir(file), baseNameWithoutExt(file))
		if cnt > 0 {
			res += "-" + fmt.Sprintf("%02d", cnt)
		}
		res += ext
		if !exists(res) {
			break
		}
		cnt++
	}
	return res
}

func runCommand(name string, arg ...string) error {
	//log.Println("EXEC:", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

func fileSize(filepath string) int64 {

	file, err := os.Open(filepath)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return fi.Size()
}

func lookPath(file string) (string, error) {

	if runtime.GOOS == "windows" {
		file += ".exe"
	}
	return exec.LookPath(file)
}
