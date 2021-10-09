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
	"fmt"
	"time"
	"hash"
	"crypto/sha256"
	"net/http"
)

const CLI_VERSION = "4.0"

func printCliVersion() {
	fmt.Printf("cli version %s env %s\n", CLI_VERSION, newPaths(getCurrentGoBin()))
}

type Paths struct {
	GoPath string
	GoRoot string
	GoBin  string
}

func (p Paths) JoinInPath(paths ...string) string {
	return filepath.Join(append([]string{p.GoPath}, paths...)...)
}

func (paths Paths) String() string {
	return fmt.Sprintf("GOPATH: %#v GOROOT: %#v GOBIN: %#v", paths.GoPath, paths.GoRoot, paths.GoBin)
}

func getCurrentGoBin() string {
	argvPath := os.Args[0]
	// Unix tends to give an incomplete path for argv[0]
	if !filepath.IsAbs(argvPath) {
		// you have to instead read the symlink of self
		// to get the full argv[0] (i.e. the path of the current bin)
		procFullPath, err := os.Readlink("/proc/self/exe")
		if err == nil {
			argvPath = filepath.Dir(procFullPath)
		}
	} else {
		// split into (dir)/(file)
		argvPath, _ = filepath.Split(argvPath)
	}
	gobin := argvPath
	return gobin
}

func newPaths(standardGoBin string) Paths {
	newGoPath := filepath.Join(standardGoBin, "gotip-built")
	newGoRoot := filepath.Join(newGoPath, "gotip")
	newGoBin := filepath.Join(newGoRoot, "bin")
	return Paths{
		GoPath: newGoPath,
		GoRoot: newGoRoot,
		GoBin:  newGoBin,
	}
}

func main() {
	standardGobin := getCurrentGoBin()
	newP := newPaths(standardGobin)

	if len(os.Args) > 1 && os.Args[1] == "download" {
		fmt.Printf("Expected paths: %s\n", newP)
		err := os.Mkdir(newP.GoPath, os.ModeDir)
		if err != nil {
			if os.IsExist(err) {
				err := os.RemoveAll(newP.GoPath)
				if err != nil {
					fmt.Printf("Tried to recreate %#v but got an error: %s\n", newP.GoPath, err.Error())
					os.Exit(1)
				}
			} else {
				fmt.Printf("failed creating gotip dir: %s\n", err.Error())
				os.Exit(1)
			}
			err = os.Mkdir(newP.GoPath, os.ModeDir)
			if err != nil {
				fmt.Printf("Tried to create %#v but got an error: %s\n", newP.GoPath, err.Error())
				os.Exit(1)
			}
		}
		extract(newP.GoPath, newP)
		os.Exit(0)
	}

	gtip, err := os.Stat(newP.GoBin)
	if err != nil || !gtip.IsDir() {
		fmt.Println("Compiled release not downloaded yet, please run 'gotip download'")
		os.Exit(1)
	}

	_, err = os.ReadFile(filepath.Join(newP.GoPath, ".tipsuccess"))
	if err != nil {
		fmt.Println("Installation looks corrupted, try 'gotip download' again")
		os.Exit(1)
	}

	callgo(newP)
}

func callgo(paths Paths) {
	var err error
	exe := paths.JoinInPath("gotip", "bin", "go")
	if len(os.Args) > 1 && os.Args[1] == "cli-version" {
		printCliVersion()
		os.Args[1] = "version"
	}

	g := exec.Command(exe, os.Args[1:]...)
	envErr := os.Setenv("GOPATH", paths.GoPath)
	envErr = os.Setenv("GOROOT", paths.GoRoot)
	envErr = os.Setenv("GOBIN", paths.GoBin)

	if envErr != nil {
		fmt.Printf("Failed setting env: %s\n", envErr.Error())
		os.Exit(1)
	}

	g.Stdout = os.Stdout
	g.Stderr = os.Stderr
	g.Stdin = os.Stdin
	//g.Env = os.Environ()
	var sigResult os.Signal
	defer func() {
		if err != nil {
			if newErr, ok := err.(*exec.ExitError); ok {
				os.Exit(newErr.ExitCode())
			} else {
				fmt.Printf("go exec err: %s\n", err)
			}
		}
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
		err = g.Run()
		sigc <- nil
	}()

	sigResult = <-sigc
}

