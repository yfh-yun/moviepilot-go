package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	cpuutil "github.com/shirou/gopsutil/v3/cpu"
	memutil "github.com/shirou/gopsutil/v3/mem"
	netutil "github.com/shirou/gopsutil/v3/net"
	processutil "github.com/shirou/gopsutil/v3/process"
)

// Execute 对应 Python SystemUtils.execute
// 使用 shell 执行命令并返回第一行输出
func Execute(cmd string) (string, error) {
	if strings.TrimSpace(cmd) == "" {
		return "", errors.New("empty command")
	}

	// 为了兼容不同平台，这里不做复杂的 shell 解析，直接交给 /bin/sh 或 cmd
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}

	out, err := c.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", nil
}

// ExecuteWithSubprocess 对应 Python SystemUtils.execute_with_subprocess
// 执行命令并捕获标准输出和错误输出
func ExecuteWithSubprocess(cmd []string) (bool, string) {
	if len(cmd) == 0 {
		return false, "空命令"
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	output := string(out)

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			errorMsg := fmt.Sprintf("命令：%s，执行失败，错误信息：%s", strings.Join(cmd, " "), strings.TrimSpace(output))
			return false, errorMsg
		}
		return false, fmt.Sprintf("未知错误，命令：%s，错误：%v", strings.Join(cmd, " "), err)
	}

	return true, output
}

// IsDocker 判断当前是否运行在 Docker 环境，对应 SystemUtils.is_docker
func IsDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// 某些环境不会有 /.dockerenv，可以再简单检查一下 cgroup 信息
	if b, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if strings.Contains(strings.ToLower(string(b)), "docker") ||
			strings.Contains(strings.ToLower(string(b)), "kubepods") {
			return true
		}
	}
	return false
}

// IsSynology 对应 Python SystemUtils.is_synology
// 判断是否为群晖系统
func IsSynology() bool {
	if IsWindows() {
		return false
	}
	output, err := Execute("uname -a")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(output), "synology")
}

// IsWindows 对应 SystemUtils.is_windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS 对应 SystemUtils.is_macos
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsAarch64 对应 SystemUtils.is_aarch64
func IsAarch64() bool {
	return runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64"
}

// IsAarch 对应 SystemUtils.is_aarch（32 位 ARM）
func IsAarch() bool {
	return strings.HasPrefix(runtime.GOARCH, "arm") && !IsAarch64()
}

// IsX86_64 对应 SystemUtils.is_x86_64
func IsX86_64() bool {
	return runtime.GOARCH == "amd64" || runtime.GOARCH == "x86_64"
}

// IsX86_32 对应 SystemUtils.is_x86_32
func IsX86_32() bool {
	arch := runtime.GOARCH
	return arch == "386" || arch == "i386" || arch == "i686" || arch == "x86"
}

// Platform 对应 SystemUtils.platform
func Platform() string {
	if IsWindows() {
		return "Windows"
	}
	if IsMacOS() {
		return "MacOS"
	}
	if IsAarch64() {
		return "Arm64"
	}
	return "Linux"
}

// CPUArch 对应 SystemUtils.cpu_arch
func CPUArch() string {
	if IsX86_64() {
		return "x86_64"
	}
	if IsX86_32() {
		return "x86_32"
	}
	if IsAarch64() {
		return "Arm64"
	}
	if IsAarch() {
		return "Arm32"
	}
	return runtime.GOARCH
}

// Copy 对应 SystemUtils.copy，复制文件
func Copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("src is not a regular file: %s", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := bufio.NewReader(srcFile).WriteTo(destFile); err != nil {
		return err
	}

	return os.Chmod(dest, info.Mode())
}

