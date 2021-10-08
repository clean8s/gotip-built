package main

import (
	"log"
	"io"
	"github.com/pierrec/lz4/v4"
	"archive/tar"
	"os"
	"github.com/pterm/pterm"
	"path/filepath"
	"strings"
	"os/exec"
	"os/signal"
	"syscall"
	"runtime"
	"net/http"
)

func IsExecAny(mode os.FileMode) bool {
	return mode&0111 != 0
}

func main() {
	origdir, _ := os.Getwd()
	oldgobin := os.Args[0]
	oldgobin, _ = filepath.Split(oldgobin)

	newgopath := filepath.Join(oldgobin, "tipbuilt")
	os.Mkdir(newgopath, os.ModeDir)
	os.Chdir(newgopath)

	gtip, err := os.Stat("gotip")
	if err != nil || !gtip.IsDir() {
		extract()
	}

	bins := filepath.Join("gotip", "bin")
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

	defer func() {
		if g.Process != nil {
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
	<-sigc
}

var SUPPORTED_OS = map[string]bool {
	"windows": true, "mac": true, "linux": true,
}
var SUPPORTED_ARCH = map[string]bool {
	"amd64": true,
}

func extract() {
	thisOs := runtime.GOOS
	if thisOs == "darwin" {
		thisOs = "mac"
	}
	if !SUPPORTED_OS[thisOs] || !SUPPORTED_ARCH[runtime.GOARCH] {
		pterm.Fatal.Println("Only Linux, Windows & Mac with amd64 are supported.")
	}

	uri := "https://github-releases.fikisipi.workers.dev/" + thisOs

	archiveReq, _ := http.Get(uri)
	archiveFile := archiveReq.Body

	data := lz4.NewReader(archiveFile)
	tarReader := tar.NewReader(data)

	expectedSize := int(3 * archiveReq.ContentLength)
	progress, _ := pterm.DefaultProgressbar.WithShowCount(false).
		WithShowPercentage(true).WithShowElapsedTime(true).
		WithTotal(expectedSize).WithTitle("Downloading precompiled binary...").Start()
	progress.Total = expectedSize

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if progress.Current+int(header.Size) <= progress.Total {
			progress.Add(int(header.Size))
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

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
	progress.Add(progress.Total - progress.Current)
	progress.Stop()

}
