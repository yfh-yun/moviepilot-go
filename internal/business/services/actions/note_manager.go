// Package actions 备注管理相关业务逻辑
package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
)

// NoteAction 备注管理动作
type NoteAction struct {
	noteRepo  repository.NoteRepository
	userRepo  repository.UserRepository
	mediaRepo repository.MediaRepository
	logger    logger.Logger
}

// NewNoteAction 创建备注管理动作实例
func NewNoteAction(
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	mediaRepo repository.MediaRepository,
	logger logger.Logger,
) *NoteAction {
	return &NoteAction{
		noteRepo:  noteRepo,
		userRepo:  userRepo,
		mediaRepo: mediaRepo,
		logger:    logger,
	}
}

// Execute 执行备注动作
func (a *NoteAction) Execute(ctx context.Context, req *NoteActionRequest) (*NoteActionResponse, error) {
	a.logger.Info("开始执行备注动作",
		logger.String("action", req.Action),
		logger.String("user_id", req.UserID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
	)

	var result interface{}
	var err error

	// 根据动作类型执行相应操作
	switch req.Action {
	case "create":
		result, err = a.createNote(ctx, req)
	case "update":
		result, err = a.updateNote(ctx, req)
	case "delete":
		result, err = a.deleteNote(ctx, req)
	case "list":
		result, err = a.listNotes(ctx, req)
	case "get":
		result, err = a.getNote(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的备注动作: %s", req.Action)
	}

	if err != nil {
		a.logger.Error("备注动作执行失败", 
			logger.String("action", req.Action),
			logger.Error(err))
		return nil, err
	}

	response := &NoteActionResponse{
		Success:   true,
		Action:    req.Action,
		Data:      result,
		Message:   fmt.Sprintf("备注%s动作执行完成", req.Action),
		ProcessedAt: time.Now(),
	}

	a.logger.Info("备注动作执行完成",
		logger.String("action", req.Action),
		logger.String("user_id", req.UserID),
	)

	return response, nil
}

// NoteActionRequest 备注动作请求
type NoteActionRequest struct {
	Action     string                 `json:"action" validate:"required"`
	UserID     string                 `json:"user_id" validate:"required"`
	TargetType string                 `json:"target_type" validate:"required"` // media, download, torrent, etc.
	TargetID   string                 `json:"target_id" validate:"required"`
	Note       *repository.Note       `json:"note,omitempty"`
	NoteID     string                 `json:"note_id,omitempty"`
	Page       int                    `json:"page,omitempty"`
	PageSize   int                    `json:"page_size,omitempty"`
}

// NoteActionResponse 备注动作响应
type NoteActionResponse struct {
	Success     bool                   `json:"success"`
	Action      string                 `json:"action"`
	Data        interface{}            `json:"data"`
	Message     string                 `json:"message"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// createNote 创建备注
func (a *NoteAction) createNote(ctx context.Context, req *NoteActionRequest) (*repository.Note, error) {
	if req.Note == nil {
		return nil, fmt.Errorf("创建备注需要提供备注信息")
	}

	// 验证用户权限
	if err := a.validateUserPermission(ctx, req.UserID, "note_create"); err != nil {
		return nil, fmt.Errorf("用户权限验证失败: %w", err)
	}

	// 验证目标对象是否存在
	if err := a.validateTargetExists(ctx, req.TargetType, req.TargetID); err != nil {
		return nil, fmt.Errorf("目标对象验证失败: %w", err)
	}

	// 设置备注信息
	req.Note.UserID = req.UserID
	req.Note.TargetType = req.TargetType
	req.Note.TargetID = req.TargetID
	req.Note.CreatedAt = time.Now()
	req.Note.UpdatedAt = time.Now()

	// 创建备注
	if err := a.noteRepo.Create(ctx, req.Note); err != nil {
		return nil, fmt.Errorf("创建备注失败: %w", err)
	}

	a.logger.Info("备注创建成功",
		logger.String("note_id", req.Note.ID),
		logger.String("target_type", req.TargetType),
		logger.String("target_id", req.TargetID),
	)

	return req.Note, nil
}

// updateNote 更新备注
func (a *NoteAction) updateNote(ctx context.Context, req *NoteActionRequest) (*repository.Note, error) {
	if req.NoteID == "" {
		return nil, fmt.Errorf("更新备注需要提供备注ID")
	}

	// 获取现有备注
	existingNote, err := a.noteRepo.GetByID(ctx, req.NoteID)
	if err != nil {
		return nil, fmt.Errorf("获取备注失败: %w", err)
	}

	// 验证权限
	if !a.canModifyNote(ctx, existingNote, req.UserID) {
		return nil, fmt.Errorf("无权限修改此备注")
	}

	// 更新备注信息
	if req.Note != nil {
		if req.Note.Title != "" {
			existingNote.Title = req.Note.Title
		}
		if req.Note.Content != "" {
			existingNote.Content = req.Note.Content
		}
		if req.Note.Priority != "" {
			existingNote.Priority = req.Note.Priority
		}
		if req.Note.Tags != nil {
			existingNote.Tags = req.Note.Tags
		}
	}

	existingNote.UpdatedAt = time.Now()

	// 保存更新
	if err := a.noteRepo.Update(ctx, existingNote); err != nil {
		return nil, fmt.Errorf("更新备注失败: %w", err)
	}

	a.logger.Info("备注更新成功",
		logger.String("note_id", existingNote.ID),
		logger.String("user_id", req.UserID),
	)

	return existingNote, nil
}

// deleteNote 删除备注
func (a *NoteAction) deleteNote(ctx context.Context, req *NoteActionRequest) (bool, error) {
	if req.NoteID == "" {
		return false, fmt.Errorf("删除备注需要提供备注ID")
	}

	// 获取现有备注
	existingNote, err := a.noteRepo.GetByID(ctx, req.NoteID)
	if err != nil {
		return false, fmt.Errorf("获取备注失败: %w", err)
	}

	// 验证权限
	if !a.canModifyNote(ctx, existingNote, req.UserID) {
		return false, fmt.Errorf("无权限删除此备注")
	}

	// 删除备注
	if err := a.noteRepo.Delete(ctx, req.NoteID); err != nil {
		return false, fmt.Errorf("删除备注失败: %w", err)
	}

	a.logger.Info("备注删除成功",
		logger.String("note_id", req.NoteID),
		logger.String("user_id", req.UserID),
	)

	return true, nil
}

// listNotes 列出备注
func (a *NoteAction) listNotes(ctx context.Context, req *NoteActionRequest) (*NoteListResult, error) {
	// 设置分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 构建查询条件
	filter := &NoteFilter{
		UserID:     req.UserID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Page:       page,
		PageSize:   pageSize,
	}

	// 查询备注列表
	notes, total, err := a.noteRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询备注列表失败: %w", err)
	}

	result := &NoteListResult{
		Notes:      notes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}

	return result, nil
}

// getNote 获取备注详情
func (a *NoteAction) getNote(ctx context.Context, req *NoteActionRequest) (*repository.Note, error) {
	if req.NoteID == "" {
		return nil, fmt.Errorf("获取备注需要提供备注ID")
	}

	// 获取备注
	note, err := a.noteRepo.GetByID(ctx, req.NoteID)
	if err != nil {
		return nil, fmt.Errorf("获取备注失败: %w", err)
	}

	// 验证查看权限
	if !a.canViewNote(ctx, note, req.UserID) {
		return nil, fmt.Errorf("无权限查看此备注")
	}

	return note, nil
}

// validateUserPermission 验证用户权限
func (a *NoteAction) validateUserPermission(ctx context.Context, userID, permission string) error {
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if !user.IsActive() {
		return fmt.Errorf("用户已被禁用")
	}

	if !user.HasPermission(permission) {
		return fmt.Errorf("用户无%s权限", permission)
	}

	return nil
}

// validateTargetExists 验证目标对象是否存在
func (a *NoteAction) validateTargetExists(ctx context.Context, targetType, targetID string) error {
	switch targetType {
	case "media":
		_, err := a.mediaRepo.GetByID(ctx, targetID)
		if err != nil {
			return fmt.Errorf("媒体对象不存在: %w", err)
		}
	// 可以添加其他目标类型的验证
	default:
		// 对于其他类型，暂时不验证
	}

	return nil
}

// canModifyNote 是否可以修改备注
func (a *NoteAction) canModifyNote(ctx context.Context, note *repository.Note, userID string) bool {
	// 备注创建者可以修改
	if note.UserID == userID {
		return true
	}

	// 管理员可以修改
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	return user.IsAdmin()
}

// canViewNote 是否可以查看备注
func (a *NoteAction) canViewNote(ctx context.Context, note *repository.Note, userID string) bool {
	// 备注创建者可以查看
	if note.UserID == userID {
		return true
	}

	// 管理员可以查看
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	return user.IsAdmin()
}

// NoteFilter 备注查询过滤器
type NoteFilter struct {
	UserID     string `json:"user_id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// NoteListResult 备注列表结果
type NoteListResult struct {
	Notes      []repository.Note `json:"notes"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int64             `json:"total_pages"`
}