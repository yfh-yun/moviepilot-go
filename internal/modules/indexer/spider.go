package indexer

import (
	"moviepilot-go/internal/modules/indexer/spider"
	"moviepilot-go/pkg/models"
)

// SiteSpider 站点爬虫基础结构
type SiteSpider struct {
	impl *SiteSpiderImpl
}

// NewSiteSpider 创建站点爬虫实例
func NewSiteSpider(indexer map[string]interface{}, keyword *string, mtype models.MediaType, cat *string, page int) *SiteSpider {
	return &SiteSpider{
		impl: NewSiteSpiderImpl(indexer, keyword, mtype, cat, page, nil),
	}
}

// GetTorrents 获取种子列表
func (s *SiteSpider) GetTorrents() []map[string]interface{} {
	return s.impl.GetTorrents()
}

// SpiderTNodeSpider TNode站点爬虫
func SpiderTNodeSpider(site map[string]interface{}, keyword *string, page int) (bool, []map[string]interface{}) {
	// 实现TNodeSpider的搜索逻辑
	tnode := spider.NewTNodeSpider(site)
	if keyword != nil {
		return tnode.Search(*keyword, page)
	}
	return tnode.Search("", page)
}

// SpiderTorrentLeech TorrentLeech站点爬虫
func SpiderTorrentLeech(site map[string]interface{}, keyword *string, page int) (bool, []map[string]interface{}) {
	// 实现TorrentLeech的搜索逻辑
	torrentleech := spider.NewTorrentLeech(site)
	if keyword != nil {
		return torrentleech.Search(*keyword, page)
	}
	return torrentleech.Search("", page)
}

// SpiderMTorrent MTorrent站点爬虫
func SpiderMTorrent(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现MTorrentSpider的搜索逻辑
	mtorrent := spider.NewMTorrentSpider(site)
	if keyword != nil {
		return mtorrent.Search(*keyword, mtype, page)
	}
	return mtorrent.Search("", mtype, page)
}

// SpiderYema Yema站点爬虫
func SpiderYema(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现YemaSpider的搜索逻辑
	yema := spider.NewYemaSpider(site)
	if keyword != nil {
		return yema.Search(*keyword, mtype, page)
	}
	return yema.Search("", mtype, page)
}

// SpiderHaidan Haidan站点爬虫
func SpiderHaidan(site map[string]interface{}, keyword *string, mtype models.MediaType) (bool, []map[string]interface{}) {
	// 实现HaiDanSpider的搜索逻辑
	haidan := spider.NewHaiDanSpider(site)
	if keyword != nil {
		return haidan.Search(*keyword, mtype)
	}
	return haidan.Search("", mtype)
}

// SpiderHDDolby HDDolby站点爬虫
func SpiderHDDolby(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现HddolbySpider的搜索逻辑
	hddolby := spider.NewHddolbySpider(site)
	if keyword != nil {
		return hddolby.Search(*keyword, mtype, page)
	}
	return hddolby.Search("", mtype, page)
}

// AsyncSpiderTNodeSpider TNode站点爬虫异步版本
func AsyncSpiderTNodeSpider(site map[string]interface{}, keyword *string, page int) (bool, []map[string]interface{}) {
	// 实现TNodeSpider的异步搜索逻辑
	tnode := spider.NewTNodeSpider(site)
	if keyword != nil {
		return tnode.AsyncSearch(*keyword, page)
	}
	return tnode.AsyncSearch("", page)
}

// AsyncSpiderTorrentLeech TorrentLeech站点爬虫异步版本
func AsyncSpiderTorrentLeech(site map[string]interface{}, keyword *string, page int) (bool, []map[string]interface{}) {
	// 实现TorrentLeech的异步搜索逻辑
	torrentleech := spider.NewTorrentLeech(site)
	if keyword != nil {
		return torrentleech.AsyncSearch(*keyword, page)
	}
	return torrentleech.AsyncSearch("", page)
}

// AsyncSpiderMTorrent MTorrent站点爬虫异步版本
func AsyncSpiderMTorrent(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现MTorrentSpider的异步搜索逻辑
	mtorrent := spider.NewMTorrentSpider(site)
	if keyword != nil {
		return mtorrent.AsyncSearch(*keyword, mtype, page)
	}
	return mtorrent.AsyncSearch("", mtype, page)
}

// AsyncSpiderYema Yema站点爬虫异步版本
func AsyncSpiderYema(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现YemaSpider的异步搜索逻辑
	yema := spider.NewYemaSpider(site)
	if keyword != nil {
		return yema.AsyncSearch(*keyword, mtype, page)
	}
	return yema.AsyncSearch("", mtype, page)
}

// AsyncSpiderHaidan Haidan站点爬虫异步版本
func AsyncSpiderHaidan(site map[string]interface{}, keyword *string, mtype models.MediaType) (bool, []map[string]interface{}) {
	// 实现HaiDanSpider的异步搜索逻辑
	haidan := spider.NewHaiDanSpider(site)
	if keyword != nil {
		return haidan.AsyncSearch(*keyword, mtype)
	}
	return haidan.AsyncSearch("", mtype)
}

// AsyncSpiderHDDolby HDDolby站点爬虫异步版本
func AsyncSpiderHDDolby(site map[string]interface{}, keyword *string, mtype models.MediaType, page int) (bool, []map[string]interface{}) {
	// 实现HddolbySpider的异步搜索逻辑
	hddolby := spider.NewHddolbySpider(site)
	if keyword != nil {
		return hddolby.AsyncSearch(*keyword, mtype, page)
	}
	return hddolby.AsyncSearch("", mtype, page)
}
