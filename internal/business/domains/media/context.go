package media

// Context 上下文聚合（meta_info + media_info + torrent_info）
type Context struct {
	MetaInfo                MetaBase // 接口或具体 struct，来自 meta 子系统
	MediaInfo               *MediaInfo
	TorrentInfo             *TorrentInfo
	MediaRecognizeFailCount int
}
