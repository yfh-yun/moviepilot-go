package media

// Context 上下文聚合（meta_info + media_info + torrent_info）
type Context struct {
	MetaInfo                *MetaBase // 改为指针类型，便于判断是否为空
	MediaInfo               *MediaInfo
	TorrentInfo             *TorrentInfo
	MediaRecognizeFailCount int
}

// ToDict 将Context转换为字典，与Python版本to_dict功能一致
func (c *Context) ToDict() map[string]interface{} {
	result := make(map[string]interface{})
	
	// 元信息
	if c.MetaInfo != nil {
		result["meta_info"] = c.MetaInfo.ToDict()
	}
	
	// 媒体信息
	if c.MediaInfo != nil {
		result["media_info"] = c.MediaInfo.ToDict()
	}
	
	// 种子信息
	if c.TorrentInfo != nil {
		result["torrent_info"] = c.TorrentInfo.ToDict()
	}
	
	// 识别失败次数
	result["media_recognize_fail_count"] = c.MediaRecognizeFailCount
	
	return result
}
