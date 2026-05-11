package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dinkelspiel/cdn/dao"
	"github.com/dinkelspiel/cdn/db"
	"github.com/dinkelspiel/cdn/models"
	"github.com/dinkelspiel/cdn/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
)

func getTeamProjectBasePath(teamProject models.TeamProject) string {
	config, _ := LoadConfig()
	return filepath.Clean(fmt.Sprintf("%s/public/%d/%d", config.StorageUrl, *teamProject.Team.Id, *teamProject.Id))
}

func resolveTeamProjectPath(teamProject models.TeamProject, relativePath string) (string, error) {
	baseDir := getTeamProjectBasePath(teamProject)
	fullPath := filepath.Clean(filepath.Join(baseDir, relativePath))

	relativeToBase, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return "", err
	}

	if relativeToBase == ".." || strings.HasPrefix(relativeToBase, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("access to path '%s' is denied", fullPath)
	}

	return fullPath, nil
}

func GetTeamProjectFiles(teamProject models.TeamProject, relativePath string) ([]storage.FSItem, error) {
	fullPath, err := resolveTeamProjectPath(teamProject, relativePath)
	if err != nil {
		return nil, err
	}

	return storage.ListFiles(fullPath)
}

type TeamProjectImage struct {
	Path            string
	Name            string
	Size            int64
	RotationDegrees int
	LastModified    time.Time
}

func normalizeTeamProjectImagePath(teamProject models.TeamProject, relativePath string) (string, error) {
	basePath := getTeamProjectBasePath(teamProject)
	fullPath, err := resolveTeamProjectPath(teamProject, relativePath)
	if err != nil {
		return "", err
	}

	filePath, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(filePath), nil
}

func normalizeRotationDegrees(rotationDegrees int) int {
	rotationDegrees = rotationDegrees % 360
	if rotationDegrees < 0 {
		rotationDegrees += 360
	}
	return rotationDegrees
}

func GetTeamProjectImageRotation(db *db.DB, teamProject models.TeamProject, relativePath string) (int, error) {
	filePath, err := normalizeTeamProjectImagePath(teamProject, relativePath)
	if err != nil {
		return 0, err
	}

	imageRotation, err := dao.GetImageRotationByProjectAndPath(db, *teamProject.Id, filePath)
	if err != nil || imageRotation == nil {
		return 0, err
	}

	return normalizeRotationDegrees(imageRotation.RotationDegrees), nil
}

func RotateTeamProjectImageClockwise(db *db.DB, teamProject models.TeamProject, relativePath string) (*models.ImageRotation, error) {
	filePath, err := normalizeTeamProjectImagePath(teamProject, relativePath)
	if err != nil {
		return nil, err
	}
	if !isAlbumImage(filePath) {
		return nil, fmt.Errorf("'%s' is not a supported album image", filePath)
	}

	fullPath, err := resolveTeamProjectPath(teamProject, relativePath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("'%s' is a directory", filePath)
	}

	return dao.IncrementImageRotationClockwise(db, *teamProject.Id, filePath)
}

