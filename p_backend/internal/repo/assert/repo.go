package assert

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"monorepo/config"
	"monorepo/pkg/xerr"
	"monorepo/proto/xadminpb"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Repo struct {
	storagePath string
}

type FileData struct {
	Reader      io.ReadCloser
	ContentType string
	FileName    string
	Size        int64
}

const (
	fileCacheMaxItems = 100
	fileCacheMaxBytes = int64(200 * 1024 * 1024)
)

type cachedFile struct {
	Data        []byte
	ContentType string
	FileName    string
	Size        int64
	LastUsedAt  int64
}

var resourceFileCache = struct {
	sync.Mutex
	items      map[string]*cachedFile
	totalBytes int64
	clock      int64
}{
	items: make(map[string]*cachedFile),
}

type SavedFile struct {
	URL          string
	ContentType  string
	Extension    string
	Size         int64
	ResourceType string
}

func NewRepo() *Repo {
	return &Repo{
		storagePath: config.GetConfig().App.Server.StoragePath,
	}
}

func (r *Repo) SaveUploadedFile(ctx context.Context, uid int32, scene xadmin.UploadScene, fileHeader *multipart.FileHeader) (string, error) {
	saved, err := r.SaveUploadedFileMeta(ctx, uid, scene, fileHeader)
	if err != nil {
		return "", err
	}
	return saved.URL, nil
}

func (r *Repo) SaveUploadedFileMeta(_ context.Context, uid int32, scene xadmin.UploadScene, fileHeader *multipart.FileHeader) (*SavedFile, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "failed to open upload file")
	}
	defer src.Close()

	contentType, ext, err := detectUploadType(src)
	if err != nil {
		return nil, err
	}
	if err = validateSceneFileType(scene, contentType); err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%d_%d%s", uid, time.Now().UnixMilli(), ext)
	relPath := path.Join(scene.String(), fileName)
	dstPath, err := r.resolveStoragePath(relPath)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to prepare storage directory")
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to create upload file")
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to save upload file")
	}
	return &SavedFile{
		URL:         relPath,
		ContentType: contentType,
		Extension:   ext,
		Size:        size,
	}, nil
}

func (r *Repo) SaveResourceUploadedFileMeta(_ context.Context, uid int32, fileHeader *multipart.FileHeader, requestedResourceType string) (*SavedFile, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "failed to open upload file")
	}
	defer src.Close()

	contentType, ext, err := detectUploadType(src)
	if err != nil {
		return nil, err
	}
	scene, resourceType, limitBytes, err := resourceSceneByContentType(contentType, requestedResourceType)
	if err != nil {
		return nil, err
	}
	if fileHeader.Size > limitBytes {
		return nil, xerr.NewBiz(xerr.CodeBadRequest, "res.size_limit", limitBytes/1024/1024)
	}

	fileName := fmt.Sprintf("%d_%d%s", uid, time.Now().UnixMilli(), ext)
	relPath := path.Join(scene.String(), fileName)
	dstPath, err := r.resolveStoragePath(relPath)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to prepare storage directory")
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to create upload file")
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to save upload file")
	}
	return &SavedFile{
		URL:          relPath,
		ContentType:  contentType,
		Extension:    ext,
		Size:         size,
		ResourceType: resourceType,
	}, nil
}

