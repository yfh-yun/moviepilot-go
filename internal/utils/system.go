package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemUtils 系统工具类，提供系统相关的操作和信息获取方法
type SystemUtils struct{}

// NewSystemUtils 创建一个新�?SystemUtils 实例
func NewSystemUtils() *SystemUtils {
	return &SystemUtils{}
}

// Execute 执行命令，获得返回结�?func (s *SystemUtils) Execute(cmd string) string {
	var out strings.Builder
	
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	
	command := exec.Command(parts[0], parts[1:]...)
	command.Stdout = &out
	command.Stderr = &out
	
	err := command.Run()
	if err != nil {
		fmt.Printf("执行命令出错: %v\n", err)
		return ""
	}
	
	return strings.TrimSpace(out.String())
}

// ExecuteWithSubprocess 执行命令并捕获标准输出和错误输出
func (s *SystemUtils) ExecuteWithSubprocess(command []string) (bool, string) {
	if len(command) == 0 {
		return false, "命令不能为空"
	}
	
	cmd := exec.Command(command[0], command[1:]...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		errorMessage := fmt.Sprintf("命令�?s，执行失败，错误信息�?s", strings.Join(command, " "), strings.TrimSpace(string(output)))
		return false, errorMessage
	}
	
	return true, strings.TrimSpace(string(output))
}

// IsDocker 判断是否为Docker环境
func (s *SystemUtils) IsDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}

// IsSynology 判断是否为群晖系�?func (s *SystemUtils) IsSynology() bool {
	if s.IsWindows() {
		return false
	}
	
	output := s.Execute("uname -a")
	return strings.Contains(strings.ToLower(output), "synology")
}

// IsWindows 判断是否为Windows系统
func (s *SystemUtils) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsFrozen 判断是否为编译后的二进制文件
// Go 程序总是编译成二进制文件，这里简单返�?true
func (s *SystemUtils) IsFrozen() bool {
	return true
}

// IsMacos 判断是否为MacOS系统
func (s *SystemUtils) IsMacos() bool {
	return runtime.GOOS == "darwin"
}

// IsAarch64 判断是否为ARM64架构
func (s *SystemUtils) IsAarch64() bool {
	arch := runtime.GOARCH
	return arch == "arm64" || arch == "aarch64"
}

// IsAarch 判断是否为ARM32架构
func (s *SystemUtils) IsAarch() bool {
	arch := runtime.GOARCH
	return (strings.HasPrefix(arch, "arm") || strings.HasPrefix(arch, "aarch")) &&
		arch != "arm64" && arch != "aarch64"
}

// IsX8664 判断是否为AMD64架构
func (s *SystemUtils) IsX8664() bool {
	return runtime.GOARCH == "amd64"
}

// IsX8632 判断是否为x86架构
func (s *SystemUtils) IsX8632() bool {
	arch := runtime.GOARCH
	return arch == "386" || arch == "x86"
}

// Platform 获取系统平台
func (s *SystemUtils) Platform() string {
	if s.IsWindows() {
		return "Windows"
	} else if s.IsMacos() {
		return "MacOS"
	} else if s.IsAarch64() {
		return "Arm64"
	} else {
		return "Linux"
	}
}

// CPUArch 获取CPU架构
func (s *SystemUtils) CPUArch() string {
	if s.IsX8664() {
		return "x86_64"
	} else if s.IsX8632() {
		return "x86_32"
	} else if s.IsAarch64() {
		return "Arm64"
	} else if s.IsAarch() {
		return "Arm32"
	} else {
		return runtime.GOARCH
	}
}

// Copy 复制文件
func (s *SystemUtils) Copy(src, dest string) (int, string) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return -1, err.Error()
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return -1, err.Error()
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return -1, err.Error()
	}

	// 复制文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return -1, err.Error()
	}
	
	err = os.Chmod(dest, srcInfo.Mode())
	if err != nil {
		return -1, err.Error()
	}

	return 0, ""
}

// Move 移动文件
func (s *SystemUtils) Move(src, dest string) (int, string) {
	err := os.Rename(src, dest)
	if err != nil {
		// 如果 Rename 失败，尝试手动移�?		result, msg := s.Copy(src, dest)
		if result != 0 {
			return result, msg
		}
		
		// 删除源文�?		err = os.Remove(src)
		if err != nil {
			return -1, err.Error()
		}
		
		return 0, ""
	}
	
	return 0, ""
}

