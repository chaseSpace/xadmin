package assert

import (
	"monorepo/internal/middleware"
	assertrepo "monorepo/internal/repo/assert"
	resourcerepo "monorepo/internal/repo/resource"
	"monorepo/pkg/xerr"
	"monorepo/proto/xadminpb"

	"github.com/gofiber/fiber/v2"
)

type Service struct {
	repo         *assertrepo.Repo
	resourceRepo *resourcerepo.Repo
}

func NewService() *Service {
	return &Service{repo: assertrepo.NewRepo(), resourceRepo: resourcerepo.NewRepo()}
}

func (s *Service) UploadFile(c *fiber.Ctx, req *xadmin.UploadFileReq, fileField string) (*xadmin.UploadFileResp, error) {
	if req.GetScene() == xadmin.UploadScene_US_Unknown {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "scene is required")
	}

	fileHeader, err := c.FormFile(fileField)
	if err != nil {
		return nil, xerr.NewWithDetail(xerr.CodeBadRequest, "missing upload file")
	}

	fileURL, err := s.repo.SaveUploadedFile(c.UserContext(), middleware.MustUID(c), req.GetScene(), fileHeader)
	if err != nil {
		return nil, err
	}
	return &xadmin.UploadFileResp{FileUrl: fileURL}, nil
}

func (s *Service) GetFile(c *fiber.Ctx, req *xadmin.GetFileReq) error {
	resourceFile, err := s.resourceRepo.GetFileByURL(c.UserContext(), req.GetFileUrl())
	if err == nil && resourceFile.RequireAuth && middleware.GetUID(c) <= 0 {
		return xerr.NewWithDetail(xerr.CodeUnauthorized, "file access requires authorization")
	}

	fileData, err := s.repo.OpenFile(c.UserContext(), req.GetFileUrl())
	if err != nil {
		return err
	}

	// SendStream registers the reader on the response and fasthttp closes it after flushing.
	c.Set(fiber.HeaderContentType, fileData.ContentType)
	if !req.IsDownload {
		return c.SendStream(fileData.Reader, int(fileData.Size))
	}

	c.Attachment(fileData.FileName)
	return c.SendStream(fileData.Reader, int(fileData.Size))
}
