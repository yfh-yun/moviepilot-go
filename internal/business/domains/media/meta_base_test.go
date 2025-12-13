package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaBase_Name(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试设置中文名
	meta.SetName("中文名称")
	assert.Equal(t, "中文名称", meta.Name())
	assert.Equal(t, "中文名称", meta.CnName)
	assert.Equal(t, "", meta.EnName)

	// 测试设置英文名
	meta.SetName("English Name")
	assert.Equal(t, "English Name", meta.Name())
	assert.Equal(t, "", meta.CnName)
	assert.Equal(t, "English Name", meta.EnName)
}

func TestMetaBase_Season(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试没有季信息的情况
	meta.Type = MediaTypeTV
	assert.Equal(t, "S01", meta.Season())
	assert.Equal(t, "", meta.Sea())
	assert.Equal(t, "1", meta.SeasonSeq())
	assert.Equal(t, []int{1}, meta.SeasonList())

	// 测试单季情况
	season1 := 1
	meta.BeginSeason = &season1
	assert.Equal(t, "S01", meta.Season())
	assert.Equal(t, "S01", meta.Sea())
	assert.Equal(t, "1", meta.SeasonSeq())
	assert.Equal(t, []int{1}, meta.SeasonList())

	// 测试多季情况
	season2 := 2
	meta.EndSeason = &season2
	assert.Equal(t, "S01-S02", meta.Season())
	assert.Equal(t, "S01-S02", meta.Sea())
	assert.Equal(t, "1", meta.SeasonSeq())
	assert.Equal(t, []int{1, 2}, meta.SeasonList())
}

func TestMetaBase_Episode(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试没有集信息的情况
	assert.Equal(t, "", meta.Episode())
	assert.Equal(t, []int{}, meta.EpisodeList())
	assert.Equal(t, "", meta.Episodes())
	assert.Equal(t, "", meta.EpisodeSeqs())
	assert.Equal(t, "", meta.EpisodeSeq())

	// 测试单集情况
	episode1 := 1
	meta.BeginEpisode = &episode1
	assert.Equal(t, "E01", meta.Episode())
	assert.Equal(t, []int{1}, meta.EpisodeList())
	assert.Equal(t, "E01", meta.Episodes())
	assert.Equal(t, "1", meta.EpisodeSeqs())
	assert.Equal(t, "1", meta.EpisodeSeq())

	// 测试多集情况
	episode3 := 3
	meta.EndEpisode = &episode3
	assert.Equal(t, "E01-E03", meta.Episode())
	assert.Equal(t, []int{1, 2, 3}, meta.EpisodeList())
	assert.Equal(t, "E01E02E03", meta.Episodes())
	assert.Equal(t, "1-3", meta.EpisodeSeqs())
	assert.Equal(t, "1", meta.EpisodeSeq())
}

func TestMetaBase_SeasonEpisode(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试电影类型
	meta.Type = MediaTypeMovie
	assert.Equal(t, "", meta.SeasonEpisode())

	// 测试电视剧类型，只有季
	meta.Type = MediaTypeTV
	season1 := 1
	meta.BeginSeason = &season1
	assert.Equal(t, "S01", meta.SeasonEpisode())

	// 测试电视剧类型，只有集
	meta.BeginSeason = nil
	episode1 := 1
	meta.BeginEpisode = &episode1
	assert.Equal(t, "E01", meta.SeasonEpisode())

	// 测试电视剧类型，既有季又有集
	meta.BeginSeason = &season1
	assert.Equal(t, "S01 E01", meta.SeasonEpisode())
}

func TestMetaBase_ResourceTerm(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试没有资源属性的情况
	assert.Equal(t, "", meta.ResourceTerm())
	assert.Equal(t, "", meta.Edition())

	// 测试有资源属性的情况
	meta.ResourceType = ResourceTypeTV
	meta.ResourceEffect = ResourceEffectBluray
	meta.ResourcePix = ResourcePix1080P
	meta.ResourceTeam = "字幕组"
	meta.VideoEncode = "H.264"
	meta.AudioEncode = "DTS"

	assert.Equal(t, "tv bluray 1080p", meta.ResourceTerm())
	assert.Equal(t, "tv bluray", meta.Edition())
	assert.Equal(t, "字幕组", meta.ReleaseGroup())
	assert.Equal(t, "H.264", meta.VideoTerm())
	assert.Equal(t, "DTS", meta.AudioTerm())
}