// Link 创建硬链�?func (s *SystemUtils) Link(src, dest string) (int, string) {
	// 准备目标路径，增加后缀 .mp
	tmpPath := dest + ".mp"
	
	// 检查目标路径是否已存在，如果存在则先删�?	if _, err := os.Stat(tmpPath); err == nil {
		err = os.Remove(tmpPath)
		if err != nil {
			return -1, err.Error()
		}
	}
	
	// 创建硬链�?	err := os.Link(src, tmpPath)
	if err != nil {
		return -1, err.Error()
	}
	
	// 硬链接完成，移除 .mp 后缀
	err = os.Rename(tmpPath, dest)
	if err != nil {
		return -1, err.Error()
	}
	
	return 0, ""
}

// Softlink 创建软链�?func (s *SystemUtils) Softlink(src, dest string) (int, string) {
	err := os.Symlink(src, dest)
	if err != nil {
		return -1, err.Error()
	}
	return 0, ""
}

// ListFiles 获取目录下所有指定扩展名的文件（包括子目录）
func (s *SystemUtils) ListFiles(directory string, extensions []string, minFilesize int64, recursive bool) ([]string, error) {
	if minFilesize < 0 {
		minFilesize = 0
	}
	
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	if !dirInfo.IsDir() {
		return []string{directory}, nil
	}
	
	var files []string
	var pattern *regexp.Regexp
	
	if extensions != nil && len(extensions) > 0 {
		extPattern := strings.Join(extensions, "|")
		pattern, err = regexp.Compile("(?i).*\\.(" + extPattern + ")$")
		if err != nil {
			return nil, err
		}
	}
	
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			// 如果不是递归且不是根目录，则跳过
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}
		
		// 检查扩展名
		if pattern != nil && !pattern.MatchString(info.Name()) {
			return nil
		}
		
		// 检查文件大�?		if info.Size() < minFilesize*1024*1024 {
			return nil
		}
		
		files = append(files, path)
		return nil
	}
	
	err = filepath.Walk(directory, walkFunc)
	if err != nil {
		return nil, err
	}
	
	return files, nil
}

// ExitsFiles 判断目录下是否存在指定扩展名的文�?func (s *SystemUtils) ExitsFiles(directory string, extensions []string, minFilesize int64, recursive bool) (bool, error) {
	if minFilesize < 0 {
		minFilesize = 0
	}
	
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return false, nil
	}
	
	if err != nil {
		return false, err
	}
	
	if !dirInfo.IsDir() {
		return true, nil
	}
	
	var pattern *regexp.Regexp
	
	if extensions != nil && len(extensions) > 0 {
		extPattern := strings.Join(extensions, "|")
		pattern, err = regexp.Compile("(?i).*\\.(" + extPattern + ")$")
		if err != nil {
			return false, err
		}
	}
	
	found := false
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return err
		}
		
		if info.IsDir() {
			// 如果不是递归且不是根目录，则跳过
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}
		
		// 检查扩展名
		if pattern != nil && !pattern.MatchString(info.Name()) {
			return nil
		}
		
		// 检查文件大�?		if info.Size() >= minFilesize*1024*1024 {
			found = true
			return filepath.SkipDir // 找到文件后可以提前结�?		}
		
		return nil
	}
	
	err = filepath.Walk(directory, walkFunc)
	if err != nil {
		return false, err
	}
	
	return found, nil
}

// ListSubFiles 列出当前目录下的所有指定扩展名的文�?不包括子目录)
func (s *SystemUtils) ListSubFiles(directory string, extensions []string) ([]string, error) {
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	if !dirInfo.IsDir() {
		return []string{directory}, nil
	}
	
	var files []string
	var pattern *regexp.Regexp
	
	if extensions != nil && len(extensions) > 0 {
		extPattern := strings.Join(extensions, "|")
		pattern, err = regexp.Compile("(?i).*\\.(" + extPattern + ")$")
		if err != nil {
			return nil, err
		}
	}
	
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// 检查扩展名
		if pattern != nil && !pattern.MatchString(info.Name()) {
			continue
		}
		
		fullPath := filepath.Join(directory, info.Name())
		files = append(files, fullPath)
	}
	
	return files, nil
}

