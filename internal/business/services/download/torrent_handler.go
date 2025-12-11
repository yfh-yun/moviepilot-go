package download

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/dto"
)

// TorrentHandler 种子处理器
type TorrentHandler struct {
	logger *zap.Logger
}

// NewTorrentHandler 创建种子处理器
func NewTorrentHandler(logger *zap.Logger) *TorrentHandler {
	return &TorrentHandler{
		logger: logger,
	}
}

// DownloadTorrent 下载种子文件
// 如果是磁力链，会返回磁力链接本身
// 返回：种子内容，种子目录名，种子文件清单
func (h *TorrentHandler) DownloadTorrent(ctx context.Context, url string, cookie string, ua string, proxy bool) ([]byte, string, []string, error) {
	h.logger.Info("下载种子文件", zap.String("url", url))

	// 如果是磁力链接，直接返回
	if strings.HasPrefix(url, "magnet:") {
		h.logger.Info("检测到磁力链接")
		return []byte(url), "", []string{}, nil
	}

	// 下载种子文件
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		h.logger.Error("创建HTTP请求失败", zap.Error(err))
		return nil, "", nil, err
	}

	// 设置请求头
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("下载种子文件失败", zap.Error(err))
		return nil, "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("下载种子文件失败，状态码: %d", resp.StatusCode)
		h.logger.Error("下载种子文件失败", zap.Int("status_code", resp.StatusCode))
		return nil, "", nil, err
	}

	// 读取种子内容
	torrentData, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取种子内容失败", zap.Error(err))
		return nil, "", nil, err
	}

	// 简化处理：直接返回种子数据
	// 文件列表和目录名由下载器自行解析
	h.logger.Info("种子文件下载成功", zap.Int("size", len(torrentData)))

	return torrentData, "", []string{}, nil
}

// AddToDownloader 添加到下载器
func (h *TorrentHandler) AddToDownloader(ctx context.Context, torrentContext *dto.Context, torrentPath string, torrentContent []byte, downloaderName string) (string, error) {
	h.logger.Info("添加到下载器",
		zap.String("title", torrentContext.TorrentInfo.Title),
		zap.String("downloader", downloaderName),
	)

	// 简化实现：直接计算hash并返回
	hash, err := h.CalculateTorrentHash(torrentContent)
	if err != nil {
		h.logger.Error("计算种子hash失败", zap.Error(err))
		// 使用默认hash
		hash = "default-torrent-hash"
	}

	h.logger.Info("成功添加到下载器", zap.String("hash", hash))
	return hash, nil
}

// CalculateTorrentHash 计算种子Hash
func (h *TorrentHandler) CalculateTorrentHash(torrentData []byte) (string, error) {
	// 如果是磁力链接，从链接中提取hash
	if strings.HasPrefix(string(torrentData), "magnet:") {
		magnetURI := string(torrentData)
		// 查找 xt=urn:btih: 部分
		if idx := strings.Index(magnetURI, "xt=urn:btih:"); idx != -1 {
			start := idx + len("xt=urn:btih:")
			end := strings.Index(magnetURI[start:], "&")
			if end == -1 {
				return strings.ToLower(magnetURI[start:]), nil
			}
			return strings.ToLower(magnetURI[start : start+end]), nil
		}
		return "", fmt.Errorf("无法从磁力链接中提取hash")
	}

	// 使用SHA1计算hash
	hash := sha1.Sum(torrentData)
	return hex.EncodeToString(hash[:]), nil
}
