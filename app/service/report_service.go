package service

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/repository"
)

type ReportService struct {
	reportRepo     *repository.ReportRepository
	mongoReportRepo *repository.MongoReportRepository
	studentRepo    *repository.StudentRepository
}

func NewReportService(
	reportRepo *repository.ReportRepository,
	mongoReportRepo *repository.MongoReportRepository,
	studentRepo *repository.StudentRepository,
) *ReportService {
	return &ReportService{
		reportRepo:     reportRepo,
		mongoReportRepo: mongoReportRepo,
		studentRepo:    studentRepo,
	}
}

func (s *ReportService) Statistics(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*model.JWTClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse("unauthorized", nil))
	}

	var studentIDs []string

	switch claims.Role {
	case "Mahasiswa":
		student, err := s.studentRepo.FindByUserID(claims.UserID)
		if err != nil || student == nil {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse("student not found", nil))
		}
		studentIDs = []string{student.ID}

	case "Dosen Wali":
		students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed fetch advisees", err.Error()))
		}
		for _, st := range students {
			studentIDs = append(studentIDs, st.ID)
		}
	case "Admin":
	default:
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse("forbidden role", nil))
	}

	ctx := context.Background()

	statusCounts, err := s.reportRepo.CountByStatus(studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed count by status", err.Error()))
	}

	typeCounts, err := s.mongoReportRepo.GetTotalByType(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed total by type", err.Error()))
	}

	competitionDist, err := s.mongoReportRepo.GetCompetitionLevelDistribution(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed competition distribution", err.Error()))
	}

	topStudents, err := s.mongoReportRepo.GetTopStudentsByPoints(ctx, int64(10), studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed top students", err.Error()))
	}

	monthly, err := s.mongoReportRepo.GetMonthlyCounts(ctx, 6, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed monthly counts", err.Error()))
	}

	topByCount, err := s.reportRepo.CountTotalPerStudent(studentIDs, 10)
	if err != nil {
		fmt.Println("warning: CountTotalPerStudent failed:", err)
	}

	resp := fiber.Map{
		"totalByStatus":    statusCounts,
		"totalByType":      typeCounts,
		"competitionLevels": competitionDist,
		"topStudentsByPoints": topStudents,
		"monthly":          monthly,
		"topByCount":       topByCount,
	}

	return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(resp))
}

func (s *ReportService) StudentStatistics(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*model.JWTClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse("unauthorized", nil))
	}

	studentID := c.Params("id")
	if studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse("student id is required", nil))
	}

	if claims.Role == "Mahasiswa" {
		st, err := s.studentRepo.FindByUserID(claims.UserID)
		if err != nil || st == nil {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse("student not found", nil))
		}
		if st.ID != studentID {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse("forbidden: not your data", nil))
		}
	} else if claims.Role == "Dosen Wali" {
		students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed fetch advisees", err.Error()))
		}
		found := false
		for _, sst := range students {
			if sst.ID == studentID {
				found = true
				break
			}
		}
		if !found {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse("forbidden: student not under your supervision", nil))
		}
	} else if claims.Role != "Admin" {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse("forbidden role", nil))
	}

	ctx := context.Background()
	studentIDs := []string{studentID}

	statusCounts, err := s.reportRepo.CountByStatus(studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed count by status", err.Error()))
	}

	typeCounts, err := s.mongoReportRepo.GetTotalByType(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed total by type", err.Error()))
	}

	competitionDist, err := s.mongoReportRepo.GetCompetitionLevelDistribution(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed competition distribution", err.Error()))
	}

	topPoints, err := s.mongoReportRepo.GetTopStudentsByPoints(ctx, 5, studentIDs) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed top points", err.Error()))
	}

	monthly, err := s.mongoReportRepo.GetMonthlyCounts(ctx, 12, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse("failed monthly counts", err.Error()))
	}

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil || student == nil {
		student = &model.Student{ID: studentID}
	}

	resp := fiber.Map{
		"student": student,
		"statistics": fiber.Map{
			"totalByStatus":    statusCounts,
			"totalByType":      typeCounts,
			"competitionLevels": competitionDist,
			"topPoints":        topPoints,
			"monthly":          monthly,
		},
	}

	return c.Status(fiber.StatusOK).JSON(model.SuccessResponse(resp))
}