// Move 对应 SystemUtils.move，移动/重命名路径
// 按照 Python 版本逻辑：先将源文件在当前目录下重命名为目标文件的名称，然后移动到目标目录
func Move(src, dest string) error {
	// 解析源文件和目标文件的路径
	srcPath := filepath.Clean(src)
	destPath := filepath.Clean(dest)

	// 获取源文件的目录
	srcDir := filepath.Dir(srcPath)
	// 获取目标文件的名称
	destName := filepath.Base(destPath)
	// 获取目标文件的目录
	destDir := filepath.Dir(destPath)

	// 先确保目标目录存在
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	// 1. 将源文件在当前目录下重命名为目标文件的名称
	tempPath := filepath.Join(srcDir, destName)
	if err := os.Rename(srcPath, tempPath); err != nil {
		return err
	}

	// 2. 将重命名后的文件移动到目标目录
	return os.Rename(tempPath, destPath)
}

// HardLink 对应 SystemUtils.link，创建硬链接
// 按照 Python 版本逻辑：先创建带有 .mp 后缀的临时硬链接，然后重命名为目标文件
func HardLink(src, dest string) error {
	// 解析源文件和目标文件的路径
	srcPath := filepath.Clean(src)
	destPath := filepath.Clean(dest)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	// 1. 准备目标路径，增加后缀 .mp
	tmpPath := destPath + ".mp"

	// 2. 检查目标路径是否已存在，如果存在则先删除
	if _, err := os.Stat(tmpPath); err == nil {
		if err := os.Remove(tmpPath); err != nil {
			return err
		}
	}

	// 3. 创建硬链接到临时路径
	if err := os.Link(srcPath, tmpPath); err != nil {
		return err
	}

	// 4. 硬链接完成，移除 .mp 后缀（将临时文件重命名为目标文件）
	return os.Rename(tmpPath, destPath)
}

// SoftLink 对应 SystemUtils.softlink，创建符号链接
func SoftLink(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	_ = os.Remove(dest)
	return os.Symlink(src, dest)
}

// IsHardlink 对应 Python SystemUtils.is_hardlink
// 判断两个路径是否为硬链接
func IsHardlink(src, dest string) bool {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}
	destInfo, err := os.Stat(dest)
	if err != nil {
		return false
	}

	// 如果源路径是文件，直接比较两个文件
	if srcInfo.Mode().IsRegular() {
		return os.SameFile(srcInfo, destInfo)
	}

	// 如果源路径是目录，遍历所有文件进行比较
	var isSame bool = true
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			isSame = false
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		// 计算目标文件路径
		relativePath, err := filepath.Rel(src, path)
		if err != nil {
			isSame = false
			return filepath.SkipDir
		}
		targetPath := filepath.Join(dest, relativePath)

		// 检查目标文件是否存在
		targetInfo, err := os.Stat(targetPath)
		if err != nil {
			isSame = false
			return filepath.SkipDir
		}

		// 检查是否是硬链接
		if !os.SameFile(info, targetInfo) {
			isSame = false
			return filepath.SkipDir
		}

		return nil
	})

	return isSame
}

// CPUUsage 对应 Python SystemUtils.cpu_usage
// 获取CPU使用率
func CPUUsage() float64 {
	percent, err := cpuutil.Percent(0, false)
	if err != nil || len(percent) == 0 {
		return 0.0
	}
	return percent[0]
}

// MemoryUsage 对应 Python SystemUtils.memory_usage
// 获取当前程序的内存使用量和使用率
// 返回值：[进程内存使用量（byte）, 进程内存使用率（%）]
func MemoryUsage() []int64 {
	// 获取当前进程ID
	pid := int32(os.Getpid())
	// 获取进程对象
	process, err := processutil.NewProcess(pid)
	if err != nil {
		return []int64{0, 0}
	}

	// 获取进程内存信息
	memInfo, err := process.MemoryInfo()
	if err != nil {
		return []int64{0, 0}
	}

	// 获取系统内存信息
	sysMemInfo, err := memutil.VirtualMemory()
	if err != nil {
		return []int64{0, 0}
	}

	// 计算进程内存使用率
	processMemory := memInfo.RSS
	systemMemory := sysMemInfo.Total
	processMemoryPercent := int64((float64(processMemory) / float64(systemMemory)) * 100)

	return []int64{int64(processMemory), processMemoryPercent}
}

