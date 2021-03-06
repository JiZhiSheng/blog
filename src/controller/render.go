package controller

import (
	"container/list"
	"framework/base/config"
	"info"
	"model"
	"sort"
	"strings"
	"sync"
	"time"
	"fmt"
)

type tagRender struct {
	Tag   string
	Count int
}

type timeRender struct {
	Date  string
	Year  int
	Month int
}

type rankRender struct {
	ID    int
	Title string
	Index int
	Hot   int
}
type rankList []*rankRender

func (r rankList) Len() int {
	return len(r)
}

func (r rankList) Swap(blog1, blog2 int) {
	r[blog1], r[blog2] = r[blog2], r[blog1]
}

func (r rankList) Less(blog1, blog2 int) bool {
	return r[blog1].Hot > r[blog2].Hot
}

type sideRender struct {
	BlogTagList     []*tagRender
	BlogTimeList    []*timeRender
	BlogHotBlogList rankList
}

type hostRender struct {
	Host string
}

var staticHostRender *hostRender = nil
var hostRenderOnce sync.Once

func buildHostRender() *hostRender {
	hostRenderOnce.Do(func() {
		staticHostRender = &hostRender{}
		staticHostRender.Host = config.GetDefaultConfigJsonReader().GetString("net.host")
		port := config.GetDefaultConfigJsonReader().GetInteger("net.port")
		protocol := config.GetDefaultConfigJsonReader().Get("net.protocol").(string)
		if !strings.HasPrefix(staticHostRender.Host, protocol+"://") {
			staticHostRender.Host = protocol + "://" + staticHostRender.Host
			if protocol == "http" &&  port != 80 {
				staticHostRender.Host += fmt.Sprintf(":%d", port)
			} else if protocol == "https" && port != 443 {
				staticHostRender.Host += fmt.Sprintf(":%d", port)
			}
		}
	})
	return staticHostRender
}

func buildSideRender(blogList *list.List) *sideRender {
	var topRender sideRender
	var tagMap map[string]int = make(map[string]int)
	var timeMap map[string]int64 = make(map[string]int64)
	for iter := blogList.Front(); iter != nil; iter = iter.Next() {
		inf := iter.Value.(info.BlogInfo)
		commentCount, _ := model.ShareCommentModel().FetchCommentCount(info.CommentType_Blog, inf.BlogID)
		rank := &rankRender{ID: inf.BlogID, Title: inf.BlogTitle, Hot: inf.BlogVisitCount + commentCount*5}
		topRender.BlogHotBlogList = append(topRender.BlogHotBlogList, rank)
		for tag := range inf.BlogTagList {
			tagMap[inf.BlogTagList[tag]]++
		}
		timeMap[time.Unix(inf.BlogTime, 0).Format("2006年01月")] = inf.BlogTime
	}
	var tagList []*tagRender = nil
	for k, v := range tagMap {
		tagList = append(tagList, &tagRender{k, v})
	}
	var blogTimeList BlogTimeList = nil
	for k, v := range timeMap {
		blogTimeList = append(blogTimeList, &BlogTime{k, v})
	}
	sort.Sort(blogTimeList)
	var blogTimeStringList []*timeRender = nil
	for i := range blogTimeList {
		renderTime := time.Unix(blogTimeList[i].Time, 0)
		blogTimeStringList = append(blogTimeStringList, &timeRender{blogTimeList[i].Tag,
			renderTime.Year(), int(renderTime.Month())})
	}
	topRender.BlogTagList = tagList
	topRender.BlogTimeList = blogTimeStringList
	if len(topRender.BlogHotBlogList) > 6 {
		topRender.BlogHotBlogList = topRender.BlogHotBlogList[:6]
	}
	sort.Sort(topRender.BlogHotBlogList)
	for i, rank := range topRender.BlogHotBlogList {
		rank.Index = i + 1
	}
	return &topRender
}