// ListSubDirectory 列出当前目录下的所有子目录（不递归�?func (s *SystemUtils) ListSubDirectory(directory string) ([]string, error) {
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	if !dirInfo.IsDir() {
		return []string{}, nil
	}
	
	var dirs []string
	
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Unix系统下忽略以.开头的隐藏目录
		if !s.IsWindows() && strings.HasPrefix(name, ".") {
			continue
		}
		
		// 忽略@eaDir目录
		if name == "@eaDir" {
			continue
		}
		
		fullPath := filepath.Join(directory, name)
		dirs = append(dirs, fullPath)
	}
	
	return dirs, nil
}

// ListSubFile 列出当前目录下的所有子目录和文件（不递归�?func (s *SystemUtils) ListSubFile(directory string) ([]string, error) {
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	if !dirInfo.IsDir() {
		return []string{directory}, nil
	}
	
	var items []string
	
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		fullPath := filepath.Join(directory, info.Name())
		items = append(items, fullPath)
	}
	
	return items, nil
}

// GetDirectorySize 计算目录的大�?func (s *SystemUtils) GetDirectorySize(path string) (int64, error) {
	var totalSize int64
	
	pathInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	
	if err != nil {
		return 0, err
	}
	
	if !pathInfo.IsDir() {
		return pathInfo.Size(), nil
	}
	
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			totalSize += info.Size()
		}
		
		return nil
	})
	
	return totalSize, err
}

// SpaceUsage 计算多个目录的总可用空�?剩余空间（单位：Byte），并去除重复磁�?func (s *SystemUtils) SpaceUsage(dirList []string) (float64, float64, error) {
	if len(dirList) == 0 {
		return 0.0, 0.0, nil
	}
	
	// 存储不重复的磁盘
	diskSet := make(map[string]bool)
	// 存储总剩余空�?	var totalFreeSpace float64
	// 存储总空�?	var totalSpace float64
	
	for _, dirPath := range dirList {
		if dirPath == "" {
			continue
		}
		
		_, err := os.Stat(dirPath)
		if os.IsNotExist(err) {
			continue
		}
		
		if err != nil {
			continue
		}
		
		// 获取目录所在磁�?		var diskName string
		if s.IsWindows() {
			// Windows 下获取盘�?			diskName = filepath.VolumeName(dirPath)
			if diskName == "" {
				// 如果没有获取到盘符，使用完整路径
				absPath, _ := filepath.Abs(dirPath)
				diskName = filepath.VolumeName(absPath)
			}
		} else {
			// Unix/Linux 下获取设备号
			_, err := os.Stat(dirPath)
			if err != nil {
				continue
			}
			
			// 这里简化处理，直接使用路径作为标识
			// 实际应用中可能需要使�?syscall.Stat_t 中的 Dev 字段
			diskName = dirPath
		}
		
		// 如果磁盘未出现过，则计算其剩余空间并加入总剩余空间中
		if _, exists := diskSet[diskName]; !exists {
			diskSet[diskName] = true
			
			space, err := disk.Usage(dirPath)
			if err != nil {
				continue
			}
			
			totalSpace += float64(space.Total)
			totalFreeSpace += float64(space.Free)
		}
	}
	
	return totalSpace, totalFreeSpace, nil
}

// FreeSpace 获取指定路径的剩余空间（单位：Byte�?func (s *SystemUtils) FreeSpace(path string) (float64, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0.0, nil
	}
	
	if err != nil {
		return 0.0, err
	}
	
	usage, err := disk.Usage(path)
	if err != nil {
		return 0.0, err
	}
	
	return float64(usage.Free), nil
}

// TotalSpace 获取指定路径的总空间（单位：Byte�?func (s *SystemUtils) TotalSpace(path string) (float64, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0.0, nil
	}
	
	if err != nil {
		return 0.0, err
	}
	
	usage, err := disk.Usage(path)
	if err != nil {
		return 0.0, err
	}
	
	return float64(usage.Total), nil
}