func TestMetaBase_IsInSeason(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试没有季信息的情况
	meta.Type = MediaTypeTV
	assert.True(t, meta.IsInSeason(1))
	assert.False(t, meta.IsInSeason(2))

	// 测试单季情况
	season1 := 1
	meta.BeginSeason = &season1
	assert.True(t, meta.IsInSeason(1))
	assert.False(t, meta.IsInSeason(2))

	// 测试多季情况
	season3 := 3
	meta.EndSeason = &season3
	assert.True(t, meta.IsInSeason(1))
	assert.True(t, meta.IsInSeason(2))
	assert.True(t, meta.IsInSeason(3))
	assert.False(t, meta.IsInSeason(4))

	// 测试列表类型
	assert.True(t, meta.IsInSeason([]int{1, 2}))
	assert.False(t, meta.IsInSeason([]int{1, 4}))

	// 测试字符串类型
	assert.True(t, meta.IsInSeason("2"))
	assert.False(t, meta.IsInSeason("4"))
}

func TestMetaBase_IsInEpisode(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试没有集信息的情况
	assert.False(t, meta.IsInEpisode(1))

	// 测试单集情况
	episode1 := 1
	meta.BeginEpisode = &episode1
	assert.True(t, meta.IsInEpisode(1))
	assert.False(t, meta.IsInEpisode(2))

	// 测试多集情况
	episode3 := 3
	meta.EndEpisode = &episode3
	assert.True(t, meta.IsInEpisode(1))
	assert.True(t, meta.IsInEpisode(2))
	assert.True(t, meta.IsInEpisode(3))
	assert.False(t, meta.IsInEpisode(4))

	// 测试列表类型
	assert.True(t, meta.IsInEpisode([]int{1, 2}))
	assert.False(t, meta.IsInEpisode([]int{1, 4}))

	// 测试字符串类型
	assert.True(t, meta.IsInEpisode("2"))
	assert.False(t, meta.IsInEpisode("4"))
}

func TestMetaBase_SetSeason(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试整数类型
	meta.SetSeason(2)
	assert.Equal(t, 2, *meta.BeginSeason)
	assert.Nil(t, meta.EndSeason)
	assert.Equal(t, 1, meta.TotalSeason)

	// 测试列表类型
	meta.SetSeason([]int{1, 3})
	assert.Equal(t, 1, *meta.BeginSeason)
	assert.Equal(t, 3, *meta.EndSeason)
	assert.Equal(t, 3, meta.TotalSeason)

	// 测试字符串类型
	meta.SetSeason("4-6")
	assert.Equal(t, 4, *meta.BeginSeason)
	assert.Equal(t, 6, *meta.EndSeason)
	assert.Equal(t, 3, meta.TotalSeason)
}

func TestMetaBase_SetEpisode(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试整数类型
	meta.SetEpisode(2)
	assert.Equal(t, 2, *meta.BeginEpisode)
	assert.Nil(t, meta.EndEpisode)
	assert.Equal(t, 1, meta.TotalEpisode)

	// 测试列表类型
	meta.SetEpisode([]int{1, 3})
	assert.Equal(t, 1, *meta.BeginEpisode)
	assert.Equal(t, 3, *meta.EndEpisode)
	assert.Equal(t, 3, meta.TotalEpisode)

	// 测试字符串类型
	meta.SetEpisode("4-6")
	assert.Equal(t, 4, *meta.BeginEpisode)
	assert.Equal(t, 6, *meta.EndEpisode)
	assert.Equal(t, 3, meta.TotalEpisode)
}

func TestMetaBase_SetEpisodes(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试设置集范围
	episode1 := 1
	episode5 := 5
	meta.SetEpisodes(&episode1, &episode5)
	assert.Equal(t, 1, *meta.BeginEpisode)
	assert.Equal(t, 5, *meta.EndEpisode)
	assert.Equal(t, 5, meta.TotalEpisode)
}

