package service

import (
	"context"
	"fmt"
	"time"

	"go-fiber/app/model"
	"go-fiber/app/repository"

	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AchievementService integrates Postgres + Mongo repos
type AchievementService struct {
	postgresRepo *repository.AchievementRepository
	mongoRepo    *repository.MongoAchievementRepository
	studentRepo  *repository.StudentRepository
}

func NewAchievementService(
	postgresRepo *repository.AchievementRepository,
	mongoRepo *repository.MongoAchievementRepository,
	studentRepo *repository.StudentRepository,
) *AchievementService {
	return &AchievementService{
		postgresRepo: postgresRepo,
		mongoRepo:    mongoRepo,
		studentRepo:  studentRepo,
	}
}

// POST /achievements
func (s *AchievementService) Create(c *fiber.Ctx) error {
	// 1. Gunakan Struct Request Khusus
	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	claims, ok := c.Locals("user").(*model.JWTClaims)
	if !ok {
		return c.Status(401).JSON(model.ErrorResponse("unauthorized", nil))
	}

	student, err := s.studentRepo.FindByUserID(claims.UserID)
	if err != nil || student == nil {
		return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
	}

	// 2. Mapping dari Request DTO ke Entity Model (untuk Database)
	achievementData := model.Achievement{
		StudentID:       student.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		Points:          0, // Default 0
		Attachments:     []model.Attachment{}, // Inisialisasi slice kosong
	}

	// Insert Mongo
	ctx := context.Background()
	// Repository tetap menerima model.Achievement
	objID, err := s.mongoRepo.CreateAchievement(ctx, &achievementData)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed create achievement (mongo)", err.Error()))
	}

	// Insert Postgre Reference
	ref := &model.AchievementReference{
		ID:                 uuid.New().String(),
		StudentID:          student.ID,
		MongoAchievementID: objID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.postgresRepo.CreateReference(ref); err != nil {
		_ = s.mongoRepo.DeleteAchievement(ctx, objID)
		return c.Status(500).JSON(model.ErrorResponse("failed create achievement reference", err.Error()))
	}

	return c.JSON(model.SuccessResponse(fiber.Map{
		"reference":     ref,
		"achievementId": objID.Hex(),
	}))
}


// GET /achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	// Gunakan struct filter jika ingin lebih rapi (opsional)
	filter := new(model.FilterAchievementRequest)
	if err := c.QueryParser(filter); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid query param", err.Error()))
	}

    var refs []model.AchievementReference
    var err error

    if filter.StudentID != "" {
        refs, err = s.postgresRepo.FindByStudentID(filter.StudentID)
    } else {
        refs, err = s.postgresRepo.FindAll()
    }
    
    if err != nil {
        c.Status(fiber.StatusInternalServerError)
        return c.JSON(model.ErrorResponse("failed fetch references", err.Error()))
    }

    ctx := context.Background()
    results := []fiber.Map{}
    for _, r := range refs {
        var ach *model.Achievement
        objID, convErr := primitive.ObjectIDFromHex(r.MongoAchievementID)
        if convErr == nil {
            ach, _ = s.mongoRepo.FindByID(ctx, objID) 
        }
        results = append(results, fiber.Map{
            "reference":   r,
            "achievement": ach,
        })
    }

    c.Status(fiber.StatusOK)
    return c.JSON(model.SuccessResponse(results))
}


// GET /achievements/:id
func (s *AchievementService) GetByID(c *fiber.Ctx) error {
    id := c.Params("id")
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        c.Status(fiber.StatusNotFound)
        return c.JSON(model.ErrorResponse("reference not found", nil))
    }

    objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
    if err != nil {
        c.Status(fiber.StatusInternalServerError)
        return c.JSON(model.ErrorResponse("invalid mongo id", err.Error()))
    }

    ach, err := s.mongoRepo.FindByID(context.Background(), objID)
    if err != nil {
        c.Status(fiber.StatusInternalServerError)
        return c.JSON(model.ErrorResponse("failed fetch achievement", err.Error()))
    }

    c.Status(fiber.StatusOK)
    return c.JSON(model.SuccessResponse(fiber.Map{
        "reference":   ref,
        "achievement": ach,
    }))
}