// Processes 获取所有进程信�?func (s *SystemUtils) Processes() ([]ProcessInfo, error) {
	processes := []ProcessInfo{}
	
	ps, err := process.Processes()
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	
	for _, p := range ps {
		// 获取进程状�?		status, err := p.Status()
		if err != nil || len(status) == 0 {
			continue
		}
		
		// 跳过僵尸进程
		isZombie := false
		for _, s := range status {
			if s == "Z" || s == "zombie" {
				isZombie = true
				break
			}
		}
		
		if isZombie {
			continue
		}
		
		// 获取进程名称
		name, err := p.Name()
		if err != nil {
			continue
		}
		
		// 获取创建时间
		createTime, err := p.CreateTime()
		if err != nil {
			continue
		}
		
		// 计算运行时间（秒�?		createTimeObj := time.Unix(createTime/1000, 0)
		runTime := int(now.Sub(createTimeObj).Seconds())
		
		// 获取内存信息
		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		
		// 计算内存占用(MB)
		memMB := float64(memInfo.RSS) / (1024 * 1024)
		
		processInfo := ProcessInfo{
			PID:     int(p.Pid),
			Name:    name,
			RunTime: runTime,
			Memory:  memMB,
		}
		
		processes = append(processes, processInfo)
	}
	
	return processes, nil
}

// ProcessInfo 进程信息结构�?type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	RunTime int     `json:"run_time"`
	Memory  float64 `json:"memory"`
}

// IsBlurayDir 判断是否为蓝光原盘目�?func (s *SystemUtils) IsBlurayDir(dirPath string) (bool, error) {
	dirInfo, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	
	if err != nil {
		return false, err
	}
	
	if !dirInfo.IsDir() {
		return false, nil
	}
	
	// 蓝光原盘目录必备的文件或文件�?	requiredFiles := []string{"BDMV", "CERTIFICATE"}
	
	// 检查目录下是否存在所需文件或文件夹
	for _, item := range requiredFiles {
		itemPath := filepath.Join(dirPath, item)
		if _, err := os.Stat(itemPath); err == nil {
			return true, nil
		}
	}
	
	return false, nil
}

// GetWindowsDrives 获取Windows所有盘�?func (s *SystemUtils) GetWindowsDrives() []string {
	if !s.IsWindows() {
		return []string{}
	}
	
	vols := []string{}
	for i := 65; i <= 90; i++ {
		vol := string(rune(i)) + ":"
		if _, err := os.Stat(vol); err == nil {
			vols = append(vols, vol)
		}
	}
	
	return vols
}

// CPUUsage 获取CPU使用�?func (s *SystemUtils) CPUUsage() (float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	
	if len(percent) > 0 {
		return percent[0], nil
	}
	
	return 0, nil
}

// MemoryUsage 获取当前程序的内存使用量和使用率
func (s *SystemUtils) MemoryUsage() ([]int, error) {
	// 获取当前进程内存信息
	pid := os.Getpid()
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}
	
	processMemory := memInfo.RSS
	
	// 获取系统内存信息
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	
	systemMemory := vmStat.Total
	processMemoryPercent := float64(processMemory) / float64(systemMemory) * 100
	
	return []int{int(processMemory), int(processMemoryPercent)}, nil
}

// NetworkUsage 获取当前网络流量（上行和下行流量，单位：bytes/s�?func (s *SystemUtils) NetworkUsage() ([]int64, error) {
	// 获取初始网络统计
	netIO1, err := gopsnet.IOCounters(false)
	if err != nil {
		return nil, err
	}
	
	time.Sleep(time.Second)
	
	// 获取1秒后的网络统�?	netIO2, err := gopsnet.IOCounters(false)
	if err != nil {
		return nil, err
	}
	
	if len(netIO1) > 0 && len(netIO2) > 0 {
		// 计算1秒内的流量变�?		uploadSpeed := int64(netIO2[0].BytesSent) - int64(netIO1[0].BytesSent)
		downloadSpeed := int64(netIO2[0].BytesRecv) - int64(netIO1[0].BytesRecv)
		return []int64{uploadSpeed, downloadSpeed}, nil
	}
	
	return []int64{0, 0}, nil
}

