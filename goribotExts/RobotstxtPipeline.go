package goribotExts

import (
	"github.com/slyrz/robots"
	"github.com/zhshch2002/goribot"
	"log"
	"net/http"
	"strings"
)

type RobotstxtPipeline struct {
	goribot.Pipeline
	RobotsTxt *robots.Robots
	SiteUrl   string
}

func NewRobotstxtPipeline(siteUrl string) *RobotstxtPipeline {
	return &RobotstxtPipeline{SiteUrl: siteUrl}
}
func (s *RobotstxtPipeline) Init(spider *goribot.Spider) {
	s.RobotsTxt = robots.New(strings.NewReader(""), spider.UserAgent)
	if !strings.HasSuffix(s.SiteUrl, "/") {
		s.SiteUrl += "/"
	}
	resp, err := http.Get(s.SiteUrl + "robots.txt")
	defer resp.Body.Close()
	if err != nil {
		log.Println("get robots.txt error", err)
	}
	s.RobotsTxt = robots.New(resp.Body, spider.UserAgent)
}
func (s *RobotstxtPipeline) OnNewRequest(spider *goribot.Spider, preResp *goribot.Response, request *goribot.Request) *goribot.Request {
	if !s.RobotsTxt.Allow(request.Url.Path) {
		return nil
	}
	return request
}