func (r *Repo) OpenFile(_ context.Context, fileURL string) (*FileData, error) {
	if strings.TrimSpace(fileURL) == "" {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "file_url cannot be empty")
	}

	filePath, err := r.resolveStoragePath(fileURL)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, xerr.NewWithDetail(xerr.CodeNotFound, "file not found")
		}
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to stat file")
	}
	if info.IsDir() {
		return nil, xerr.NewWithDetail(xerr.CodeNotFound, "file not found")
	}

	cacheKey := filePath
	if cached := getCachedFile(cacheKey); cached != nil {
		return &FileData{
			Reader:      io.NopCloser(bytes.NewReader(cached.Data)),
			ContentType: cached.ContentType,
			FileName:    cached.FileName,
			Size:        cached.Size,
		}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, xerr.NewWithDetail(xerr.CodeNotFound, "file not found")
		}
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to open file")
	}

	contentType, err := detectContentType(filePath, file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	if info.Size() <= fileCacheMaxBytes {
		data, readErr := io.ReadAll(file)
		if readErr == nil {
			_ = file.Close()
			setCachedFile(cacheKey, &cachedFile{
				Data:        data,
				ContentType: contentType,
				FileName:    filepath.Base(filePath),
				Size:        info.Size(),
			})
			return &FileData{
				Reader:      io.NopCloser(bytes.NewReader(data)),
				ContentType: contentType,
				FileName:    filepath.Base(filePath),
				Size:        info.Size(),
			}, nil
		}
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			_ = file.Close()
			return nil, xerr.NewWithDetail(xerr.CodeInternalError, "failed to rewind file")
		}
	}

	return &FileData{
		Reader:      file,
		ContentType: contentType,
		FileName:    filepath.Base(filePath),
		Size:        info.Size(),
	}, nil
}

func (r *Repo) FileExists(_ context.Context, fileURL string) bool {
	if strings.TrimSpace(fileURL) == "" {
		return false
	}
	filePath, err := r.resolveStoragePath(fileURL)
	if err != nil {
		return false
	}
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func getCachedFile(key string) *cachedFile {
	resourceFileCache.Lock()
	defer resourceFileCache.Unlock()
	item := resourceFileCache.items[key]
	if item == nil {
		return nil
	}
	resourceFileCache.clock++
	item.LastUsedAt = resourceFileCache.clock
	data := make([]byte, len(item.Data))
	copy(data, item.Data)
	return &cachedFile{
		Data:        data,
		ContentType: item.ContentType,
		FileName:    item.FileName,
		Size:        item.Size,
	}
}

func setCachedFile(key string, item *cachedFile) {
	if item == nil || item.Size > fileCacheMaxBytes {
		return
	}
	resourceFileCache.Lock()
	defer resourceFileCache.Unlock()
	if old := resourceFileCache.items[key]; old != nil {
		resourceFileCache.totalBytes -= old.Size
	}
	resourceFileCache.clock++
	item.LastUsedAt = resourceFileCache.clock
	resourceFileCache.items[key] = item
	resourceFileCache.totalBytes += item.Size
	for len(resourceFileCache.items) > fileCacheMaxItems || resourceFileCache.totalBytes > fileCacheMaxBytes {
		evictOldestCachedFile()
	}
}

func evictOldestCachedFile() {
	var oldestKey string
	var oldestAt int64
	for key, item := range resourceFileCache.items {
		if oldestKey == "" || item.LastUsedAt < oldestAt {
			oldestKey = key
			oldestAt = item.LastUsedAt
		}
	}
	if oldestKey == "" {
		return
	}
	resourceFileCache.totalBytes -= resourceFileCache.items[oldestKey].Size
	delete(resourceFileCache.items, oldestKey)
}

func (r *Repo) resolveStoragePath(fileURL string) (string, error) {
	cleanURL := path.Clean("/" + strings.TrimSpace(fileURL))

	root, err := filepath.Abs(r.storagePath)
	if err != nil {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "invalid storage path")
	}

	target := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(cleanURL, "/")))
	target, err = filepath.Abs(target)
	if err != nil {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "invalid file path")
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", xerr.NewWithDetail(xerr.CodeForbidden, "file path is not allowed")
	}
	return target, nil
}

func detectContentType(filePath string, file *os.File) (string, error) {
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	if contentType != "" {
		return contentType, nil
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "failed to detect file content-type")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "failed to rewind file")
	}
	return http.DetectContentType(buf[:n]), nil
}

func detectUploadType(file multipart.File) (string, string, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", "", xerr.NewWithDetail(xerr.CodeBadRequest, "failed to read upload file")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", "", xerr.NewWithDetail(xerr.CodeInternalError, "failed to rewind upload file")
	}

	contentType := detectUploadContentType(buf[:n])
	ext, ok := extByDetectedContentType(contentType)
	if !ok {
		return "", "", xerr.NewWithDetail(xerr.CodeBadRequest, "unsupported file type: %s", contentType)
	}
	return contentType, ext, nil
}

