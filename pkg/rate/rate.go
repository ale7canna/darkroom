package rate

import (
	"darkroom/pkg/database"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var StoreRate = &cobra.Command{
	Use:  "store-rate",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		_, err := database.InitClient()
		if err != nil {
			return err
		}
		rates, err := readRates(dir)
		if err != nil {
			return err
		}
		for p, rate := range rates {
			log.Println(p, rate)
		}
		return nil
	},
}

func readRates(dir string) (map[string]int, error) {
	rates := make(map[string]int)
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		p := path.Join(dir, info.Name())
		rate, err := getPictureRate(p)
		if err != nil {
			return nil, err
		}
		rates[path.Base(p)] = rate
	}
	return rates, nil
}

func getPictureRate(path string) (int, error) {
	buffer := make([]byte, 1024)
	attrSize, err := unix.Getxattr(path, "user.baloo.rating", buffer)
	if err != nil {
		if errors.Is(err, unix.ENODATA) {
			return 0, errors.New(fmt.Sprintf("file %s has no rating", path))
		}
		return 0, err
	}

	val := strings.TrimSpace(string(buffer[:attrSize]))
	return strconv.Atoi(val)
}