// NetworkUsage 对应 Python SystemUtils.network_usage
// 获取当前网络流量（上行和下行流量，单位：bytes/s）
// 返回值：[上行速度（byte/s）, 下行速度（byte/s）]
func NetworkUsage() []int64 {
	// 获取初始网络统计
	netIO1, err := netutil.IOCounters(false)
	if err != nil || len(netIO1) == 0 {
		return []int64{0, 0}
	}

	// 等待1秒
	time.Sleep(time.Second)

	// 获取1秒后的网络统计
	netIO2, err := netutil.IOCounters(false)
	if err != nil || len(netIO2) == 0 {
		return []int64{0, 0}
	}

	// 计算1秒内的流量变化
	uploadSpeed := netIO2[0].BytesSent - netIO1[0].BytesSent
	downloadSpeed := netIO2[0].BytesRecv - netIO1[0].BytesRecv

	return []int64{int64(uploadSpeed), int64(downloadSpeed)}
}

// ListFiles 对应 SystemUtils.list_files
//   - directory: 根目录
//   - extensions: 扩展名（不带点，例如 ["mkv","mp4"]），大小写不敏感
//   - minFilesizeMB: 最小文件大小（MB）
//   - recursive: 是否递归
func ListFiles(directory string, extensions []string, minFilesizeMB int64, recursive bool) ([]string, error) {
	if directory == "" {
		return nil, nil
	}

	info, err := os.Stat(directory)
	if err != nil {
		return nil, nil
	}

	if info.Mode().IsRegular() {
		return []string{directory}, nil
	}

	var pattern *regexp.Regexp
	if len(extensions) > 0 {
		for i, ext := range extensions {
			extensions[i] = strings.TrimPrefix(strings.ToLower(ext), ".")
		}
		// 类似 Python: ".*(mkv|mp4)$"
		pattern = regexp.MustCompile(".*(\\." + strings.Join(extensions, "|\\.") + ")$")
	} else {
		pattern = regexp.MustCompile(".*")
	}

	minSizeBytes := minFilesizeMB * 1024 * 1024
	files := make([]string, 0)

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}
		if !pattern.MatchString(strings.ToLower(d.Name())) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() >= minSizeBytes {
			files = append(files, path)
		}
		return nil
	}

	if err := filepath.WalkDir(directory, walkFn); err != nil {
		return nil, err
	}
	return files, nil
}

// ExistsFiles 对应 SystemUtils.exits_files（注意原拼写）
func ExistsFiles(directory string, extensions []string, minFilesizeMB int64, recursive bool) bool {
	files, err := ListFiles(directory, extensions, minFilesizeMB, recursive)
	if err != nil {
		return false
	}
	return len(files) > 0
}

// ListSubFiles 对应 SystemUtils.list_sub_files（当前目录下的指定扩展文件，不递归）
func ListSubFiles(directory string, extensions []string) ([]string, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, nil
	}
	if info.Mode().IsRegular() {
		return []string{directory}, nil
	}

	pattern := regexp.MustCompile(".*")
	if len(extensions) > 0 {
		for i, ext := range extensions {
			extensions[i] = strings.TrimPrefix(strings.ToLower(ext), ".")
		}
		pattern = regexp.MustCompile(".*(\\." + strings.Join(extensions, "|\\.") + ")$")
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if pattern.MatchString(strings.ToLower(e.Name())) {
			files = append(files, filepath.Join(directory, e.Name()))
		}
	}
	return files, nil
}

// ListSubDirectories 对应 SystemUtils.list_sub_directory（当前目录下的子目录，不递归）
func ListSubDirectories(directory string) ([]string, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, nil
	}
	if info.Mode().IsRegular() {
		return nil, nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !IsWindows() && strings.HasPrefix(name, ".") {
			continue
		}
		if name == "@eaDir" {
			continue
		}
		dirs = append(dirs, filepath.Join(directory, name))
	}
	return dirs, nil
}