// IsHardlink 判断是否为硬链接
func (s *SystemUtils) IsHardlink(src, dest string) bool {
	srcInfo, srcErr := os.Stat(src)
	destInfo, destErr := os.Stat(dest)
	
	if os.IsNotExist(srcErr) || os.IsNotExist(destErr) {
		return false
	}
	
	if srcInfo.IsDir() != destInfo.IsDir() {
		return false
	}
	
	if srcInfo.IsDir() {
		// 对于目录，检查其中的文件是否为硬链接
		srcEntries, err := os.ReadDir(src)
		if err != nil {
			return false
		}
		
		for _, entry := range srcEntries {
			if entry.IsDir() {
				continue
			}
			
			srcFilePath := filepath.Join(src, entry.Name())
			destFilePath := filepath.Join(dest, entry.Name())
			
			srcFileInfo, err := os.Stat(srcFilePath)
			if err != nil {
				return false
			}
			
			destFileInfo, err := os.Stat(destFilePath)
			if err != nil {
				return false
			}
			
			// 比较 inode �?device id
			if os.SameFile(srcFileInfo, destFileInfo) {
				continue
			} else {
				return false
			}
		}
		
		return true
	} else {
		// 对于文件，直接比�?		return os.SameFile(srcInfo, destInfo)
	}
}

// IsNetworkFilesystem 检测是否为网络文件系统
func (s *SystemUtils) IsNetworkFilesystem(directory string) bool {
	// 在Go中检测网络文件系统较为复杂，这里提供一个基础实现
	// 可以根据具体需求进行增�?	
	system := runtime.GOOS
	try := func(cmd string, args ...string) string {
		command := exec.Command(cmd, args...)
		output, err := command.Output()
		if err != nil {
			return ""
		}
		return strings.ToLower(string(output))
	}
	
	switch system {
	case "linux":
		output := try("df", "-T", directory)
		if output != "" {
			// 以下本地文件系统含有fuse关键�?			localFS := []string{
				"fuse.shfs",    // Unraid
				"zfuse.zfsv",   // 极空�?			}
			
			for _, fs := range localFS {
				if strings.Contains(output, fs) {
					return false
				}
			}
			
			networkFS := []string{"nfs", "cifs", "smbfs", "fuse", "sshfs", "ftpfs"}
			for _, fs := range networkFS {
				if strings.Contains(output, fs) {
					return true
				}
			}
		}
	case "darwin": // macOS
		output := try("df", "-T", directory)
		if output != "" {
			return strings.Contains(output, "nfs") || strings.Contains(output, "smbfs")
		}
	case "windows":
		// Windows 检查网络驱动器
		return strings.HasPrefix(directory, "\\\\")
	}
	
	return false
}

// IsSameDisk 判断两个路径是否在同一磁盘
func (s *SystemUtils) IsSameDisk(src, dest string) bool {
	srcInfo, srcErr := os.Stat(src)
	destInfo, destErr := os.Stat(dest)
	
	if os.IsNotExist(srcErr) || os.IsNotExist(destErr) {
		return false
	}
	
	if s.IsWindows() {
		srcAbs, _ := filepath.Abs(src)
		destAbs, _ := filepath.Abs(dest)
		return filepath.VolumeName(srcAbs) == filepath.VolumeName(destAbs)
	}
	
	// Unix/Linux 系统通过设备号判�?	srcDev := getDeviceID(srcInfo)
	destDev := getDeviceID(destInfo)
	return srcDev == destDev
}

// getDeviceID 获取文件的设备ID（简化实现）
func getDeviceID(info os.FileInfo) uint64 {
	// 在实际实现中，你可能需要使�?syscall.Stat_t
	// 这里为了简化，返回一个固定�?	// 实际项目中应根据平台具体实现
	return 0
}

// GetConfigPath 获取配置路径
func (s *SystemUtils) GetConfigPath(configDir string) string {
	if configDir == "" {
		configDir = os.Getenv("CONFIG_DIR")
	}
	
	if configDir != "" {
		return configDir
	}
	
	if s.IsDocker() {
		return "/config"
	} else if s.IsFrozen() {
		execPath, err := os.Executable()
		if err != nil {
			// 回退到当前工作目�?			wd, _ := os.Getwd()
			return filepath.Join(wd, "config")
		}
		return filepath.Join(filepath.Dir(execPath), "config")
	} else {
		// 获取当前文件的目录，然后向上两级找到config目录
		_, filename, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(filename), "..", "..", "config")
	}
}