func detectUploadContentType(header []byte) string {
	headerText := string(header)
	brand := ""
	if len(header) >= 12 {
		brand = string(header[4:12])
	}
	riffType := ""
	if len(header) >= 12 {
		riffType = string(header[8:12])
	}

	switch {
	case startsWithSignature(header, []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg"
	case startsWithSignature(header, []byte{0x89, 0x50, 0x4e, 0x47}):
		return "image/png"
	case strings.HasPrefix(headerText, "GIF87a") || strings.HasPrefix(headerText, "GIF89a"):
		return "image/gif"
	case strings.HasPrefix(headerText, "RIFF") && riffType == "WEBP":
		return "image/webp"
	case strings.Contains(brand, "avif") || strings.Contains(brand, "avis"):
		return "image/avif"
	case strings.HasPrefix(headerText, "ID3") || isMPEGAudioFrame(header):
		return "audio/mpeg"
	case strings.HasPrefix(headerText, "RIFF") && riffType == "WAVE":
		return "audio/wav"
	case strings.HasPrefix(headerText, "OggS"):
		return "audio/ogg"
	case strings.HasPrefix(headerText, "fLaC"):
		return "audio/flac"
	case strings.Contains(brand, "M4A") || strings.Contains(brand, "mp42"):
		return "audio/mp4"
	case strings.Contains(brand, "mp4") || strings.Contains(brand, "isom"):
		return "video/mp4"
	case strings.Contains(brand, "qt  "):
		return "video/quicktime"
	case startsWithSignature(header, []byte{0x1a, 0x45, 0xdf, 0xa3}):
		return "video/webm"
	case strings.HasPrefix(headerText, "RIFF") && riffType == "AVI ":
		return "video/x-msvideo"
	case strings.HasPrefix(headerText, "%PDF"):
		return "application/pdf"
	case startsWithSignature(header, []byte{0xd0, 0xcf, 0x11, 0xe0}):
		return "application/msword"
	case startsWithSignature(header, []byte{0x50, 0x4b, 0x03, 0x04}):
		return "application/zip"
	case startsWithSignature(header, []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07}):
		return "application/vnd.rar"
	case startsWithSignature(header, []byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}):
		return "application/x-7z-compressed"
	case startsWithSignature(header, []byte{0x1f, 0x8b}):
		return "application/gzip"
	case isPlainTextHeader(header):
		return "text/plain; charset=utf-8"
	default:
		return http.DetectContentType(header)
	}
}

func startsWithSignature(header []byte, signature []byte) bool {
	if len(header) < len(signature) {
		return false
	}
	for index, expected := range signature {
		if header[index] != expected {
			return false
		}
	}
	return true
}

func isMPEGAudioFrame(header []byte) bool {
	return len(header) >= 2 && header[0] == 0xff && header[1]&0xe0 == 0xe0
}

func isPlainTextHeader(header []byte) bool {
	if len(header) == 0 {
		return false
	}
	for _, byteValue := range header {
		if byteValue == 9 || byteValue == 10 || byteValue == 13 || (byteValue >= 32 && byteValue <= 126) {
			continue
		}
		return false
	}
	return true
}

