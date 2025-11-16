package rate

import (
	"darkroom/pkg/database"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type rates struct {
	db *sqlx.DB
}

var StoreRate = &cobra.Command{
	Use:  "store-rate",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		db, err := database.InitClient()
		if err != nil {
			return err
		}
		rates := &rates{
			db: db,
		}
		picRates, err := rates.read(dir)
		if err != nil {
			return err
		}
		picturesWithStats, err := rates.storeAndUpdateStats(picRates)
		if err != nil {
			return err
		}
		for _, p := range picturesWithStats {
			log.Println(p)
		}
		return nil
	},
}

type PicWithRate struct {
	PicName   string
	AvgRate   float32
	RateCount int
}

type pictureRate struct {
	Id                 int `db:"id"`
	PictureId          int `db:"picture_id"`
	Rating             int `db:"rating"`
	CreatedAtTimeStamp int `db:"created_at_ts"`
}

func (r *rates) storeAndUpdateStats(rates map[string]int) ([]*PicWithRate, error) {
	var result []*PicWithRate
	for pic, rate := range rates {
		stat, err := r.storeAndUpdatePic(pic, rate)
		if err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	return result, nil
}

func (*rates) read(dir string) (map[string]int, error) {
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

type Picture struct {
	Id        int    `db:"id"`
	Name      string `db:"name"`
	Path      string `db:"path"`
	Directory string `db:"directory"`
}

func (r *rates) storeAndUpdatePic(pic string, rate int) (*PicWithRate, error) {
	p := &Picture{}
	dir := path.Dir(pic)
	name := path.Base(pic)
	err := r.db.Get(p, `
insert into picture (name, path, directory) values (?, ?, ?)
on conflict(name) do update set
path = excluded.path,
directory = excluded.directory
returning *`, name, pic, dir)
	if err != nil {
		return nil, err
	}
	return &PicWithRate{}, nil
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