func GetTeamProjectImages(db *db.DB, teamProject models.TeamProject) ([]TeamProjectImage, error) {
	basePath, err := resolveTeamProjectPath(teamProject, "/")
	if err != nil {
		return nil, err
	}
	rotations, err := dao.GetImageRotationsByProject(db, *teamProject.Id)
	if err != nil {
		return nil, err
	}

	images := []TeamProjectImage{}
	err = filepath.WalkDir(basePath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !isAlbumImage(entry.Name()) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		images = append(images, TeamProjectImage{
			Path:            filepath.ToSlash(relativePath),
			Name:            entry.Name(),
			Size:            info.Size(),
			RotationDegrees: normalizeRotationDegrees(rotations[filepath.ToSlash(relativePath)]),
			LastModified:    info.ModTime(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(images, func(i, j int) bool {
		return strings.ToLower(images[i].Path) < strings.ToLower(images[j].Path)
	})

	return images, nil
}

func isAlbumImage(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func SerializeTeamProjectImage(image TeamProjectImage) gin.H {
	return gin.H{
		"path":            image.Path,
		"name":            image.Name,
		"size":            image.Size,
		"rotationDegrees": image.RotationDegrees,
		"lastModified":    image.LastModified.Format(time.RFC3339),
	}
}

func GetFilePathToFileInTeamProject(teamProject models.TeamProject, relativePath string) (string, error) {
	return resolveTeamProjectPath(teamProject, relativePath)
}

func GetTeamProjectFile(teamProject models.TeamProject, relativePath string) ([]byte, error) {
	path, err := GetFilePathToFileInTeamProject(teamProject, relativePath)
	if err != nil {
		return nil, err
	}

	return storage.ReadFile(path)
}

func EnsureFoldersExists() error {
	config, _ := LoadConfig()
	err := os.MkdirAll(config.StorageUrl, os.ModePerm)
	if err != nil {
		return err
	}

	publicDir := fmt.Sprintf("%s/public", config.StorageUrl)
	err = os.MkdirAll(publicDir, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func CreateTeamFolder(team models.Team) error {
	config, _ := LoadConfig()

	teamDir := filepath.Clean(fmt.Sprintf("%s/public/%d", config.StorageUrl, *team.Id))
	return os.MkdirAll(teamDir, os.ModePerm)
}

func CreateTeamProjectFolder(teamProject models.TeamProject, relativePath string) (string, error) {
	fullPath, err := GetFilePathToFileInTeamProject(teamProject, relativePath)
	if err != nil {
		return "", err
	}

	return fullPath, os.MkdirAll(fullPath, os.ModePerm)
}

type DiskStats struct {
	Total           uint64
	Free            uint64
	HostUsed        uint64
	DunlinFilesUsed uint64
	DunlinCacheUsed uint64
}

func GetHostDiskStats() DiskStats {
	config, _ := LoadConfig()

	var stat unix.Statfs_t
	err := unix.Statfs(config.HostRoot, &stat)
	if err != nil {
		panic(err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	hostUsed := total - free

	dunlinFilesUsed := getDirectorySize(filepath.Join(config.StorageUrl, "public"))
	dunlinCacheUsed := getDirectorySize(filepath.Join(config.StorageUrl, ".cache"))

	return DiskStats{
		Total:           total,
		Free:            free,
		HostUsed:        hostUsed - dunlinFilesUsed - dunlinCacheUsed,
		DunlinFilesUsed: dunlinFilesUsed,
		DunlinCacheUsed: dunlinCacheUsed,
	}
}

func getDirectorySize(path string) uint64 {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return size
}

type TeamProjectSize struct {
	TeamProject models.TeamProject
	Size        uint64
}

func GetSizeOfTeamProjectsOnDisk(db *db.DB) (*[]TeamProjectSize, error) {
	teamProjects, err := dao.GetTeamProjects(db)
	if err != nil {
		return nil, err
	}

	var teamProjectsSize []TeamProjectSize

	for _, teamProject := range *teamProjects {
		path, err := GetFilePathToFileInTeamProject(teamProject, "/")
		if err != nil {
			return nil, err
		}

		size := getDirectorySize(path)
		teamProjectsSize = append(teamProjectsSize, TeamProjectSize{
			Size:        size,
			TeamProject: teamProject,
		})

	}

	return &teamProjectsSize, nil
}

func SerializeTeamProjectSize(teamProjectSize TeamProjectSize) gin.H {
	return gin.H{
		"teamProject": models.SerializeTeamProject(teamProjectSize.TeamProject),
		"size":        teamProjectSize.Size,
	}
}

func SerializeDiskStats(diskStats DiskStats) gin.H {
	return gin.H{
		"total":           diskStats.Total,
		"free":            diskStats.Free,
		"hostUsed":        diskStats.HostUsed,
		"dunlinFilesUsed": diskStats.DunlinFilesUsed,
		"dunlinCacheUsed": diskStats.DunlinCacheUsed,
	}
}