// PUT /achievements/:id
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	if err := s.checkOwnership(c, ref.StudentID); err != nil {
		return err
	}

	// 1. Gunakan Struct UpdateRequest
	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("invalid mongo id", err.Error()))
	}

	update := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"details":     req.Details,
		"tags":        req.Tags,
		// "points": req.Points, // Poin biasanya tidak diupdate manual user
	}

	if err := s.mongoRepo.UpdateAchievement(context.Background(), objID, update); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed update achievement", err.Error()))
	}

	return c.JSON(model.SuccessResponse("achievement updated"))
}

// POST /achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
    id := c.Params("id")

    // 1. Gunakan Struct RejectRequest
    var req model.RejectAchievementRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
    }

    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }

    // 2. Gunakan data dari request struct (req.Note)
    if err := s.postgresRepo.UpdateReferenceStatus(
        id, "rejected", nil, nil, nil, &req.Note,
    ); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to reject", err.Error()))
    }

    c.Status(fiber.StatusOK)
    return c.JSON(model.SuccessResponse("achievement rejected"))
}

// DELETE /achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
    id := c.Params("id")

    // 1. Cari reference
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(fiber.StatusNotFound).
            JSON(model.ErrorResponse("reference not found", nil))
    }

    if err := s.checkOwnership(c, ref.StudentID); err != nil {
        return c.Status(fiber.StatusForbidden).
            JSON(model.ErrorResponse("you are not the owner of this achievement", nil))
    }

    if ref.Status != "draft" {
        return c.Status(fiber.StatusForbidden).
            JSON(model.ErrorResponse("cannot delete: achievement already submitted", nil))
    }

    if objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID); err == nil {
        _ = s.mongoRepo.DeleteAchievement(context.Background(), objID)
    }

    note := "soft deleted by student"
    if err := s.postgresRepo.UpdateReferenceStatus(
        id,
        "deleted",
        nil,
        nil,
        nil,
        &note,
    ); err != nil {
        return c.Status(fiber.StatusInternalServerError).
            JSON(model.ErrorResponse("failed to delete achievement", err.Error()))
    }

    return c.Status(fiber.StatusOK).
        JSON(model.SuccessResponse("achievement deleted"))
}

// POST /achievements/:id/submit
func (s *AchievementService) Submit(c *fiber.Ctx) error {
    id := c.Params("id")
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }
    if err := s.checkOwnership(c, ref.StudentID); err != nil {
        return err
    }
    now := time.Now()
    if err := s.postgresRepo.UpdateReferenceStatus(
        id, "submitted", &now, nil, nil, nil,
    ); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to submit", err.Error()))
    }
    return c.JSON(model.SuccessResponse("achievement submitted"))
}

// POST /achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
    id := c.Params("id")
    claims, _ := c.Locals("user").(*model.JWTClaims)
    userID := claims.UserID
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }
    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }
    now := time.Now()
    if err := s.postgresRepo.UpdateReferenceStatus(
        id, "verified", nil, &now, &userID, nil,
    ); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to verify", err.Error()))
    }
    c.Status(fiber.StatusOK)
    return c.JSON(model.SuccessResponse("achievement verified"))
}

func (s *AchievementService) List(c *fiber.Ctx) error {
    claims := c.Locals("user").(*model.JWTClaims)
    if claims.Role == "Mahasiswa" {
        student, err := s.studentRepo.FindByUserID(claims.UserID)
        if err != nil {
            return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
        }
        refs, err := s.postgresRepo.FindByStudentID(student.ID)
        if err != nil {
            return c.Status(500).JSON(model.ErrorResponse("failed to fetch achievements", err.Error()))
        }
        return c.JSON(model.SuccessResponse(refs))
    }
    if claims.Role == "Dosen Wali" {
        students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
        if err != nil {
            return c.Status(500).JSON(model.ErrorResponse("failed to fetch advisory students", nil))
        }
        var allRefs []model.AchievementReference
        for _, stu := range students {
            refs, _ := s.postgresRepo.FindByStudentID(stu.ID)
            allRefs = append(allRefs, refs...)
        }
        return c.JSON(model.SuccessResponse(allRefs))
    }
    refs, err := s.postgresRepo.FindAll()
    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to fetch achievements", err.Error()))
    }
    return c.JSON(model.SuccessResponse(refs))
}

