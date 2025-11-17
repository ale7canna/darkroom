package rate

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sys/unix"
)

const ratingAttrKey = "user.baloo.rating"

type rates struct {
	db *sqlx.DB
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
	slices.SortFunc(result, func(a, b *PicWithRate) int {
		if a.AvgRate <= b.AvgRate {
			return 1
		} else {
			return -1
		}
	})
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

type dbPicture struct {
	Id        int    `db:"id"`
	Name      string `db:"name"`
	Path      string `db:"path"`
	Directory string `db:"directory"`
}

type dbPicStats struct {
	AvgRating float32 `db:"avg_rating"`
	RateCount int     `db:"rate_count"`
}

func (r *rates) storeAndUpdatePic(pic string, rate int) (*PicWithRate, error) {
	dir := path.Dir(pic)
	name := path.Base(pic)
	p := &dbPicture{}
	err := r.db.Get(p, `
insert into picture (name, path, directory) values (?, ?, ?)
on conflict(name) do update set
path = excluded.path,
directory = excluded.directory
returning *`, name, pic, dir)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(`insert into picture_rating (picture_id, rating, created_at_ts) values (?, ?, ?)`,
		p.Id, rate, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}

	stats := &dbPicStats{}
	err = r.db.Get(stats, `
insert into picture_stats (picture_id, avg_rating, rate_count)
select picture_id, avg(rating) as avg_rating, count(*) as rate_count from picture_rating where picture_id = ?
group by picture_id
on conflict(picture_id) do update set
	avg_rating = excluded.avg_rating,
	rate_count = excluded.rate_count
returning avg_rating, rate_count;`, p.Id)
	if err != nil {
		return nil, err
	}
	return &PicWithRate{
		PicName:   p.Name,
		AvgRate:   stats.AvgRating,
		RateCount: stats.RateCount,
	}, nil
}

func (r *rates) reset(dir string) error {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, info := range infos {
		p := path.Join(dir, info.Name())
		err := resetPictureRate(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPictureRate(path string) (int, error) {
	buffer := make([]byte, 1024)
	attrSize, err := unix.Getxattr(path, ratingAttrKey, buffer)
	if err != nil {
		if errors.Is(err, unix.ENODATA) {
			return 0, errors.New(fmt.Sprintf("file %s has no rating", path))
		}
		return 0, err
	}

	val := strings.TrimSpace(string(buffer[:attrSize]))
	return strconv.Atoi(val)
}

func resetPictureRate(path string) error {
	return unix.Removexattr(path, ratingAttrKey)
}
