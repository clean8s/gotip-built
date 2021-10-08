package main

import (
	"log"
	"io"
	"github.com/pierrec/lz4/v4"
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
	"os/signal"
	"syscall"
	"runtime"
	"net/http"
	"fmt"
	"time"
	"bufio"
)

func main() {
	origdir, _ := os.Getwd()
	oldgobin := os.Args[0]
	oldgobin, _ = filepath.Split(oldgobin)

	newgopath := filepath.Join(oldgobin, "tipbuilt")
	if len(os.Args) > 1 && os.Args[1] == "fetch" {
		err := os.Mkdir(newgopath, os.ModeDir)
		if err != nil {
			if os.IsExist(err) {
				fmt.Printf("Warning: there's an already installed version.\nDo you want to overwrite? [y/n]: ")
				S := bufio.NewScanner(os.Stdin)
				S.Scan()
				if strings.TrimSpace(S.Text()) != "y" {
					os.Exit(1)
				}
			} else {
				fmt.Printf("failed creating gotip dir: %s\n", err.Error())
				os.Exit(1)
			}
			os.RemoveAll(newgopath)
			os.Mkdir(newgopath, os.ModeDir)
		}
		extract(newgopath)
		os.Exit(0)
	}

	gtip, err := os.Stat(filepath.Join(newgopath, "gotip"))
	if err != nil || !gtip.IsDir() {
		fmt.Println("Warning: compiled release not downloaded yet, please run 'gotip fetch'!")
		os.Exit(1)
	}

	bins := filepath.Join(newgopath, "gotip", "bin")
	entries, _ := os.ReadDir(bins)
	found := ""
	for _, binEntry := range entries {
		filename := strings.TrimSuffix(binEntry.Name(), filepath.Ext(binEntry.Name()))
		if filename == "go" {
			found = filepath.Join(bins, binEntry.Name())
		}
	}
	if found != "" {
		callgo(found, newgopath, origdir)
	}

}

func callgo(exe string, gopath string, workdir string) {
	g := exec.Command(exe, os.Args[1:]...)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GOROOT", filepath.Join(gopath, "gotip"))
	os.Setenv("GOBIN", filepath.Join(gopath, "..", "bin"))

	g.Stdout = os.Stdout
	g.Stderr = os.Stderr
	g.Stdin = os.Stdin
	g.Dir = workdir
	//g.Env = os.Environ()
	var sigResult os.Signal
	defer func() {
		if g.Process != nil && sigResult != nil {
			g.Process.Signal(sigResult)
		} else if g.Process != nil {
			g.Process.Kill()
		}
	}()
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		g.Run()
		sigc <- nil
	}()

	sigResult = <-sigc
}

var SUPPORTED_OS = map[string]bool {
	"windows": true, "mac": true, "linux": true,
}
var SUPPORTED_ARCH = map[string]bool {
	"amd64": true,
}

func extract(gopath string) {
	thisOs := runtime.GOOS
	if thisOs == "darwin" {
		thisOs = "mac"
	}
	if !SUPPORTED_OS[thisOs] || !SUPPORTED_ARCH[runtime.GOARCH] {
		log.Fatalln("Only Linux, Windows & Mac with amd64 are supported.")
	}

	uri := "https://github-releases.fikisipi.workers.dev/" + thisOs

	archiveReq, _ := http.Get(uri)
	archiveFile := archiveReq.Body

	data := lz4.NewReader(archiveFile)
	tarReader := tar.NewReader(data)

	expectedSize := int(2 * archiveReq.ContentLength)
	currentSize := 0
	lastProgress := -15.0
	tm := time.Now()
	draw := func(prog int) float64{
		dt := time.Now().Sub(tm)
		currentSize += prog
		if currentSize > expectedSize {
			currentSize = expectedSize
		}
		progress := float64(currentSize) / float64(expectedSize)
		{
			dots := "."
			n := currentSize * 10 / expectedSize
			dots += strings.Repeat(".", n)
			dots += strings.Repeat(" ", 10 - n)
			K := fmt.Sprintf("Downloading precompiled bin (%.1f" + "%%" + " in %ds)", progress * 100, int(dt.Seconds()))
			fmt.Printf("%-40s |%s|\r", K, dots)
			lastProgress = progress
		}
		return progress
	}

	go func() {
		for {
			nowFrac := draw(0)
			if nowFrac > 0.8 {
				break
			}
			time.Sleep(time.Second)
		}
	}()
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		draw(int(header.Size))

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}
		header.Name = filepath.Join(gopath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}
	}

}
