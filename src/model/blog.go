package model

import (
	"container/list"
	"fmt"
	"framework/database"
	"info"
	"strings"
	"sync"
	"time"
)

type blogModel struct {
}

const (
	kBlogTableName    = "blog"
	kBlogId           = "id"
	kBlogUUID         = "uuid"
	kBlogTitle        = "title"
	kBlogSortType     = "sort"
	kBlogTag          = "tag"
	kBlogTime         = "time"
	kBlogVisitCount   = "visit"
	kBlogPraiseCount  = "praise"
	kBlogDissentCount = "dissent"
)

var blogModelInstance *blogModel = nil

var blogOnce sync.Once

func ShareBlogModel() *blogModel {
	blogOnce.Do(func() {
		blogModelInstance = &blogModel{}
	})
	return blogModelInstance
}

func (c *blogModel) CreateTable() error {
	if database.DatabaseInstance().DoesTableExist(kBlogTableName) {
		return nil
	}
	sql := fmt.Sprintf(`
	CREATE TABLE %s (
		%s int(32) unsigned NOT NULL AUTO_INCREMENT,
		%s varchar(128) NOT NULL,
		%s varchar(256) NOT NULL,
		%s varchar(256) NOT NULL,
		%s varchar(256) NOT NULL,
		%s int(64) NOT NULL,
		%s int(32) DEFAULT '0',
		%s int(32) DEFAULT '0',
		%s int(32) DEFAULT '0',
		PRIMARY KEY (%s)
	) CHARSET=utf8;`, kBlogTableName, kBlogId,
		kBlogUUID, kBlogTitle, kBlogSortType, kBlogTag, kBlogTime, kBlogVisitCount,
		kBlogPraiseCount, kBlogDissentCount, kBlogId)
	_, err := database.DatabaseInstance().DB.Exec(sql)
	return err
}

func (b *blogModel) InsertBlog(uuid string, title string, sortType string, tagList []string) error {
	currentTime := time.Now().Unix()
	tag := strings.Join(tagList, "||")
	sql := fmt.Sprintf("insert into %s(%s, %s, %s, %s, %s) values(?, ?, ?, ?, ?)",
		kBlogTableName, kBlogUUID, kBlogTitle, kBlogSortType, kBlogTag, kBlogTime)
	stat, err := database.DatabaseInstance().DB.Prepare(sql)
	if err == nil {
		defer stat.Close()
		_, err := stat.Exec(uuid, title, sortType, tag, currentTime)
		return err
	}
	return err
}

func (b *blogModel) UpdateBlog(uuid string, title string, sortType string, tagList []string) error {
	currentTime := time.Now().Unix()
	tag := strings.Join(tagList, "||")
	sql := fmt.Sprintf("update %s set %s = ?, %s = ?, %s = ?, %s = ? where %s = ?",
		kBlogTableName, kBlogTitle, kBlogSortType, kBlogTag, kBlogTime, kBlogUUID)
	_, err := database.DatabaseInstance().DB.Exec(sql, title, sortType, tag, currentTime, uuid)
	return err
}

func (b *blogModel) BlogIsExistByUUID(uuid string) (bool, error) {
	sql := fmt.Sprintf("select * from %s where %s = ?", kBlogTableName, kBlogUUID)
	rows, err := database.DatabaseInstance().DB.Query(sql, uuid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			return true, nil
		}
	}
	return false, err
}

func (b *blogModel) BlogIsExistByBlogID(blogId int) (bool, error) {
	sql := fmt.Sprintf("select * from %s where %s = ?", kBlogTableName, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql, blogId)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			return true, nil
		}
	}
	return false, err
}

func (b *blogModel) FetchAllBlog() (*list.List, error) {
	sql := fmt.Sprintf("select * from %s order by %s desc", kBlogTableName, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql)
	if err == nil {
		defer rows.Close()
		var blogList *list.List = list.New()
		for rows.Next() {
			var blog info.BlogInfo
			var tag string
			err = rows.Scan(&blog.BlogID, &blog.BlogUUID, &blog.BlogTitle,
				&blog.BlogSortType, &tag, &blog.BlogTime, &blog.BlogVisitCount,
				&blog.BlogPraiseCount, &blog.BlogDissentCount)
			if err == nil {
				blog.BlogTagList = strings.Split(tag, "||")
				blogList.PushBack(blog)
			}
		}
		return blogList, err
	}
	fmt.Println(err)
	return nil, err
}

