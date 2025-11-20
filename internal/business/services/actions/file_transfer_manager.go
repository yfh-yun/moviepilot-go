// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// FileTransferManager 文件传输管理器
// 提供高效、可靠的文件上传、下载和传输管理
type FileTransferManager struct {
	storageService  StorageService
	securityManager SecurityManager
	logger          *zap.Logger
	transfers       map[string]*TransferSession
	mutex           sync.RWMutex
}

// StorageService 存储服务接口
type StorageService interface {
	UploadFile(ctx context.Context, file *model.File) error
	DownloadFile(ctx context.Context, fileID string) (*model.File, error)
	DeleteFile(ctx context.Context, fileID string) error
	GetFileInfo(ctx context.Context, fileID string) (*model.FileInfo, error)
}

// SecurityManager 安全管理器接口
type SecurityManager interface {
	ValidateFileAccess(ctx context.Context, fileID, userID string) error
	ScanFileForThreats(ctx context.Context, filePath string) error
	GetFileQuota(ctx context.Context, userID string) (*FileQuota, error)
}

// NewFileTransferManager 创建文件传输管理器实例
func NewFileTransferManager(
	storageService StorageService,
	securityManager SecurityManager,
) *FileTransferManager {
	return &FileTransferManager{
		storageService:  storageService,
		securityManager: securityManager,
		logger:          logger.NewLogger("file_transfer_manager"),
		transfers:       make(map[string]*TransferSession),
	}
}

// UploadFile 上传文件
func (m *FileTransferManager) UploadFile(ctx context.Context, request *UploadRequest) (*UploadResponse, error) {
	m.logger.Info("开始文件上传",
		zap.String("filename", request.FileHeader.Filename),
		zap.String("user_id", request.UserID))

	// 1. 安全检查
	if err := m.securityManager.ValidateFileAccess(ctx, "", request.UserID); err != nil {
		return nil, fmt.Errorf("文件访问安全检查失败: %w", err)
	}

	// 2. 配额检查
	quota, err := m.securityManager.GetFileQuota(ctx, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户配额失败: %w", err)
	}

	if request.FileHeader.Size > quota.MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制: %d > %d", request.FileHeader.Size, quota.MaxFileSize)
	}

	// 3. 创建传输会话
	sessionID := generateSessionID()
	session := &TransferSession{
		ID:        sessionID,
		UserID:    request.UserID,
		FileName:  request.FileHeader.Filename,
		FileSize:  request.FileHeader.Size,
		Status:    "uploading",
		StartedAt: time.Now(),
		progress:  make(chan int64),
		error:     make(chan error),
		cancel:    make(chan struct{}),
	}

	m.registerSession(session)

	// 4. 异步上传
	go m.handleUpload(ctx, session, request.File)

	response := &UploadResponse{
		SessionID: sessionID,
		FileName:  request.FileHeader.Filename,
		FileSize:  request.FileHeader.Size,
		Status:    "uploading",
		StartedAt: session.StartedAt,
	}

	m.logger.Info("文件上传会话创建成功",
		zap.String("session_id", sessionID),
		zap.String("filename", request.FileHeader.Filename))

	return response, nil
}