var SUPPORTED_OS = map[string]bool{
	"windows": true, "mac": true, "linux": true,
}
var SUPPORTED_ARCH = map[string]bool{
	"amd64": true,
}

type hashInterceptReader struct {
	sourceReader io.Reader
	sha256        hash.Hash
	readByteCount int
}

func (hasher *hashInterceptReader) Read(p []byte) (n int, err error) {
	if hasher.sha256 == nil {
		hasher.sha256 = sha256.New()
	}
	n, err = hasher.sourceReader.Read(p)
	hasher.readByteCount += n
	hasher.sha256.Write(p[:n])
	return
}

func extract(gopath string, newP Paths) {
	thisOs := runtime.GOOS
	if thisOs == "darwin" {
		thisOs = "mac"
	}
	if !SUPPORTED_OS[thisOs] || !SUPPORTED_ARCH[runtime.GOARCH] {
		log.Fatalln("Only Linux, Windows & Mac with amd64 are supported.")
	}

	// GitHub lets you do GET .json of the releases, but the
	// assets/artifacts in releases cannot be downloaded without a browser.
	// (most likely due to abuse prevention).

	// scripts that need an artifact have another option:
	// send an auth token and you can download anything.
	// this worker periodically does download artifacts with a personal token.

	uri := "https://github-releases.fikisipi.workers.dev/" + thisOs
	archiveReq, err := http.Get(uri)
	if err != nil {
		log.Fatalln(err)
	}
	archiveFile := archiveReq.Body
	path, _ := filepath.Split(archiveReq.Request.URL.Path)
	fileName := filepath.Base(path)
	archMegabyteSize := float64(archiveReq.ContentLength) / 1e6
	fmt.Printf("Last GitHub precompiled release: %s size: %.1fMB\n", fileName, archMegabyteSize)
	hasher := &hashInterceptReader{sourceReader: archiveFile}

	data := lz4.NewReader(hasher)
	tarReader := tar.NewReader(data)

	estimatedSize := int(2 * archiveReq.ContentLength)
	currentSize := 0
	lastProgress := 0.0
	tm := time.Now()
	drawProgress := func(prog int) float64 {
		dt := time.Now().Sub(tm)
		currentSize += prog
		if currentSize > estimatedSize {
			currentSize = estimatedSize
		}
		progress := float64(currentSize) / float64(estimatedSize)
		{
			dots := "."
			n := currentSize * 10 / estimatedSize
			dots += strings.Repeat(".", n)
			dots += strings.Repeat(" ", 10-n)
			K := fmt.Sprintf("Downloading and installing (%.1f"+"%%"+" in %ds)", progress*100, int(dt.Seconds()))
			fmt.Printf("%-40s |%s|\r", K, dots)
			lastProgress = progress
		}
		_ = lastProgress
		return progress
	}

	var refreshThreshold = 0.8

	go func() {
		// refresh screen with progress bar
		for {
			nowFrac := drawProgress(0)
			if nowFrac > refreshThreshold {
				break
			}
			time.Sleep(time.Second)
		}
	}()
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			// there can be lz4 end block that's skipped
			// but you have to read it to properly calculate the hash
			actuallyRead, potentiallyRead := hasher.readByteCount, int(archiveReq.ContentLength)
			if actuallyRead < potentiallyRead {
				hasher.Read(make([]byte, potentiallyRead - actuallyRead))
			}
			shasum := hasher.sha256.Sum(nil)
			fmt.Printf("\nDone, SHA256 is: %x\n", shasum)
			fmt.Printf(`You can use gotip with any go command now. Start with 'gotip help' or 'gotip version'`)
			os.WriteFile(filepath.Join(newP.GoPath, ".tipsuccess"), []byte(fmt.Sprintf("%s\n%x", fileName, shasum)), os.ModePerm)
			break
		}


		if err != nil {
			log.Fatalf("\nextract: Next() failed: %s", err.Error())
		}
		drawProgress(int(header.Size))

		header.Name = filepath.Join(gopath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				log.Fatalf("\nextract: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			outFile.Chmod(os.FileMode(header.Mode))
			if err != nil {
				log.Fatalf("\nextract: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("\nextract: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"\nextract: unknown type: %b in %s",
				header.Typeflag,
				header.Name)
		}
	}
}