func (b *blogModel) FetchBlogByBlogID(blogID int) (*info.BlogInfo, error) {
	sql := fmt.Sprintf("select* from %s where %s = ?", kBlogTableName, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql, blogID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var blog info.BlogInfo
			var tag string
			err = rows.Scan(&blog.BlogID, &blog.BlogUUID, &blog.BlogTitle,
				&blog.BlogSortType, &tag, &blog.BlogTime, &blog.BlogVisitCount,
				&blog.BlogPraiseCount, &blog.BlogDissentCount)
			if err == nil {
				blog.BlogTagList = strings.Split(tag, "||")
				return &blog, nil
			}
			break
		}
	}
	return nil, err
}

func (b *blogModel) GetBlogUUIDByBlogID(blogID int) (string, error) {
	sql := fmt.Sprintf("select %s from %s where %s = ?", kBlogUUID, kBlogTableName, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql, blogID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var uuid string
			err = rows.Scan(&uuid)
			if err == nil {
				return uuid, nil
			}
			break
		}
	}
	return "", err
}

func (b *blogModel) FetchBlogByUUID(uuid string) (*info.BlogInfo, error) {
	sql := fmt.Sprintf("select* from %s where %s = ?", kBlogTableName, kBlogUUID)
	rows, err := database.DatabaseInstance().DB.Query(sql, uuid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var blog info.BlogInfo
			var tag string
			err = rows.Scan(&blog.BlogID, &blog.BlogUUID, &blog.BlogTitle,
				&blog.BlogSortType, &tag, &blog.BlogTime, &blog.BlogVisitCount,
				&blog.BlogPraiseCount, &blog.BlogDissentCount)
			if err == nil {
				blog.BlogTagList = strings.Split(tag, "||")
				return &blog, nil
			}
			break
		}
	}
	return nil, err
}

func (b *blogModel) FetchAllSortType() ([]string, error) {
	sql := fmt.Sprintf("select %s from %s distinct", kBlogSortType, kBlogTableName)
	rows, err := database.DatabaseInstance().DB.Query(sql)
	if err == nil {
		defer rows.Close()
		var sortTypeList []string
		for rows.Next() {
			var sortType string
			err = rows.Scan(&sortType)
			if err == nil {
				sortTypeList = append(sortTypeList, sortType)
			}
		}
		return sortTypeList, nil
	}
	return nil, err
}

func (b *blogModel) FetchAllBlogBySortType(sortType string) (*list.List, error) {
	sql := fmt.Sprintf("select * from %s where %s = ? order by %s desc",
		kBlogTableName, kBlogSortType, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql, sortType)
	if err == nil {
		defer rows.Close()
		var blogList *list.List = list.New()
		for rows.Next() {
			var blog info.BlogInfo
			var tag string
			err = rows.Scan(&blog.BlogID, &blog.BlogUUID, &blog.BlogTitle,
				&blog.BlogSortType, &tag, &blog.BlogTime, &blog.BlogVisitCount,
				&blog.BlogPraiseCount, &blog.BlogDissentCount)
			if err == nil {
				blog.BlogTagList = strings.Split(tag, "||")
				blogList.PushBack(blog)
			}
		}
		return blogList, err
	}
	return nil, err
}

func (b *blogModel) FetchAllBlogByTime(beginTime int64, endTime int64) (*list.List, error) {
	sql := fmt.Sprintf("select * from %s where %s >= ? and %s <= ? order by %s desc", kBlogTableName, kBlogTime, kBlogTime, kBlogId)
	rows, err := database.DatabaseInstance().DB.Query(sql, beginTime, endTime)
	if err == nil {
		defer rows.Close()
		var blogList *list.List = list.New()
		for rows.Next() {
			var blog info.BlogInfo
			var tag string
			err = rows.Scan(&blog.BlogID, &blog.BlogUUID, &blog.BlogTitle,
				&blog.BlogSortType, &tag, &blog.BlogTime, &blog.BlogVisitCount,
				&blog.BlogPraiseCount, &blog.BlogDissentCount)
			if err == nil {
				blog.BlogTagList = strings.Split(tag, "||")
				blogList.PushBack(blog)
			}
		}
		return blogList, err
	}
	return nil, err
}

func (b *blogModel) AddVisitCount(blogId int) error {
	sql := fmt.Sprintf("update %s set visit = visit + 1 where %s = ?", kBlogTableName, kBlogId)
	_, err := database.DatabaseInstance().DB.Exec(sql, blogId)
	return err
}

func (b *blogModel) DeleteBlog(blogId int) error {
	sql := fmt.Sprintf("delete from %s where %s = ?", kBlogTableName, kBlogId)
	_, err := database.DatabaseInstance().DB.Exec(sql, blogId)
	return err
}