func (s *AchievementService) Detail(c *fiber.Ctx) error {
    id := c.Params("id")
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }
    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }
    objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid mongo id", nil))
    }
    ach, err := s.mongoRepo.FindByID(context.Background(), objID)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("achievement not found in mongo", nil))
    }
    return c.JSON(model.SuccessResponse(fiber.Map{
        "reference": ref,
        "achievement": ach,
    }))
}

func (s *AchievementService) History(c *fiber.Ctx) error {
    id := c.Params("id")
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }
    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }
    c.Status(fiber.StatusOK)
    return c.JSON(model.SuccessResponse(fiber.Map{
        "id":          ref.ID,
        "status":      ref.Status,
        "submittedAt": ref.SubmittedAt,
        "verifiedAt":  ref.VerifiedAt,
        "rejection":   ref.RejectionNote,
    }))
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
    id := c.Params("id")
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("file is required", err.Error()))
    }
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    savePath := fmt.Sprintf("./uploads/%s", filename)
    if err := c.SaveFile(file, savePath); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to save file", err.Error()))
    }
    attachment := model.Attachment{
        FileName:   file.Filename,
        FileURL:    "/uploads/" + filename,
        FileType:   file.Header.Get("Content-Type"),
        UploadedAt: time.Now(),
    }
    objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid mongo id", nil))
    }
    if err := s.mongoRepo.AddAttachment(context.Background(), objID, attachment); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to update mongo", err.Error()))
    }
    return c.JSON(model.SuccessResponse(fiber.Map{
        "message": "file uploaded successfully",
        "file":    attachment,
    }))
}

func (s *AchievementService) checkOwnership(c *fiber.Ctx, refStudentID string) error {
    claims := c.Locals("user").(*model.JWTClaims)
    if claims.Role != "Mahasiswa" {
        return nil 
    }
    student, err := s.studentRepo.FindByUserID(claims.UserID)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
    }
    if student.ID != refStudentID {
        return c.Status(403).JSON(model.ErrorResponse("forbidden: not owner of the achievement", nil))
    }
    return nil
}

func (s *AchievementService) checkAdvisor(c *fiber.Ctx, refStudentID string) error {
    claims := c.Locals("user").(*model.JWTClaims)
    if claims.Role == "Admin" {
        return nil
    }
    if claims.Role != "Dosen Wali" {
        return nil
    }
    students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to fetch advisory students", nil))
    }
    for _, s := range students {
        if s.ID == refStudentID {
            return nil 
        }
    }
    return c.Status(403).JSON(model.ErrorResponse("forbidden: student is not under your supervision", nil))
}

func (s *AchievementService) GetByStudentID(studentID string) ([]fiber.Map, error) {
    refs, err := s.postgresRepo.FindByStudentID(studentID)
    if err != nil {
        return nil, err
    }
    results := []fiber.Map{}
    ctx := context.Background()
    for _, ref := range refs {
        objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
        if err != nil {
            continue
        }
        ach, err := s.mongoRepo.FindByID(ctx, objID)
        if err != nil {
            continue
        }
        results = append(results, fiber.Map{
            "referenceId": ref.ID,
            "status":      ref.Status,
            "createdAt":   ref.CreatedAt,
            "updatedAt":   ref.UpdatedAt,
            "achievement": ach,
        })
    }
    return results, nil
}

func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
    studentID := c.Params("id")
    refs, err := s.postgresRepo.FindByStudentID(studentID)
    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to fetch references", err.Error()))
    }
    ctx := context.Background()
    results := []fiber.Map{}
    for _, ref := range refs {
        objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
        if err != nil {
            continue
        }
        ach, err := s.mongoRepo.FindByID(ctx, objID)
        if err != nil {
            continue
        }
        results = append(results, fiber.Map{
            "referenceId": ref.ID,
            "status":      ref.Status,
            "createdAt":   ref.CreatedAt,
            "updatedAt":   ref.UpdatedAt,
            "achievement": ach,
        })
    }
    return c.JSON(model.SuccessResponse(results))
}