// ListSubFile 对应 SystemUtils.list_sub_file（当前目录下所有文件，不递归）
func ListSubFile(directory string) ([]string, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, nil
	}
	if info.Mode().IsRegular() {
		return []string{directory}, nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files = append(files, filepath.Join(directory, e.Name()))
	}
	return files, nil
}

// GetDirectorySize 对应 SystemUtils.get_directory_size，返回字节数
func GetDirectorySize(path string) int64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if info.Mode().IsRegular() {
		return info.Size()
	}

	var total int64
	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// FreeSpace 使用 `df -B1` 获取路径所在磁盘的可用空间，单位：Byte
func FreeSpace(path string) int64 {
	if path == "" {
		return 0
	}
	cmd := exec.Command("df", "-B1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0
	}
	// 第二行类似：Filesystem 1B-blocks Used Available Use% Mounted on
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0
	}
	v, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// TotalSpace 使用 `df -B1` 获取路径所在磁盘的总空间，单位：Byte
func TotalSpace(path string) int64 {
	if path == "" {
		return 0
	}
	cmd := exec.Command("df", "-B1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0
	}
	v, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// SpaceUsage 对应 SystemUtils.space_usage，统计多个目录所在磁盘的总空间和可用空间（去重磁盘）
// 返回值：(总空间, 可用空间)，单位：Byte
func SpaceUsage(dirs []string) (total, free int64) {
	if len(dirs) == 0 {
		return 0, 0
	}

	// 用于存储已处理的磁盘
	seen := make(map[string]struct{})

	for _, d := range dirs {
		if d == "" {
			continue
		}
		dirPath := filepath.Clean(d)
		info, err := os.Stat(dirPath)
		if err != nil {
			continue
		}

		// 获取目录所在磁盘
		var diskKey string
		if IsWindows() {
			// Windows：使用驱动器号（如 C:）
			diskKey = filepath.VolumeName(dirPath)
		} else {
			// 其他系统：使用设备号
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				diskKey = fmt.Sprintf("%d", stat.Dev)
			} else {
				// 如果无法获取设备号，使用目录所在的挂载点
				diskKey = string(os.PathSeparator)
			}
		}

		// 如果磁盘未处理过，计算其空间并加入总和
		if _, ok := seen[diskKey]; !ok {
			seen[diskKey] = struct{}{}
			total += TotalSpace(dirPath)
			free += FreeSpace(dirPath)
		}
	}
	return total, free
}

// IsBlurayDir 对应 SystemUtils.is_bluray_dir
func IsBlurayDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "BDMV")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, "CERTIFICATE")); err == nil {
		return true
	}
	return false
}

// GetWindowsDrives 对应 SystemUtils.get_windows_drives
func GetWindowsDrives() []string {
	if !IsWindows() {
		return nil
	}
	var vols []string
	for i := 'A'; i <= 'Z'; i++ {
		vol := fmt.Sprintf("%c:", i)
		if _, err := os.Stat(vol + "\\"); err == nil {
			vols = append(vols, vol)
		}
	}
	return vols
}

// IsNetworkFilesystem 对应 SystemUtils.is_network_filesystem，简单版
func IsNetworkFilesystem(directory string) bool {
	if directory == "" {
		return false
	}
	if IsWindows() {
		// 以 \\ 开头视为网络盘
		return strings.HasPrefix(directory, "\\\\")
	}

	cmd := exec.Command("df", "-T", directory)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	output := strings.ToLower(string(out))
	// 本地文件系统中含有 fuse 关键字的需要排除
	if strings.Contains(output, "fuse.shfs") || strings.Contains(output, "zfuse.zfsv") {
		return false
	}

	networkFS := []string{"nfs", "cifs", "smbfs", "fuse", "sshfs", "ftpfs"}
	for _, fs := range networkFS {
		if strings.Contains(output, fs) {
			return true
		}
	}
	return false
}

