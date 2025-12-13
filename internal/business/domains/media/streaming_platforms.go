package media

import (
	"strings"
	"sync"
)

// StreamingPlatforms 流媒体平台映射
type StreamingPlatforms struct {
	lookup map[string]string // UPPER(alias) -> canonical name
	once   sync.Once         // 单例初始化锁
}

var (
	// streamingPlatformsInstance 单例实例
	streamingPlatformsInstance *StreamingPlatforms
	// streamingPlatformsMutex 单例初始化互斥锁
	streamingPlatformsMutex sync.Mutex
)

// streamingPlatformData 流媒体平台数据 - 与Python版本完全一致
var streamingPlatformData = []struct {
	shortName string
	fullName  string
}{
	{"AMZN", "Amazon"},
	{"NF", "Netflix"},
	{"ATVP", "Apple TV+"},
	{"iT", "iTunes"},
	{"DSNP", "Disney+"},
	{"HS", "Hotstar"},
	{"APPS", "Disney+ MENA"},
	{"PMTP", "Paramount+"},
	{"HMAX", "Max"},
	{"", "Max"},
	{"HULU", "Hulu Networks"},
	{"MA", "Movies Anywhere"},
	{"BCORE", "Bravia Core"},
	{"MS", "Microsoft Store"},
	{"SHO", "Showtime"},
	{"STAN", "Stan"},
	{"PCOK", "Peacock"},
	{"SKST", "SkyShowtime"},
	{"NOW", "Now"},
	{"FXTL", "Foxtel Now"},
	{"BNGE", "Binge"},
	{"CRKL", "Crackle"},
	{"RKTN", "Rakuten TV"},
	{"ALL4", "Channel 4"},
	{"AS", "Adult Swim"},
	{"BRTB", "Brtb TV"},
	{"CNLP", "Canal+"},
	{"CRIT", "Criterion Channel"},
	{"DSCP", "Discovery+"},
	{"FOOD", "Food Network"},
	{"MUBI", "Mubi"},
	{"PLAY", "Google Play"},
	{"YT", "YouTube"},
	{"", "friDay"},
	{"", "KKTV"},
	{"", "ofiii"},
	{"", "LiTV"},
	{"", "MyVideo"},
	{"Hami", "Hami Video"},
	{"HamiVideo", "Hami Video"},
	{"MW", "meWATCH"},
	{"CATCHPLAY", "CATCHPLAY+"},
	{"CPP", "CATCHPLAY+"},
	{"LINETV", "LINE TV"},
	{"VIU", "Viu"},
	{"IQ", ""},
	{"", "WeTV"},
	{"ABMA", "Abema"},
	{"ADN", ""},
	{"AT-X", ""},
	{"Baha", ""},
	{"BG", "B-Global"},
	{"CR", "Crunchyroll"},
	{"", "DMM"},
	{"FOD", ""},
	{"FUNi", "Funimation"},
	{"HIDI", "HIDIVE"},
	{"UNXT", "U-NEXT"},
	{"FAA", "Filmarchiv Austria"},
	{"CC", "Comedy Central"},
	{"iP", "BBC iPlayer"},
	{"9NOW", "9Now"},
	{"ABC", ""},
	{"", "AMC"},
	{"", "ZEE5"},
	{"", "WAVO"},
	{"SHAHID", "Shahid"},
	{"Flixole", "FlixOlé"},
	{"TOU", "Ici TOU.TV"},
	{"ROKU", "Roku"},
	{"KNPY", "Kanopy"},
	{"SNXT", "Sun NXT"},
	{"CUR", "Curiosity Stream"},
	{"MY5", "Channel 5"},
	{"AHA", "aha"},
	{"WOWP", "WOW Presents Plus"},
	{"JC", "JioCinema"},
	{"", "Dekkoo"},
	{"FILMZIE", "Filmzie"},
	{"HoiChoi", "Hoichoi"},
	{"VIKI", "Rakuten Viki"},
	{"SF", "SF Anytime"},
	{"PLEX", "Plex"},
	{"SHDR", "Shudder"},
	{"CRAV", "Crave"},
	{"CPE", "Cineplex Entertainment"},
	{"JF HC", ""},
	{"JF", ""},
	{"JFFP", ""},
	{"VIAP", "Viaplay"},
	{"TUBI", "TubiTV"},
	{"", "PBS"},
	{"PBSK", "PBS KIDS"},
	{"LGP", "Lionsgate Play"},
	{"", "CTV"},
	{"", "Cineverse"},
	{"LN", "Love Nature"},
	{"MP", "Movistar Plus+"},
	{"RUNTIME", "Runtime"},
	{"STZ", "STARZ"},
	{"FUBO", "fuboTV"},
	{"TENK", "Tënk"},
	{"KNOW", "Knowledge Network"},
	{"TVO", "tvo"},
	{"", "OVID"},
	{"CBC", "CBC Gem"},
	{"FANDOR", "fandor"},
	{"CW", "The CW"},
	{"KNPY", "Kanopy"},
	{"FREE", "Freeform"},
	{"AE", "A&E"},
	{"LIFE", "Lifetime"},
	{"WWEN", "WWE Network"},
	{"CMAX", "Cinemax"},
	{"HLMK", "Hallmark"},
	{"BYU", "BYUtv"},
	{"", "ViX"},
	{"VICE", "Viceland"},
	{"", "TVING"},
	{"USAN", "USA Network"},
	{"FOX", ""},
	{"", "TCM"},
	{"BRAV", "BravoTV"},
	{"", "TNT"},
	{"", "ZDF"},
	{"", "IndieFlix"},
	{"", "TLC"},
	{"", "HGTV"},
	{"ANPL", "Animal Planet"},
	{"TRVL", "Travel Channel"},
	{"", "VH1"},
	{"SAINA", "Saina Play"},
	{"SP", "Saina Play"},
	{"OXGN", "Oxygen"},
	{"PSN", "PlayStation Network"},
	{"PMNT", "Paramount Network"},
	{"FAWESOME", "Fawesome"},
	{"KLASSIKI", "Klassiki"},
	{"STRP", "Star+"},
	{"NATG", "National Geographic"},
	{"REVEEL", "Reveel"},
	{"FYI", "FYI Network"},
	{"WatchiT", "WATCH IT"},
	{"ITVX", "ITV"},
	{"GAIA", "Gaia"},
	{"", "FlixLatino"},
	{"CNNP", "CNN+"},
	{"TROMA", "Troma"},
	{"IVI", "Ivi"},
	{"9NOW", "9Now"},
	{"A3P", "Atresplayer"},
	{"7PLUS", "7plus"},
	{"", "SBS"},
	{"TEN", "10Play"},
	{"AUBC", ""},
	{"DSNY", "Disney Networks"},
	{"OSN", "OSN+"},
	{"SVT", "Sveriges Television"},
	{"LACINETEK", "LaCinetek"},
	{"", "Maxdome"},
	{"RTL", "RTL+"},
	{"ARTE", "Arte"},
	{"JOYN", "Joyn"},
	{"TV2", "TV 2"},
	{"3SAT", "3sat"},
	{"FILMINGO", "filmingo"},
	{"", "WOW"},
	{"OKKO", "Okko"},
	{"", "Go3"},
	{"ARGP", "Argo"},
	{"VOYO", "Voyo"},
	{"VMAX", "vivamax"},
	{"FILMIN", "Filmin"},
	{"", "Mitele"},
	{"MY5", "Channel 5"},
	{"", "ARD"},
	{"BK", "Bentkey"},
	{"BOOM", "Boomerang"},
	{"", "CBS"},
	{"CLBI", "Club illico"},
	{"CMOR", "C More"},
	{"CMT", ""},
	{"", "CNBC"},
	{"COOK", "Cooking Channel"},
	{"CWS", "CW Seed"},
	{"DCU", "DC Universe"},
	{"DDY", "Digiturk Dilediğin Yerde"},
	{"DEST", "Destination America"},
	{"DISC", "Discovery Channel"},
	{"DW", "DailyWire+"},
	{"DLWP", "DailyWire+"},
	{"DPLY", "dplay"},
	{"DRPO", "Dropout"},
	{"EPIX", "EPIX MGM+"},
	{"ESQ", "Esquire"},
	{"ETV", "E!"},
	{"FBWatch", "Facebook Watch"},
	{"FPT", "FPT Play"},
	{"FTV", "France.tv"},
	{"GLOB", "GloboSat Play"},
	{"GLBO", "Globoplay"},
	{"GO90", "go90"},
	{"HIST", "History Channel"},
	{"HPLAY", "Hungama Play"},
	{"KS", "Kaleidescape"},
	{"", "MBC"},
	{"MMAX", "ManoramaMAX"},
	{"MNBC", "MSNBC"},
	{"MTOD", "Motor Trend OnDemand"},
	{"NBC", ""},
	{"NBLA", "Nebula"},
	{"NICK", "Nickelodeon"},
	{"ODK", "OnDemandKorea"},
	{"POGO", "PokerGO"},
	{"PUHU", "puhutv"},
	{"QIBI", "Quibi"},
	{"RTE", "RTÉ"},
	{"SESO", "Seeso"},
	{"SPIK", "Spike"},
	{"SS", "Simply South"},
	{"SYFY", "SyFy"},
	{"TIMV", "TIMvision"},
	{"TK", "Tentkotta"},
	{"", "TV4"},
	{"TVL", "TV Land"},
	{"", "TVNZ"},
	{"", "UKTV"},
	{"VLCT", "Discovery Velocity"},
	{"VMEO", "Vimeo"},
	{"VRV", "VRV Defunct"},
	{"WTCH", "Watcha"},
	{"", "NowPlayer"},
	{"HuluJP", "Hulu Networks"},
	{"Gaga", "GagaOOLala"},
	{"MyTVS", "MyTVSuper"},
	{"", "BBC"},
	{"CC", "Comedy Central"},
	{"NowE", "Now E"},
	{"WAVVE", "Wavve"},
	{"SE", ""},
	{"", "BritBox"},
	{"AOD", "Anime on Demand"},
	{"AF", ""},
	{"BCH", "Bandai Channel"},
	{"VMJ", "VideoMarket"},
	{"LFTL", "Laftel"},
	{"WAKA", "Wakanim"},
	{"WAKANIM", "Wakanim"},
	{"AO", "AnimeOnegai"},
	{"", "Lemino"},
	{"VIDIO", "Vidio"},
	{"TVER", "TVer"},
	{"", "MBS"},
	{"LFTLNET", "Laftel"},
	{"JONU", "Jonu Play"},
	{"PlutoTV", "Pluto TV"},
	{"AbemaTV", "Abema"},
	{"", "dTV"},
	{"NYMEY", "Nymey"},
	{"SMNS", "SAMANSA"},
	{"CTHP", "CATCHPLAY+"},
	{"HBOGO", "HBO GO"},
	{"HBO", "HBO"},
	{"FPTP", "FPT Play"},
	{"", "LOCIPO"},
	{"DANT", "DANET"},
	{"OV", "OceanVeil"},
}

