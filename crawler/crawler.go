package crawler

import (
	"math/rand"
	"time"

	"github.com/flywave/go-tileproxy/downloader"
	"github.com/flywave/go-tileproxy/processer"
	"github.com/flywave/go-tileproxy/request"
	"github.com/flywave/go-tileproxy/resamper"
	"github.com/flywave/go-tileproxy/scheduler"
	"github.com/flywave/go-tileproxy/tile"
)

type Crawler struct {
	taskname         string
	pTileProcesser   processer.TileProcesser
	pDownloader      downloader.Downloader
	pScheduler       scheduler.Scheduler
	pResampes        []resamper.Resamper
	mc               ResourceManage
	threadnum        uint
	exitWhenComplete bool
	startSleeptime   uint
	endSleeptime     uint
	sleeptype        string
}

func NewCrawler(pageinst processer.TileProcesser, taskname string) *Crawler {
	ap := &Crawler{taskname: taskname, pTileProcesser: pageinst}

	ap.exitWhenComplete = true
	ap.sleeptype = "fixed"
	ap.startSleeptime = 0

	if ap.pScheduler == nil {
		ap.SetScheduler(scheduler.NewQueueScheduler(false))
	}

	if ap.pDownloader == nil {
		ap.SetDownloader(downloader.NewHttpDownloader())
	}

	ap.pResampes = make([]resamper.Resamper, 0)

	return ap
}

func (this *Crawler) Taskname() string {
	return this.taskname
}

func (this *Crawler) Get(url string, respType string) tile.Tile {
	req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
	return this.GetByRequest(req)
}

func (this *Crawler) GetAll(urls []string, respType string) []tile.Tile {
	for _, u := range urls {
		req := request.NewRequest(u, respType, "", "GET", "", nil, nil, nil, nil)
		this.AddRequest(req)
	}

	pip := resamper.NewTileResamper()
	this.AddResamper(pip)

	this.Run()

	return pip.GetCollected()
}

func (this *Crawler) GetByRequest(req *request.Request) tile.Tile {
	var reqs []*request.Request
	reqs = append(reqs, req)
	items := this.GetAllByRequest(reqs)
	if len(items) != 0 {
		return items[0]
	}
	return nil
}

func (this *Crawler) GetAllByRequest(reqs []*request.Request) []tile.Tile {
	for _, req := range reqs {
		this.AddRequest(req)
	}

	pip := resamper.NewTileResamper()
	this.AddResamper(pip)

	this.Run()

	return pip.GetCollected()
}

func (this *Crawler) Run() {
	if this.threadnum == 0 {
		this.threadnum = 1
	}
	this.mc = NewResourceManageChan(this.threadnum)

	for {
		req := this.pScheduler.Poll()

		if this.mc.Has() == 0 && req == nil && this.exitWhenComplete {
			this.pTileProcesser.Finish()
			break
		} else if req == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		this.mc.GetOne()

		go func(req *request.Request) {
			defer this.mc.FreeOne()
			this.tileProcess(req)
		}(req)
	}
	this.close()
}

func (this *Crawler) close() {
	this.SetScheduler(scheduler.NewQueueScheduler(false))
	this.SetDownloader(downloader.NewHttpDownloader())
	this.pResampes = make([]resamper.Resamper, 0)
	this.exitWhenComplete = true
}

func (this *Crawler) AddResamper(p resamper.Resamper) *Crawler {
	this.pResampes = append(this.pResampes, p)
	return this
}

func (this *Crawler) SetScheduler(s scheduler.Scheduler) *Crawler {
	this.pScheduler = s
	return this
}

func (this *Crawler) GetScheduler() scheduler.Scheduler {
	return this.pScheduler
}

func (this *Crawler) SetDownloader(d downloader.Downloader) *Crawler {
	this.pDownloader = d
	return this
}

func (this *Crawler) GetDownloader() downloader.Downloader {
	return this.pDownloader
}

func (this *Crawler) SetThreadnum(i uint) *Crawler {
	this.threadnum = i
	return this
}

func (this *Crawler) GetThreadnum() uint {
	return this.threadnum
}

func (this *Crawler) SetExitWhenComplete(e bool) *Crawler {
	this.exitWhenComplete = e
	return this
}

func (this *Crawler) GetExitWhenComplete() bool {
	return this.exitWhenComplete
}

func (this *Crawler) SetSleepTime(sleeptype string, s uint, e uint) *Crawler {
	this.sleeptype = sleeptype
	this.startSleeptime = s
	this.endSleeptime = e
	if this.sleeptype == "rand" && this.startSleeptime >= this.endSleeptime {
		panic("startSleeptime must smaller than endSleeptime")
	}
	return this
}

func (this *Crawler) sleep() {
	if this.sleeptype == "fixed" {
		time.Sleep(time.Duration(this.startSleeptime) * time.Millisecond)
	} else if this.sleeptype == "rand" {
		sleeptime := rand.Intn(int(this.endSleeptime-this.startSleeptime)) + int(this.startSleeptime)
		time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	}
}

func (this *Crawler) AddUrl(url string, respType string) *Crawler {
	req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
	this.AddRequest(req)
	return this
}

func (this *Crawler) AddUrlEx(url string, respType string, headerFile string, proxyHost string) *Crawler {
	req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
	this.AddRequest(req.AddHeaderFile(headerFile).AddProxyHost(proxyHost))
	return this
}

func (this *Crawler) AddUrlWithHeaderFile(url string, respType string, headerFile string) *Crawler {
	req := request.NewRequestWithHeaderFile(url, respType, headerFile)
	this.AddRequest(req)
	return this
}

func (this *Crawler) AddUrls(urls []string, respType string) *Crawler {
	for _, url := range urls {
		req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
		this.AddRequest(req)
	}
	return this
}

func (this *Crawler) AddUrlsWithHeaderFile(urls []string, respType string, headerFile string) *Crawler {
	for _, url := range urls {
		req := request.NewRequestWithHeaderFile(url, respType, headerFile)
		this.AddRequest(req)
	}
	return this
}

func (this *Crawler) AddUrlsEx(urls []string, respType string, headerFile string, proxyHost string) *Crawler {
	for _, url := range urls {
		req := request.NewRequest(url, respType, "", "GET", "", nil, nil, nil, nil)
		this.AddRequest(req.AddHeaderFile(headerFile).AddProxyHost(proxyHost))
	}
	return this
}

func (this *Crawler) AddRequest(req *request.Request) *Crawler {
	if req == nil {
		return this
	} else if req.GetUrl() == "" {
		return this
	}
	this.pScheduler.Push(req)
	return this
}

func (this *Crawler) AddRequests(reqs []*request.Request) *Crawler {
	for _, req := range reqs {
		this.AddRequest(req)
	}
	return this
}

func (this *Crawler) tileProcess(req *request.Request) {
	var p tile.Tile

	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(string); ok {
				//logger.FatalIf(errors.New(""), "[ERROR] strerr")
			} else {
				//logger.FatalIf(errors.New(""), "[ERROR] tileProcess error")
			}
		}
	}()

	for i := 0; i < 3; i++ {
		this.sleep()
		p = this.pDownloader.Download(req)
		if p.Success() {
			break
		}
	}

	if !p.Success() {
		return
	}

	this.pTileProcesser.Process(p)
}