// GetEnvPath 获取环境配置路径
func (s *SystemUtils) GetEnvPath() string {
	configPath := s.GetConfigPath("")
	return filepath.Join(configPath, "app.env")
}

// Clear 清理指定目录中指定天数前的文件，递归删除子文件及空文件夹
func (s *SystemUtils) Clear(tempPath string, days int) error {
	// 检查目录是否存�?	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		return nil
	}
	
	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	
	// 先删除过期文�?	err := filepath.Walk(tempPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 如果是文件且修改时间早于截止时间，则删除
		if !info.IsDir() && info.ModTime().Before(cutoffTime) {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// 删除空的文件夹（逆序遍历以确保从最深层开始删除）
	return s.removeEmptyDirs(tempPath)
}

// removeEmptyDirs 删除空目�?func (s *SystemUtils) removeEmptyDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	
	// 递归处理子目�?	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			err := s.removeEmptyDirs(subDir)
			if err != nil {
				return err
			}
		}
	}
	
	// 再次检查目录是否为�?	entries, err = os.ReadDir(dir)
	if err != nil {
		return err
	}
	
	// 如果目录为空，则删除
	if len(entries) == 0 {
		return os.Remove(dir)
	}
	
	return nil
}

// GenerateUserUniqueID 根据优先级依次尝试生成稳定唯一ID
func (s *SystemUtils) GenerateUserUniqueID() *string {
	methods := []func() *string{
		s.getFilesystemUniqueID,
		s.getMacAddressID,
	}
	
	for _, method := range methods {
		uniqueID := method()
		if uniqueID != nil {
			return uniqueID
		}
	}
	
	return nil
}

// getFilesystemUniqueID 获取文件系统的唯一标识�?func (s *SystemUtils) getFilesystemUniqueID() *string {
	// 在Unix-like系统上尝试获取根目录的设备信�?	if runtime.GOOS == "windows" {
		return nil
	}
	
	info, err := os.Stat("/")
	if err != nil {
		return nil
	}
	
	// 简化实现，实际应该获取设备号和inode
	fsID := fmt.Sprintf("%s-%d", info.Name(), info.Size())
	hasher := sha256.New()
	hasher.Write([]byte(fsID))
	uniqueID := fmt.Sprintf("%x", hasher.Sum(nil))
	return &uniqueID
}

// getMacAddressID 获取设备�?MAC 地址并生成唯一标识�?func (s *SystemUtils) getMacAddressID() *string {
	interfaces, err := gopsnet.Interfaces()
	if err != nil {
		return nil
	}
	
	for _, iface := range interfaces {
		// 跳过回环接口
		isLoopback := false
		for _, flag := range iface.Flags {
			if flag == "loopback" {
				isLoopback = true
				break
			}
		}
		if isLoopback {
			continue
		}
		
		// 跳过禁用的接�?		isUp := false
		for _, flag := range iface.Flags {
			if flag == "up" {
				isUp = true
				break
			}
		}
		if !isUp {
			continue
		}
		
		// 确保有硬件地址
		if len(iface.HardwareAddr) > 0 {
			macStr := fmt.Sprintf("%x", iface.HardwareAddr)
			hasher := sha256.New()
			hasher.Write([]byte(macStr))
			uniqueID := fmt.Sprintf("%x", hasher.Sum(nil))
			return &uniqueID
		}
	}
	
	return nil
}

// getHostID 获取主机ID（基于主机信息）
func (s *SystemUtils) getHostID() *string {
	info, err := host.Info()
	if err != nil {
		return nil
	}
	
	hostInfo := fmt.Sprintf("%s-%s-%s", info.Hostname, info.Platform, info.PlatformVersion)
	hasher := sha256.New()
	hasher.Write([]byte(hostInfo))
	uniqueID := fmt.Sprintf("%x", hasher.Sum(nil))
	return &uniqueID
}
