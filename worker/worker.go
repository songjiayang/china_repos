package worker

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/songjiayang/china_repos/github"
)

const (
	dataFilePath  = "./data.csv"
	cacheFilePath = "./.cache"
	maxPage       = 100
)

type Worker struct {
	language string
	maxStars int
	minStars int

	page     int
	client   *github.Client
	dataFile *os.File
}

func New(l string, minStars int, client *github.Client) *Worker {
	w := &Worker{
		language: l,
		minStars: minStars,
		page:     1,
		client:   client,
	}

	w.prepare()

	return w
}

func (w *Worker) Run() {
	defer w.dataFile.Close()

	for {
		w.updateCache()
		repos := w.fetchRepos()

		if len(repos) == 0 {
			return
		}

		w.updateData(repos)

		if w.page == maxPage {
			w.maxStars = repos[len(repos)-1].IntStars()
			w.page = 0
		}

		w.page += 1
	}
}

func (w *Worker) fetchRepos() []*github.Repository {
	q := w.queryStr()

	log.Printf("start fetch repos with params: %s %d", q, w.page)
	defer log.Printf("finish fetch repos with params: %s %d", q, w.page)

	return w.client.Repositories(q, w.page)
}

func (w *Worker) prepare() {
	dataFile, isNew := w.createOrOpenFile(dataFilePath, os.O_APPEND|os.O_RDWR)
	if isNew {
		header := "名称,描述,开源协议,关注人数,最后更新时间,国家,地区,邮箱,是否为组织\n"
		dataFile.Write([]byte(header))
	}

	w.dataFile = dataFile
}

func (w *Worker) createOrOpenFile(filePath string, mod int) (file *os.File, isNew bool) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		isNew = true
		os.Create(filePath)
	}

	file, err = os.OpenFile(filePath, mod, 0644)
	if err != nil {
		log.Fatalf("open file with error: %v", err)
	}

	return
}

func (w *Worker) queryStr() string {
	if w.maxStars != 0 {
		return fmt.Sprintf("stars:%d..%d language:%s", w.minStars, w.maxStars, w.language)
	}

	return fmt.Sprintf("stars:>%d language:%s", w.minStars, w.language)
}

func (w *Worker) updateCache() {
	content := fmt.Sprintf("%s,%d", w.queryStr(), w.page)
	ioutil.WriteFile(cacheFilePath, []byte(content), 0644)
}

func (w *Worker) updateData(repos []*github.Repository) {
	for _, repo := range repos {
		content := fmt.Sprintf(`%s,"%s","%s",%s,%s,%s,"%s",%s,%t`,
			repo.Name,
			repo.Description,
			repo.Protocol,
			repo.Stars,
			repo.UpdatedAt,
			repo.Country,
			repo.Location,
			repo.Email,
			repo.IsOrganization)

		lineContent := fmt.Sprintf("%s\n", content)
		w.dataFile.Write([]byte(lineContent))
	}
}
