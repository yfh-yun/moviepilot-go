package chain

import (
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"regexp"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/site"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
)

// SiteChain 站点管理处理链
type SiteChain struct {
	logger        *utils.Logger
	siteRepo      *repository.SiteRepository
	userDataRepo  *repository.SiteUserDataRepository
	statisticRepo *repository.SiteStatisticRepository
	iconRepo      *repository.SiteIconRepository
	siteService   *site.SiteService

	// 特殊站点登录验证
	specialSiteTest map[string]func(site *models.Site) (bool, string)
}

// NewSiteChain 创建站点业务链实例
func NewSiteChain(
	logger *utils.Logger,
	siteRepo *repository.SiteRepository,
	userDataRepo *repository.SiteUserDataRepository,
	statisticRepo *repository.SiteStatisticRepository,
	iconRepo *repository.SiteIconRepository,
	siteService *service.SiteService,
) *SiteChain {
	chain := &SiteChain{
		logger:        logger,
		siteRepo:      siteRepo,
		userDataRepo:  userDataRepo,
		statisticRepo: statisticRepo,
		iconRepo:      iconRepo,
		siteService:   siteService,
	}

	chain.initSpecialSiteTests()
	return chain
}

// initSpecialSiteTests 初始化特殊站点测试函数
func (s *SiteChain) initSpecialSiteTests() {
	s.specialSiteTest = map[string]func(site *models.Site) (bool, string){
		"zhuque.in":      s.zhuqueTest,
		"m-team.io":      s.mteamTest,
		"m-team.cc":      s.mteamTest,
		"ptlsp.com":      s.indexphpTest,
		"1ptba.com":      s.indexphpTest,
		"star-space.net": s.indexphpTest,
		"yemapt.org":     s.yemaTest,
		"hddolby.com":    s.hddolbyTest,
	}
}

// RefreshUserData 刷新站点的用户数据
func (s *SiteChain) RefreshUserData(site *models.Site) (*models.SiteUserData, error) {
	// 调用站点服务获取用户数据
	userData, err := s.siteService.RefreshSiteUserData(site)
	if err != nil {
		return nil, err
	}

	if userData != nil {
		// 更新用户数据
		err = s.userDataRepo.UpdateByDomainAndName(site.Domain, site.Name, userData)
		if err != nil {
			s.logger.Error("更新站点用户数据失败", "domain", site.Domain, "error", err)
		} else {
			s.logger.Info("站点用户数据更新成功", "domain", site.Domain, "site_name", site.Name)
		}

		// 处理站点消息通知
		s.processSiteMessages(site, userData)

		// 处理低分享率警告
		s.processLowRatioWarning(site, userData)
	}

	return userData, nil
}

// RefreshUserDatas 刷新所有站点的用户数据
func (s *SiteChain) RefreshUserDatas() (map[string]*models.SiteUserData, error) {
	result := make(map[string]*models.SiteUserData)

	// 获取所有活跃站点
	sites, err := s.siteRepo.GetActiveSites()
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		if s.isSystemStopped() {
			return nil, fmt.Errorf("系统已停止，刷新被中断")
		}

		userData, err := s.RefreshUserData(site)
		if err != nil {
			s.logger.Error("刷新站点用户数据失败", "site_name", site.Name, "error", err)
			continue
		}

		if userData != nil {
			result[site.Name] = userData
		}
	}

	return result, nil
}

// IsSpecialSite 判断是否特殊站点
func (s *SiteChain) IsSpecialSite(domain string) bool {
	_, exists := s.specialSiteTest[domain]
	return exists
}

// SyncCookies 通过CookieCloud同步站点Cookie
func (s *SiteChain) SyncCookies(manual bool) (bool, string) {
	s.logger.Info("开始同步CookieCloud站点...")

	// TODO: 实现CookieCloud集成
	// cookies, msg := s.cookieCloudHelper.Download()
	// if cookies == nil {
	//     s.logger.Error("CookieCloud同步失败", "error", msg)
	//     if manual {
	//         // 发送系统消息
	//     }
	//     return false, msg
	// }

	// 模拟同步逻辑
	updateCount := 0
	addCount := 0
	failCount := 0

	// TODO: 实现完整的CookieCloud同步逻辑

	retMsg := fmt.Sprintf("更新了%d个站点，新增了%d个站点", updateCount, addCount)
	if failCount > 0 {
		retMsg += fmt.Sprintf("，%d个站点添加失败，下次同步时将重试，也可以手动添加", failCount)
	}

	s.logger.Info("CookieCloud同步成功", "message", retMsg)
	return true, retMsg
}

// TestSite 测试站点是否可用
func (s *SiteChain) TestSite(domain string) (bool, string) {
	// 获取站点信息
	siteInfo, err := s.siteRepo.GetByDomain(domain)
	if err != nil {
		return false, fmt.Sprintf("站点【%s】不存在", domain)
	}

	startTime := time.Now()

	// 特殊站点测试
	if s.IsSpecialSite(domain) {
		state, message := s.specialSiteTest[domain](siteInfo)
		s.recordSiteTestResult(domain, state, time.Since(startTime))
		return state, message
	}

	// 通用站点测试
	state, message := s.genericSiteTest(siteInfo)
	s.recordSiteTestResult(domain, state, time.Since(startTime))
	return state, message
}

