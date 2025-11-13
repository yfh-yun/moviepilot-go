package douban

import (
	"path/filepath"
	
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/dom"
)

// DoubanScraper 豆瓣刮削�?type DoubanScraper struct {
	forceNfo bool
	forceImg bool
}

// NewDoubanScraper 创建DoubanScraper实例
func NewDoubanScraper() *DoubanScraper {
	return &DoubanScraper{}
}

// GetMetadataNfo 获取NFO文件内容文本
func (ds *DoubanScraper) GetMetadataNfo(mediainfo *models.MediaInfo, season *int) string {
	var doc *dom.Document
	
	if mediainfo.Type == types.MediaTypeMovie {
		// 电影元数据文�?		doc = ds.genMovieNfoFile(mediainfo)
	} else {
		if season != nil {
			// 季元数据文件
			doc = ds.genTvSeasonNfoFile(mediainfo, *season)
		} else {
			// 电视剧元数据文件
			doc = ds.genTvNfoFile(mediainfo)
		}
	}
	
	if doc != nil {
		return doc.ToPrettyXML("  ")
	}
	
	return ""
}

// GetMetadataImg 获取图片内容
func (ds *DoubanScraper) GetMetadataImg(mediainfo *models.MediaInfo, season, episode *int) map[string]string {
	retMap := make(map[string]string)
	
	if season != nil {
		// 豆瓣无季图片
		return retMap
	}
	
	if episode != nil {
		// 豆瓣无集图片
		return retMap
	}
	
	if mediainfo.PosterPath != "" {
		ext := filepath.Ext(mediainfo.PosterPath)
		retMap["poster"+ext] = mediainfo.PosterPath
	}
	
	if mediainfo.BackdropPath != "" {
		ext := filepath.Ext(mediainfo.BackdropPath)
		retMap["backdrop"+ext] = mediainfo.BackdropPath
	}
	
	return retMap
}

// genCommonNfo 生成公共NFO信息
func (ds *DoubanScraper) genCommonNfo(mediainfo *models.MediaInfo, doc *dom.Document, root *dom.Node) *dom.Document {
	// 简�?	xplot := doc.AddNode(root, "plot")
	xplot.AddCDATA(mediainfo.Overview)
	
	xoutline := doc.AddNode(root, "outline")
	xoutline.AddCDATA(mediainfo.Overview)
	
	// 导演
	for _, director := range mediainfo.Directors {
		doc.AddNode(root, "director", director.Name)
	}
	
	// 演员
	for _, actor := range mediainfo.Actors {
		xactor := doc.AddNode(root, "actor")
		doc.AddNode(xactor, "name", actor.Name)
		doc.AddNode(xactor, "type", "Actor")
		
		character := coalesceString(actor.Character, actor.Role)
		doc.AddNode(xactor, "role", character)
		
		if actor.Avatar != nil {
			if normal, ok := actor.Avatar["normal"]; ok {
				if normalStr, ok := normal.(string); ok {
					doc.AddNode(xactor, "thumb", normalStr)
				}
			}
		}
		
		if actor.URL != "" {
			doc.AddNode(xactor, "profile", actor.URL)
		}
	}
	
	// 评分
	doc.AddNode(root, "rating", mediainfo.VoteAverage)
	
	return doc
}

// genMovieNfoFile 生成电影的NFO描述文件
func (ds *DoubanScraper) genMovieNfoFile(mediainfo *models.MediaInfo) *dom.Document {
	// 开始生成XML
	doc := dom.NewDocument()
	root := doc.AddNode(doc, "movie")
	
	// 公共部分
	doc = ds.genCommonNfo(mediainfo, doc, root)
	
	// 标题
	doc.AddNode(root, "title", mediainfo.Title)
	
	// 年份
	doc.AddNode(root, "year", mediainfo.Year)
	
	return doc
}

// genTvNfoFile 生成电视剧的NFO描述文件
func (ds *DoubanScraper) genTvNfoFile(mediainfo *models.MediaInfo) *dom.Document {
	// 开始生成XML
	doc := dom.NewDocument()
	root := doc.AddNode(doc, "tvshow")
	
	// 公共部分
	doc = ds.genCommonNfo(mediainfo, doc, root)
	
	// 标题
	doc.AddNode(root, "title", mediainfo.Title)
	
	// 年份
	doc.AddNode(root, "year", mediainfo.Year)
	doc.AddNode(root, "season", "-1")
	doc.AddNode(root, "episode", "-1")
	
	return doc
}

// genTvSeasonNfoFile 生成电视剧季的NFO描述文件
func (ds *DoubanScraper) genTvSeasonNfoFile(mediainfo *models.MediaInfo, season int) *dom.Document {
	doc := dom.NewDocument()
	root := doc.AddNode(doc, "season")
	
	// 简�?	xplot := doc.AddNode(root, "plot")
	xplot.AddCDATA(mediainfo.Overview)
	
	xoutline := doc.AddNode(root, "outline")
	xoutline.AddCDATA(mediainfo.Overview)
	
	// 标题
	doc.AddNode(root, "title", "�?"+string(rune(season+'0')))
	
	// 发行日期
	doc.AddNode(root, "premiered", mediainfo.ReleaseDate)
	doc.AddNode(root, "releasedate", mediainfo.ReleaseDate)
	
	// 发行年份
	if mediainfo.ReleaseDate != "" && len(mediainfo.ReleaseDate) >= 4 {
		doc.AddNode(root, "year", mediainfo.ReleaseDate[:4])
	}
	
	// seasonnumber
	doc.AddNode(root, "seasonnumber", string(rune(season+'0')))
	
	return doc
}