// IsSameDisk 对应 SystemUtils.is_same_disk，判断两个路径是否在同一磁盘
func IsSameDisk(src, dest string) bool {
	if src == "" || dest == "" {
		return false
	}

	// 清理路径
	srcPath := filepath.Clean(src)
	destPath := filepath.Clean(dest)

	// 获取源路径的信息
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false
	}

	// 获取目标路径的信息
	destInfo, err := os.Stat(destPath)
	if err != nil {
		return false
	}

	// Windows：使用驱动器号比较
	if IsWindows() {
		return strings.EqualFold(filepath.VolumeName(srcPath), filepath.VolumeName(destPath))
	}

	// 其他系统：使用设备号比较
	srcStat, srcOk := srcInfo.Sys().(*syscall.Stat_t)
	destStat, destOk := destInfo.Sys().(*syscall.Stat_t)

	if srcOk && destOk {
		return srcStat.Dev == destStat.Dev
	}

	// 回退方案：使用路径前缀比较
	return strings.HasPrefix(srcPath, string(os.PathSeparator)) && strings.HasPrefix(destPath, string(os.PathSeparator))
}

// GetConfigPath 对应 SystemUtils.get_config_path
func GetConfigPath(configDir string) string {
	if configDir == "" {
		configDir = os.Getenv("CONFIG_DIR")
	}
	if configDir != "" {
		return configDir
	}
	if IsDocker() {
		return "/config"
	}
	// Go 版本没有 Python frozen 的概念，这里统一认为是普通二进制，使用工作目录上级的 config
	cwd, err := os.Getwd()
	if err != nil {
		return "config"
	}
	return filepath.Join(cwd, "config")
}

// GetEnvPath 对应 SystemUtils.get_env_path
func GetEnvPath(configDir string) string {
	return filepath.Join(GetConfigPath(configDir), "app.env")
}

// ClearOldFiles 对应 SystemUtils.clear，删除指定目录中 N 天前的文件并清理空目录
func ClearOldFiles(tempPath string, days int) {
	info, err := os.Stat(tempPath)
	if err != nil || !info.IsDir() {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -days)

	// 删除旧文件
	_ = filepath.Walk(tempPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
		return nil
	})

	// 再遍历一遍删除空目录
	_ = filepath.Walk(tempPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() || path == tempPath {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		if len(entries) == 0 {
			_ = os.Remove(path)
		}
		return nil
	})
}

// GenerateUserUniqueID 对应 SystemUtils.generate_user_unique_id
// 按照 Python 版本逻辑，优先使用：1. 文件系统唯一标识符；2. MAC 地址；3. 主机名
func GenerateUserUniqueID() string {
	if id := filesystemUniqueID(); id != "" {
		return id
	}
	if id := macAddressID(); id != "" {
		return id
	}
	// 3. 主机名
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		sum := sha256.Sum256([]byte(hostname))
		return hex.EncodeToString(sum[:])
	}
	return ""
}

func filesystemUniqueID() string {
	// 使用根目录的设备号和 inode 生成唯一标识符
	rootPath := string(os.PathSeparator)
	info, err := os.Stat(rootPath)
	if err != nil {
		return ""
	}

	// 获取设备号和 inode
	var dev, ino uint64
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		dev = uint64(stat.Dev)
		ino = stat.Ino
	} else {
		// 如果无法获取设备号和 inode，退化为使用文件信息
		return ""
	}

	// 生成哈希
	data := fmt.Sprintf("%d-%d", dev, ino)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func macAddressID() string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifs {
		mac := iface.HardwareAddr.String()
		if mac == "" {
			continue
		}

		// 检查是否是虚拟 MAC 地址（第一个字节的第二个最低位为 1）
		macBytes := iface.HardwareAddr
		if len(macBytes) > 0 && (macBytes[0]&0x02) != 0 {
			// 虚拟 MAC 地址，跳过
			continue
		}

		// 生成哈希
		macStr := strings.ReplaceAll(mac, ":", "")
		sum := sha256.Sum256([]byte(macStr))
		return hex.EncodeToString(sum[:])
	}
	return ""
}