func resourceSceneByContentType(contentType string, requestedResourceType string) (xadmin.UploadScene, string, int64, error) {
	const mb = int64(1024 * 1024)
	if contentType == "application/zip" && strings.EqualFold(strings.TrimSpace(requestedResourceType), "archive") {
		return xadmin.UploadScene_US_ResourceArchive, "archive", 100 * mb, nil
	}
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/avif":
		return xadmin.UploadScene_US_ResourceImage, "image", 5 * mb, nil
	case "audio/mpeg", "audio/wav", "audio/ogg", "audio/flac", "audio/mp4":
		return xadmin.UploadScene_US_ResourceAudio, "audio", 5 * mb, nil
	case "video/mp4", "video/quicktime", "video/webm", "video/x-msvideo":
		return xadmin.UploadScene_US_ResourceVideo, "video", 100 * mb, nil
	case "application/pdf", "application/msword", "application/zip", "text/plain", "text/plain; charset=utf-8":
		return xadmin.UploadScene_US_ResourceDocument, "document", 20 * mb, nil
	case "application/vnd.rar", "application/x-7z-compressed", "application/gzip":
		return xadmin.UploadScene_US_ResourceArchive, "archive", 100 * mb, nil
	default:
		return xadmin.UploadScene_US_Unknown, "", 0, xerr.NewBiz(xerr.CodeBadRequest, "res.unsupported_type")
	}
}

func validateSceneFileType(scene xadmin.UploadScene, contentType string) error {
	allowedByScene := map[xadmin.UploadScene]map[string]struct{}{
		xadmin.UploadScene_US_Avatar: {
			"image/jpeg": {},
			"image/png":  {},
		},
		xadmin.UploadScene_US_BarLogo: {
			"image/jpeg": {},
			"image/png":  {},
			"image/webp": {},
		},
		xadmin.UploadScene_US_BarAlbum: {
			"image/jpeg": {},
			"image/png":  {},
			"image/webp": {},
		},
		xadmin.UploadScene_US_EventCover: {
			"image/jpeg": {},
			"image/png":  {},
			"image/webp": {},
		},
		xadmin.UploadScene_US_EventMisc: {
			"image/jpeg": {},
			"image/png":  {},
			"image/webp": {},
		},
		xadmin.UploadScene_US_ResourceImage: {
			"image/jpeg": {},
			"image/png":  {},
			"image/gif":  {},
			"image/webp": {},
			"image/avif": {},
		},
		xadmin.UploadScene_US_ResourceAudio: {
			"audio/mpeg": {},
			"audio/wav":  {},
			"audio/ogg":  {},
			"audio/flac": {},
			"audio/mp4":  {},
		},
		xadmin.UploadScene_US_ResourceVideo: {
			"video/mp4":       {},
			"video/quicktime": {},
			"video/webm":      {},
			"video/x-msvideo": {},
		},
		xadmin.UploadScene_US_ResourceDocument: {
			"application/pdf":           {},
			"application/msword":        {},
			"application/zip":           {},
			"text/plain; charset=utf-8": {},
			"text/plain":                {},
		},
		xadmin.UploadScene_US_ResourceArchive: {
			"application/zip":             {},
			"application/vnd.rar":         {},
			"application/x-7z-compressed": {},
			"application/gzip":            {},
		},
	}

	allowedTypes, ok := allowedByScene[scene]
	if !ok {
		return xerr.NewWithDetail(xerr.CodeBadRequest, "unsupported upload scene: %s", scene.String())
	}
	if _, ok = allowedTypes[contentType]; !ok {
		return xerr.NewWithDetail(xerr.CodeBadRequest, "scene %s does not allow file type %s", scene.String(), contentType)
	}
	return nil
}

func extByDetectedContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	case "image/avif":
		return ".avif", true
	case "audio/mpeg":
		return ".mp3", true
	case "audio/wav":
		return ".wav", true
	case "audio/ogg":
		return ".ogg", true
	case "audio/flac":
		return ".flac", true
	case "audio/mp4":
		return ".m4a", true
	case "video/mp4":
		return ".mp4", true
	case "video/quicktime":
		return ".mov", true
	case "video/webm":
		return ".webm", true
	case "video/x-msvideo":
		return ".avi", true
	case "application/pdf":
		return ".pdf", true
	case "application/msword":
		return ".doc", true
	case "application/zip":
		return ".zip", true
	case "application/vnd.rar":
		return ".rar", true
	case "application/x-7z-compressed":
		return ".7z", true
	case "application/gzip":
		return ".gz", true
	case "text/plain; charset=utf-8", "text/plain":
		return ".txt", true
	default:
		return "", false
	}
}