// handleUpload 处理文件上传
func (m *FileTransferManager) handleUpload(ctx context.Context, session *TransferSession, file multipart.File) {
	defer file.Close()
	defer m.unregisterSession(session.ID)

	// 创建临时文件
	tempFile, err := m.createTempFile(session.FileName)
	if err != nil {
		session.error <- err
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 复制文件内容并监控进度
	written, err := m.copyWithProgress(tempFile, file, session.FileSize, session.progress, session.cancel)
	if err != nil {
		session.error <- err
		return
	}

	// 安全检查
	if err := m.securityManager.ScanFileForThreats(ctx, tempFile.Name()); err != nil {
		session.error <- fmt.Errorf("文件安全检查失败: %w", err)
		return
	}

	// 保存文件到存储服务
	fileModel := &model.File{
		ID:        generateFileID(),
		Name:      session.FileName,
		Size:      session.FileSize,
		Path:      tempFile.Name(),
		UserID:    session.UserID,
		CreatedAt: time.Now(),
	}

	if err := m.storageService.UploadFile(ctx, fileModel); err != nil {
		session.error <- fmt.Errorf("文件存储失败: %w", err)
		return
	}

	session.Status = "completed"
	session.CompletedAt = time.Now()
	session.BytesTransferred = written

	m.logger.Info("文件上传完成",
		zap.String("session_id", session.ID),
		zap.String("filename", session.FileName),
		zap.Int64("bytes_written", written),
		zap.Duration("duration", time.Since(session.StartedAt)))
}

// DownloadFile 下载文件
func (m *FileTransferManager) DownloadFile(ctx context.Context, request *DownloadRequest) (*DownloadResponse, error) {
	m.logger.Info("开始文件下载",
		zap.String("file_id", request.FileID),
		zap.String("user_id", request.UserID))

	// 1. 安全检查
	if err := m.securityManager.ValidateFileAccess(ctx, request.FileID, request.UserID); err != nil {
		return nil, fmt.Errorf("文件访问安全检查失败: %w", err)
	}

	// 2. 获取文件信息
	fileInfo, err := m.storageService.GetFileInfo(ctx, request.FileID)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 3. 创建下载会话
	sessionID := generateSessionID()
	session := &TransferSession{
		ID:        sessionID,
		UserID:    request.UserID,
		FileName:  fileInfo.Name,
		FileSize:  fileInfo.Size,
		Status:    "downloading",
		StartedAt: time.Now(),
		progress:  make(chan int64),
		error:     make(chan error),
		cancel:    make(chan struct{}),
	}

	m.registerSession(session)

	// 4. 异步下载
	go m.handleDownload(ctx, session, request.FileID)

	response := &DownloadResponse{
		SessionID: sessionID,
		FileName:  fileInfo.Name,
		FileSize:  fileInfo.Size,
		Status:    "downloading",
		StartedAt: session.StartedAt,
	}

	m.logger.Info("文件下载会话创建成功",
		zap.String("session_id", sessionID),
		zap.String("filename", fileInfo.Name))

	return response, nil
}

// handleDownload 处理文件下载
func (m *FileTransferManager) handleDownload(ctx context.Context, session *TransferSession, fileID string) {
	defer m.unregisterSession(session.ID)

	// 获取文件
	file, err := m.storageService.DownloadFile(ctx, fileID)
	if err != nil {
		session.error <- err
		return
	}

	session.Status = "completed"
	session.CompletedAt = time.Now()
	session.BytesTransferred = file.Size

	m.logger.Info("文件下载完成",
		zap.String("session_id", session.ID),
		zap.String("filename", session.FileName),
		zap.Int64("bytes_transferred", file.Size),
		zap.Duration("duration", time.Since(session.StartedAt)))
}

// GetTransferStatus 获取传输状态
func (m *FileTransferManager) GetTransferStatus(ctx context.Context, sessionID string) (*TransferStatus, error) {
	session, exists := m.getSession(sessionID)
	if !exists {
		return nil, fmt.Errorf("传输会话不存在: %s", sessionID)
	}

	status := &TransferStatus{
		SessionID:        session.ID,
		FileName:         session.FileName,
		FileSize:         session.FileSize,
		Status:           session.Status,
		BytesTransferred: session.BytesTransferred,
		Progress:         float64(session.BytesTransferred) / float64(session.FileSize) * 100,
		StartedAt:        session.StartedAt,
		CompletedAt:      session.CompletedAt,
	}

	return status, nil
}

// CancelTransfer 取消传输
func (m *FileTransferManager) CancelTransfer(ctx context.Context, sessionID string) error {
	session, exists := m.getSession(sessionID)
	if !exists {
		return fmt.Errorf("传输会话不存在: %s", sessionID)
	}

	if session.Status == "completed" || session.Status == "cancelled" {
		return fmt.Errorf("传输已结束，无法取消")
	}

	close(session.cancel)
	session.Status = "cancelled"
	session.CompletedAt = time.Now()

	m.logger.Info("文件传输已取消",
		zap.String("session_id", sessionID),
		zap.String("filename", session.FileName))

	return nil
}

// copyWithProgress 带进度监控的文件复制
func (m *FileTransferManager) copyWithProgress(dst io.Writer, src io.Reader, totalSize int64, progress chan<- int64, cancel <-chan struct{}) (int64, error) {
	var written int64
	buffer := make([]byte, 32*1024) // 32KB缓冲区

	for {
		select {
		case <-cancel:
			return written, fmt.Errorf("传输被取消")
		default:
			nr, err := src.Read(buffer)
			if nr > 0 {
				nw, err := dst.Write(buffer[:nr])
				if err != nil {
					return written, err
				}
				if nw != nr {
					return written, io.ErrShortWrite
				}
				written += int64(nw)

				// 发送进度更新
				select {
				case progress <- written:
				default:
					// 进度通道已满，跳过此次更新
				}
			}
			if err != nil {
				if err == io.EOF {
					return written, nil
				}
				return written, err
			}
		}
	}
}

// createTempFile 创建临时文件
func (m *FileTransferManager) createTempFile(filename string) (*os.File, error) {
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "moviepilot_upload_*"+filepath.Ext(filename))
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	return tempFile, nil
}

// registerSession 注册传输会话
func (m *FileTransferManager) registerSession(session *TransferSession) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.transfers[session.ID] = session
}

// unregisterSession 注销传输会话
func (m *FileTransferManager) unregisterSession(sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.transfers, sessionID)
}

// getSession 获取传输会话
func (m *FileTransferManager) getSession(sessionID string) (*TransferSession, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, exists := m.transfers[sessionID]
	return session, exists
}

// TransferSession 传输会话
type TransferSession struct {
	ID               string
	UserID           string
	FileName         string
	FileSize         int64
	Status           string
	BytesTransferred int64
	StartedAt        time.Time
	CompletedAt      time.Time
	progress         chan int64
	error            chan error
	cancel           chan struct{}
}

// UploadRequest 上传请求
type UploadRequest struct {
	FileHeader *multipart.FileHeader
	File       multipart.File
	UserID     string
	Metadata   map[string]string
}

// UploadResponse 上传响应
type UploadResponse struct {
	SessionID string    `json:"session_id"`
	FileName  string    `json:"filename"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

// DownloadResponse 下载响应
type DownloadResponse struct {
	SessionID string    `json:"session_id"`
	FileName  string    `json:"filename"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

// TransferStatus 传输状态
type TransferStatus struct {
	SessionID        string    `json:"session_id"`
	FileName         string    `json:"filename"`
	FileSize         int64     `json:"file_size"`
	Status           string    `json:"status"`
	BytesTransferred int64     `json:"bytes_transferred"`
	Progress         float64   `json:"progress"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
}

// FileQuota 文件配额
type FileQuota struct {
	MaxFileSize    int64 `json:"max_file_size"`
	MaxStorageSize int64 `json:"max_storage_size"`
	UsedStorage    int64 `json:"used_storage"`
}

// 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// 生成文件ID
func generateFileID() string {
	return fmt.Sprintf("file_%d", time.Now().UnixNano())
}