// genericSiteTest 通用站点测试
func (s *SiteChain) genericSiteTest(siteInfo *models.Site) (bool, string) {
	// TODO: 实现通用站点测试逻辑
	// 包括渲染页面测试和直接HTTP请求测试

	// 模拟测试逻辑
	return true, "连接成功"
}

// recordSiteTestResult 记录站点测试结果
func (s *SiteChain) recordSiteTestResult(domain string, success bool, duration time.Duration) {
	if success {
		s.siteRepo.RecordSuccess(domain, int(duration.Seconds()))
	} else {
		s.siteRepo.RecordFailure(domain)
	}
}

// processSiteMessages 处理站点消息通知
func (s *SiteChain) processSiteMessages(site *models.Site, userData *models.SiteUserData) {
	if userData.MessageUnread > 0 {
		if len(userData.MessageUnreadContents) > 0 {
			for _, content := range userData.MessageUnreadContents {
				s.sendSiteMessageNotification(site, content.Head, content.Date, content.Content)
			}
		} else {
			s.sendGenericSiteMessageNotification(site, userData.MessageUnread)
		}
	}
}

// processLowRatioWarning 处理低分享率警告
func (s *SiteChain) processLowRatioWarning(site *models.Site, userData *models.SiteUserData) {
	if userData.Ratio < 1.0 && !s.isVIPUser(userData.UserLevel) {
		s.sendLowRatioWarning(site, userData.Ratio)
	}
}

// isVIPUser 判断是否为VIP用户
func (s *SiteChain) isVIPUser(userLevel string) bool {
	vipPattern := regexp.MustCompile(`(?i)(贵宾|VIP)`)
	return vipPattern.MatchString(userLevel)
}

// sendSiteMessageNotification 发送站点消息通知
func (s *SiteChain) sendSiteMessageNotification(site *models.Site, head, date, content string) {
	// TODO: 实现消息通知系统
	msgTitle := fmt.Sprintf("【站点 %s 消息】", site.Name)
	msgText := fmt.Sprintf("时间：%s\\n标题：%s\\n内容：\\n%s", date, head, content)
	s.logger.Info("站点消息通知", "title", msgTitle, "text", msgText)
}

// sendGenericSiteMessageNotification 发送通用站点消息通知
func (s *SiteChain) sendGenericSiteMessageNotification(site *models.Site, messageCount int) {
	// TODO: 实现消息通知系统
	msgTitle := fmt.Sprintf("站点 %s 收到 %d 条新消息，请登陆查看", site.Name, messageCount)
	s.logger.Info("站点消息通知", "title", msgTitle)
}

// sendLowRatioWarning 发送低分享率警告
func (s *SiteChain) sendLowRatioWarning(site *models.Site, ratio float64) {
	// TODO: 实现消息通知系统
	msgTitle := "【站点分享率低预警】"
	msgText := fmt.Sprintf("站点 %s 分享率 %.2f，请注意！", site.Name, ratio)
	s.logger.Warn("低分享率警告", "title", msgTitle, "text", msgText)
}

// isSystemStopped 检查系统是否已停止
func (s *SiteChain) isSystemStopped() bool {
	// TODO: 实现系统状态检查
	return false
}

// 特殊站点测试方法实现

func (s *SiteChain) zhuqueTest(site *models.Site) (bool, string) {
	// TODO: 实现朱雀站点测试逻辑
	return true, "连接成功"
}

func (s *SiteChain) mteamTest(site *models.Site) (bool, string) {
	// TODO: 实现M-Team站点测试逻辑
	return true, "连接成功"
}

func (s *SiteChain) yemaTest(site *models.Site) (bool, string) {
	// TODO: 实现野马站点测试逻辑
	return true, "连接成功"
}

func (s *SiteChain) indexphpTest(site *models.Site) (bool, string) {
	// TODO: 实现index.php站点测试逻辑
	return s.genericSiteTest(site)
}

func (s *SiteChain) hddolbyTest(site *models.Site) (bool, string) {
	// TODO: 实现HDDolby站点测试逻辑
	return true, "连接成功"
}

// RemoteList 查询所有站点并发送消息
func (s *SiteChain) RemoteList(channel string, userID interface{}, source string) {
	// TODO: 实现远程站点列表查询
}

// RemoteDisable 远程禁用站点
func (s *SiteChain) RemoteDisable(argStr string, channel string, userID interface{}, source string) {
	// TODO: 实现远程禁用站点
}

// RemoteEnable 远程启用站点
func (s *SiteChain) RemoteEnable(argStr string, channel string, userID interface{}, source string) {
	// TODO: 实现远程启用站点
}

// RemoteCookie 远程更新站点Cookie
func (s *SiteChain) RemoteCookie(argStr string, channel string, userID interface{}, source string) {
	// TODO: 实现远程更新站点Cookie
}

// RemoteRefreshUserDatas 远程刷新所有站点用户数据
func (s *SiteChain) RemoteRefreshUserDatas(channel string, userID interface{}, source string) {
	// TODO: 实现远程刷新站点用户数据
}