// NewStreamingPlatforms 创建新的StreamingPlatforms实例
func NewStreamingPlatforms() *StreamingPlatforms {
	// 双重检查锁定实现单例模式
	if streamingPlatformsInstance == nil {
		streamingPlatformsMutex.Lock()
		defer streamingPlatformsMutex.Unlock()
		if streamingPlatformsInstance == nil {
			sp := &StreamingPlatforms{
				lookup: make(map[string]string),
			}
			sp.buildCache()
			streamingPlatformsInstance = sp
		}
	}
	return streamingPlatformsInstance
}

// buildCache 构建查找缓存
func (sp *StreamingPlatforms) buildCache() {
	// 清空现有缓存
	for k := range sp.lookup {
		delete(sp.lookup, k)
	}

	// 构建新缓存
	for _, data := range streamingPlatformData {
		shortName := data.shortName
		fullName := data.fullName
		
		// 确定标准名称
		canonicalName := fullName
		if canonicalName == "" {
			canonicalName = shortName
		}
		if canonicalName == "" {
			continue
		}

		// 添加别名到缓存
		aliases := []string{shortName, fullName}
		for _, alias := range aliases {
			if alias != "" {
				sp.lookup[strings.ToUpper(alias)] = canonicalName
			}
		}
	}
}

// GetStreamingPlatformName 根据流媒体平台简称或全称获取标准名称
func (sp *StreamingPlatforms) GetStreamingPlatformName(platformCode string) string {
	if platformCode == "" {
		return ""
	}
	return sp.lookup[strings.ToUpper(platformCode)]
}

// IsStreamingPlatform 判断给定的字符串是否为已知的流媒体平台代码或名称
func (sp *StreamingPlatforms) IsStreamingPlatform(name string) bool {
	if name == "" {
		return false
	}
	_, exists := sp.lookup[strings.ToUpper(name)]
	return exists
}

// IsStreamingPlatform 判断给定的字符串是否为已知的流媒体平台代码或名称
// 兼容旧版方法签名
func (sp *StreamingPlatforms) IsStreamingPlatformOld(name string) bool {
	return sp.IsStreamingPlatform(name)
}
