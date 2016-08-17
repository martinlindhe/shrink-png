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
	dry     = kingpin.Flag("dry", "Dry-run").Bool()
	verbose = kingpin.Flag("verbose", "Verbose").Short('v').Bool()
)

func main() {

	var err error

	// support -h for --help
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if !exists(*inFile) {
		log.Fatal("File", *inFile, " not found")
	}

	inFileSize := fileSize(*inFile)
	tempFile := findFreeOutFileName(*inFile)

	pngcrushPath, err := lookPath("pngcrush")
	if err != nil {
		log.Fatal(err)
	}
	optipngPath, err := lookPath("optipng")
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(pngcrushPath, "-s", "-brute", "-rem", "alla", *inFile, tempFile)
	if err != nil {
		log.Fatal("Error occured with pngcrush while processing", *inFile, ":", err)
	}
	outSizePngcrush := fileSize(tempFile)
	diffSizePngcrush := inFileSize - outSizePngcrush
	if *verbose {
		fmt.Println("pngcrush:", inFileSize, "->", outSizePngcrush, "(saved", diffSizePngcrush, "bytes)")
	}

	err = runCommand(optipngPath, "-o7", tempFile)
	if err != nil {
		log.Fatal("Error occured with optipng while processing", *inFile, ":", err)
	}

	outFileSize := fileSize(tempFile)
	diffSizeOptipng := outSizePngcrush - outFileSize
	if *verbose {
		fmt.Println("optipng :", outSizePngcrush, "->", outFileSize, "(saved", diffSizeOptipng, "bytes)")
	}

	diffSize := inFileSize - outFileSize
	fmt.Println(*inFile+":", inFileSize, "->", outFileSize, "(saved", diffSize, "bytes)")

	if !*dry {
		if *outFile == "" {
			// overwrite input file
			err = os.Remove(*inFile)
			if err != nil {
				log.Fatal(err)
			}

			err = os.Rename(tempFile, *inFile)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// write to new file
			fmt.Println("Writing to", *outFile)
			err = os.Rename(tempFile, *outFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		// if dry-run, remove temp file
		err = os.Remove(tempFile)
		if err != nil {
			log.Fatal(err)
		}
	}

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