func TestMetaBase_InitSubtitle(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("", "", "", false)

	// 测试解析"Episode X"格式
	meta.InitSubtitle("Episode 1")
	assert.Equal(t, 1, *meta.BeginEpisode)
	assert.Equal(t, 1, meta.TotalEpisode)
	assert.Equal(t, MediaTypeTV, meta.Type)

	// 测试解析"全X季"格式
	meta = NewMetaBase("", "", "", false)
	meta.InitSubtitle("全2季")
	assert.Equal(t, 1, *meta.BeginSeason)
	assert.Equal(t, 2, *meta.EndSeason)
	assert.Equal(t, 2, meta.TotalSeason)
	assert.Equal(t, MediaTypeTV, meta.Type)

	// 测试解析"第X季"格式
	meta = NewMetaBase("", "", "", false)
	meta.InitSubtitle("第3季")
	assert.Equal(t, 3, *meta.BeginSeason)
	assert.Nil(t, meta.EndSeason)
	assert.Equal(t, 1, meta.TotalSeason)
	assert.Equal(t, MediaTypeTV, meta.Type)

	// 测试解析"第X-Y集"格式
	meta = NewMetaBase("", "", "", false)
	meta.InitSubtitle("第1-5集")
	assert.Equal(t, 1, *meta.BeginEpisode)
	assert.Equal(t, 5, *meta.EndEpisode)
	assert.Equal(t, 5, meta.TotalEpisode)
	assert.Equal(t, MediaTypeTV, meta.Type)

	// 测试解析"第X集"格式
	meta = NewMetaBase("", "", "", false)
	meta.InitSubtitle("第2集")
	assert.Equal(t, 2, *meta.BeginEpisode)
	assert.Nil(t, meta.EndEpisode)
	assert.Equal(t, 1, meta.TotalEpisode)
	assert.Equal(t, MediaTypeTV, meta.Type)

	// 测试解析"X集全"格式
	meta = NewMetaBase("", "", "", false)
	meta.InitSubtitle("10集全")
	assert.Equal(t, 10, meta.TotalEpisode)
	assert.Equal(t, MediaTypeTV, meta.Type)
}

func TestMetaBase_Merge(t *testing.T) {
	// 创建两个MetaBase实例
	meta1 := NewMetaBase("", "", "", false)
	meta2 := NewMetaBase("", "", "", false)

	// 设置meta2的属性
	meta2.Type = MediaTypeTV
	meta2.CnName = "中文名称"
	meta2.Year = "2023"
	season1 := 1
	meta2.BeginSeason = &season1
	episode1 := 1
	meta2.BeginEpisode = &episode1
	meta2.ResourceType = ResourceTypeTV

	// 合并meta2到meta1
	meta1.Merge(meta2)

	// 验证合并结果
	assert.Equal(t, MediaTypeTV, meta1.Type)
	assert.Equal(t, "中文名称", meta1.CnName)
	assert.Equal(t, "2023", meta1.Year)
	assert.Equal(t, 1, *meta1.BeginSeason)
	assert.Equal(t, 1, *meta1.BeginEpisode)
	assert.Equal(t, ResourceTypeTV, meta1.ResourceType)
}

func TestMetaBase_ToDict(t *testing.T) {
	// 创建MetaBase实例
	meta := NewMetaBase("标题", "原始字符串", "副标题", true)
	meta.Type = MediaTypeTV
	meta.CnName = "中文名称"
	season1 := 1
	meta.BeginSeason = &season1
	episode1 := 1
	meta.BeginEpisode = &episode1
	meta.ResourceType = ResourceTypeTV

	// 转换为字典
	dict := meta.ToDict()

	// 验证字典内容
	assert.Equal(t, true, dict["isfile"])
	assert.Equal(t, "标题", dict["title"])
	assert.Equal(t, "原始字符串", dict["org_string"])
	assert.Equal(t, "副标题", dict["subtitle"])
	assert.Equal(t, "中文名称", dict["cn_name"])
	assert.Equal(t, "tv", dict["type"])
	assert.Equal(t, 1, *dict["begin_season"].(*int))
	assert.Equal(t, 1, *dict["begin_episode"].(*int))
	assert.Equal(t, "tv", dict["resource_term"])
	assert.Equal(t, "tv", dict["edition"])
	assert.Equal(t, "S01", dict["season"])
	assert.Equal(t, "E01", dict["episode"])
	assert.Equal(t, "S01 E01", dict["season_episode"])
}
