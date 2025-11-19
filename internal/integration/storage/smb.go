package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/hirochachacha/go-smb2"
)

// SMBStorage SMB存储提供商
type SMBStorage struct {
	name       string
	server     string
	share      string
	username   string
	password   string
	domain     string
	connected  bool
	smbSession *smb2.Session
	smbShare   *smb2.Share
	mu         sync.RWMutex
}

// NewSMBStorage 创建SMB存储实例
func NewSMBStorage(name, server, share, username, password, domain string) *SMBStorage {
	return &SMBStorage{
		name:     name,
		server:   server,
		share:    share,
		username: username,
		password: password,
		domain:   domain,
	}
}

// Name 返回存储名称
func (s *SMBStorage) Name() string {
	return s.name
}

// Type 返回存储类型
func (s *SMBStorage) Type() string {
	return ProviderSMB
}

// IsConnected 检查是否连接
func (s *SMBStorage) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// Connect 连接SMB共享
func (s *SMBStorage) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil
	}

	// 建立SMB连接
	conn, err := net.Dial("tcp", s.server+":445")
	if err != nil {
		return fmt.Errorf("连接SMB服务器失败: %w", err)
	}

	// 创建SMB会话
	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     s.username,
			Password: s.password,
			Domain:   s.domain,
		},
	}

	session, err := dialer.Dial(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("建立SMB会话失败: %w", err)
	}

	// 挂载共享
	share, err := session.Mount(s.share)
	if err != nil {
		session.Logoff()
		conn.Close()
		return fmt.Errorf("挂载SMB共享失败: %w", err)
	}

	s.smbSession = session
	s.smbShare = share
	s.connected = true

	return nil
}

// Disconnect 断开连接
func (s *SMBStorage) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil
	}

	var errors []string

	if s.smbShare != nil {
		if err := s.smbShare.Umount(); err != nil {
			errors = append(errors, fmt.Sprintf("卸载共享失败: %v", err))
		}
	}

	if s.smbSession != nil {
		if err := s.smbSession.Logoff(); err != nil {
			errors = append(errors, fmt.Sprintf("注销会话失败: %v", err))
		}
	}

	s.connected = false
	s.smbShare = nil
	s.smbSession = nil

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

	return nil
}

// Upload 上传文件
func (s *SMBStorage) Upload(ctx context.Context, filePath string, reader io.Reader, size int64) error {
	if !s.IsConnected() {
		return ErrNotConnected
	}

	// 创建目录
	if err := s.Mkdir(ctx, filepath.Dir(filePath)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建文件
	file, err := s.smbShare.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	written, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	if size > 0 && written != size {
		return fmt.Errorf("文件大小不匹配: 期望 %d, 实际 %d", size, written)
	}

	return nil
}

// Download 下载文件
func (s *SMBStorage) Download(ctx context.Context, filePath string) (io.ReadCloser, error) {
	if !s.IsConnected() {
		return nil, ErrNotConnected
	}

	file, err := s.smbShare.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	return file, nil
}

// 其他SMB存储实现方法...
// [Delete, Exists, Move, Copy, List, Mkdir, Rmdir, GetQuota, GetFileInfo等方法的实现]

// GetQuota 获取配额信息（SMB通常不支持）
func (s *SMBStorage) GetQuota(ctx context.Context) (*QuotaInfo, error) {
	return nil, ErrNotImplemented
}

// GetFileInfo 获取文件信息
func (s *SMBStorage) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	if !s.IsConnected() {
		return nil, ErrNotConnected
	}

	info, err := s.smbShare.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	isDir := info.IsDir()

	fileInfo := &FileInfo{
		Name:         filepath.Base(path),
		Path:         path,
		IsDir:        isDir,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		MimeType:     s.getMimeType(info.Name(), isDir),
	}

	return fileInfo, nil
}
