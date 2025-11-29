package main

import (
	"darkroom/pkg/rate"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(split)
	rootCmd.AddCommand(pick)
	rootCmd.AddCommand(rate.StoreRate)
	rootCmd.AddCommand(rate.ResetRate)
	rootCmd.AddCommand(rate.PickRated)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var pick = &cobra.Command{
	Use:  "pick",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		err := makeDirs(dir)
		if err != nil {
			return err
		}
		err = linkNEFFromJPG(dir, "to_process_jpg", "original_raw", "to_process_raw")
		if err != nil {
			return err
		}
		log.Println("completed")
		return nil
	},
}

func linkNEFFromJPG(basePath string, baseDir string, nefSource string, nefTarget string) error {
	infos, err := os.ReadDir(path.Join(basePath, baseDir))
	if err != nil {
		return err
	}
	err = createDir(path.Join(basePath, nefTarget))
	if err != nil {
		return err
	}
	var directories []string
	for _, info := range infos {
		if info.IsDir() {
			directories = append(directories, info.Name())
			continue
		}
		if filepath.Ext(info.Name()) != ".JPG" {
			log.Println(fmt.Sprintf("skipping %s", info.Name()))
			continue
		}
		p := path.Join(basePath, nefSource, info.Name())
		rawFile := strings.Replace(p, ".JPG", ".NEF", 1)
		err = os.Link(rawFile, path.Join(basePath, nefTarget, filepath.Base(rawFile)))
		if err != nil {
			var fsErr *os.LinkError
			if errors.As(err, &fsErr) && errors.Is(fsErr.Err, fs.ErrExist) {
				log.Println(fmt.Sprintf("file %s already linked", filepath.Base(rawFile)))
			} else {
				log.Println(fmt.Sprintf("error linking %s. err: %s", filepath.Base(rawFile), err))
			}
		}
	}
	for _, dir := range directories {
		err = linkNEFFromJPG(basePath, path.Join(baseDir, dir), nefSource, path.Join(nefTarget, path.Base(dir)))
		if err != nil {
			return err
		}
	}
	return nil
}

var split = &cobra.Command{
	Use:  "split",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		err := makeDirs(dir)
		if err != nil {
			log.Fatalln(err)
		}
		infos, err := os.ReadDir(dir)
		if err != nil {
			log.Fatalln(err)
		}
		for _, info := range infos {
			p := path.Join(dir, info.Name())
			log.Println(p)
			e := moveIf(p, ".NEF", "original_raw")
			if e != nil {
				log.Fatalln(e)
			}
			e = moveIf(p, ".JPG", "original_jpg")
			if e != nil {
				log.Fatalln(e)
			}
		}
		jpgs, err := os.ReadDir(path.Join(dir, "original_jpg"))
		if err != nil {
			log.Fatalln(err)
		}
		for _, jpg := range jpgs {
			p := path.Join(dir, "original_jpg", jpg.Name())
			log.Println(p)
			e := cloneIf(p, ".JPG", path.Join(dir, "to_process_jpg"))
			if e != nil {
				log.Fatalln(e)
			}
		}
		log.Println("completed")
	},
}

func cloneIf(p string, ext string, dest string) error {
	if filepath.Ext(p) == ext {
		return os.Link(p, path.Join(dest, filepath.Base(p)))
	}
	return nil
}

func makeDirs(dir string) error {
	err := createDir(path.Join(dir, "original_raw"))
	if err != nil {
		return err
	}
	err = createDir(path.Join(dir, "original_jpg"))
	if err != nil {
		return err
	}
	err = createDir(path.Join(dir, "to_process_jpg"))
	if err != nil {
		return err
	}
	err = createDir(path.Join(dir, "to_process_raw"))
	if err != nil {
		return err
	}
	return nil
}

func moveIf(p string, ext string, dest string) error {
	if filepath.Ext(p) == ext {
		return os.Rename(p, path.Join(filepath.Dir(p), dest, filepath.Base(p)))
	}
	return nil
}

func createDir(dir string) error {
	err := os.Mkdir(dir, fs.ModeDir|fs.ModePerm)
	if err != nil {
		var pErr *fs.PathError
		if errors.As(err, &pErr) && errors.Is(pErr.Err, fs.ErrExist) {
			log.Println(fmt.Sprintf("%s directory already exists", dir))
			return nil
		}
	}
	return err
}
