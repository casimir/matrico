package main

import (
	"archive/zip"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/casimir/matrico/generate"
)

func buildSpecURL(release string) string {
	return fmt.Sprintf("https://github.com/matrix-org/matrix-doc/archive/%s.zip", release)
}

func buildSpecCacheDir(name, version string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("could not get cwd: %s", err))
	}
	return path.Join(cwd, "specs", name, version)
}

func downloadArchive(release string) (string, error) {
	out, err := ioutil.TempFile("", strings.ReplaceAll(release, "/", "-"))
	if err != nil {
		return "", err
	}
	defer out.Close()
	url := buildSpecURL(release)
	log.Printf("downloading archive from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return out.Name(), err
}

func cacheArchive(archive, cacheDir, root string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		components := strings.SplitN(f.Name, "/", 4)
		keep := len(components) > 3 && components[1] == "api" && components[2] == root
		if !keep {
			continue
		}

		dest := path.Join(cacheDir, components[3])
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, os.ModePerm); err != nil {
				return err
			}
		} else {
			r, err := f.Open()
			if err != nil {
				return err
			}
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(dest, data, f.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

func runGenerate(name, version, release string, skipOperationIDs []string) error {
	cacheDir := buildSpecCacheDir(name, version)
	info, _ := os.Stat(cacheDir)
	var archive string
	if info == nil || !info.IsDir() {
		var err error
		archive, err = downloadArchive(release)
		if err != nil {
			return err
		}
		log.Printf("filling cache: %s", cacheDir)
		if err := cacheArchive(archive, cacheDir, name); err != nil {
			return err
		}
	} else {
		log.Printf("using cache: %s", cacheDir)
	}

	major := strings.SplitN(version, ".", 2)[0]
	pkg := name + version
	pkg = strings.ReplaceAll(pkg, ".", "")
	pkg = strings.ReplaceAll(pkg, "-", "")
	api, err := generate.ParseAPISpec(cacheDir, major, pkg, skipOperationIDs)
	if err != nil {
		panic(err)
	}
	pkgDir := filepath.Join("api", pkg)
	sourceFile := filepath.Join(pkgDir, "api.go")
	if err := os.MkdirAll(pkgDir, os.ModePerm); err != nil {
		return err
	}
	source := api.Source()
	formatted, err := format.Source(source)
	if err != nil {
		log.Printf("failed to format source: %s", err)
		formatted = source
	}
	if err := ioutil.WriteFile(sourceFile, formatted, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	_ = config
	for _, it := range config.Specs {
		if err := runGenerate(it.Name, it.Version, it.Release, it.Blacklist); err != nil {
			log.Fatal(err)
		}
	}
}
