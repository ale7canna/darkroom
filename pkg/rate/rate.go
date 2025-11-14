package rate

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		infos, err := os.ReadDir(dir)
		if err != nil {
			log.Fatalln(err)
		}
		rates := make(map[string]int)
		for _, info := range infos {
			p := path.Join(dir, info.Name())
			rate, err := getPictureRate(p)
			if err != nil {
				log.Fatalln(err)
			}
			rates[path.Base(p)] = rate
		}
		for p, rate := range rates {
			log.Println(p, rate)
		}
	},
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
