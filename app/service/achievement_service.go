package service

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go-fiber/app/model"
	"go-fiber/app/repository"

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
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	// cek student exists
	student, err := s.studentRepo.FindByID(req.StudentID)
	if err != nil || student == nil {
		return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
	}

	// create in mongo
	ctx := context.Background()
	objID, err := s.mongoRepo.CreateAchievement(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed create achievement (mongo)", err.Error()))
	}

	// create reference in postgres
	ref := &model.AchievementReference{
		ID:                 uuid.New().String(),
		StudentID:          req.StudentID,
		MongoAchievementID: objID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := s.postgresRepo.CreateReference(ref); err != nil {
		// rollback mongo insert (best-effort)
		_ = s.mongoRepo.DeleteAchievement(ctx, objID)
		return c.Status(500).JSON(model.ErrorResponse("failed create achievement reference", err.Error()))
	}

	return c.JSON(model.SuccessResponse(fiber.Map{
		"reference": ref,
		"achievementId": objID.Hex(),
	}))
}

// GET /achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	// For simplicity: support optional query ?studentId=...
	studentId := c.Query("studentId")
	var refs []model.AchievementReference
	var err error
	if studentId != "" {
		refs, err = s.postgresRepo.FindByStudentID(studentId)
	} else {
		refs, err = s.postgresRepo.FindAll()
	}
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed fetch references", err.Error()))
	}

	// optionally fetch details from mongo (small set)
	ctx := context.Background()
	results := []fiber.Map{}
	for _, r := range refs {
		var ach *model.Achievement
		objID, convErr := primitive.ObjectIDFromHex(r.MongoAchievementID)
		if convErr == nil {
			ach, _ = s.mongoRepo.FindByID(ctx, objID) // ignore error for now
		}
		results = append(results, fiber.Map{
			"reference":    r,
			"achievement":  ach,
		})
	}

	return c.JSON(model.SuccessResponse(results))
}

// GET /achievements/:id
func (s *AchievementService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("invalid mongo id", err.Error()))
	}
	ach, err := s.mongoRepo.FindByID(context.Background(), objID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed fetch achievement", err.Error()))
	}

	return c.JSON(model.SuccessResponse(fiber.Map{
		"reference":   ref,
		"achievement": ach,
	}))
}

// PUT /achievements/:id (only allowed when draft or by owner - auth not implemented here)
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	// update mongo doc
	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("invalid mongo id", err.Error()))
	}
	update := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"details":     req.Details,
		"tags":        req.Tags,
		"points":      req.Points,
	}
	if err := s.mongoRepo.UpdateAchievement(context.Background(), objID, update); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed update achievement", err.Error()))
	}

	return c.JSON(model.SuccessResponse("achievement updated"))
}

// DELETE /achievements/:id (soft-delete not implemented, do hard delete for draft)
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	// delete mongo document
	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err == nil {
		_ = s.mongoRepo.DeleteAchievement(context.Background(), objID)
	}
	note := "deleted"
	_ = s.postgresRepo.UpdateReferenceStatus(id, "rejected", nil, nil, nil, &note)

	return c.JSON(model.SuccessResponse("achievement deleted"))
}

// POST /achievements/:id/submit
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id := c.Params("id")
	_, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	now := time.Now()
	if err := s.postgresRepo.UpdateReferenceStatus(id, "submitted", &now, nil, nil, nil); err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed to submit", err.Error()))
	}

	// TODO: create notification to lecturer - out of scope here

	return c.JSON(model.SuccessResponse("achievement submitted"))
}

// POST /achievements/:id/verify
func (s *AchievementService) Verify(c *fiber.Ctx) error {
    id := c.Params("id")

    // Ambil claims dari middleware
    claims, ok := c.Locals("user").(*model.JWTClaims)
    if !ok {
        return c.Status(401).JSON(model.ErrorResponse("unauthorized", nil))
    }

    userID := claims.UserID

    _, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    now := time.Now()
    if err := s.postgresRepo.UpdateReferenceStatus(
        id,
        "verified",
        nil,
        &now,
        &userID,
        nil,
    ); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to verify", err.Error()))
    }

    return c.JSON(model.SuccessResponse("achievement verified"))
}


// POST /achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
    id := c.Params("id")

    var body struct {
        Note string `json:"note"`
    }

    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
    }

    if body.Note == "" {
        return c.Status(400).JSON(model.ErrorResponse("note is required", nil))
    }

    err := s.postgresRepo.UpdateReferenceStatus(
        id,
        "rejected",
        nil,        // verified_at (tidak diisi)
        nil,        // verified_by (tidak diisi)
        nil,        // submitted_at (tidak diubah)
		&body.Note,
    )

    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to reject", err.Error()))
    }

    return c.JSON(model.SuccessResponse("achievement rejected"))
}

func (s *AchievementService) List(c *fiber.Ctx) error {
    claims := c.Locals("user").(*model.JWTClaims)

    if claims.Role == "Mahasiswa" {
        // Ambil student_id dari token atau dari table
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

    // Admin atau Dosen Wali
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

    objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid mongo id", nil))
    }

    ach, err := s.mongoRepo.FindByID(context.Background(), objID)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("achievement not found in mongo", nil))
    }

    return c.JSON(model.SuccessResponse(fiber.Map{
        "reference":   ref,
        "achievement": ach,
    }))
}

func (s *AchievementService) History(c *fiber.Ctx) error {
    id := c.Params("id")

    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    // Minimal version: return created/submitted/verified timestamps
    return c.JSON(model.SuccessResponse(fiber.Map{
        "id":          ref.ID,
        "status":      ref.Status,
        "submittedAt": ref.SubmittedAt,
        "verifiedAt":  ref.VerifiedAt,
        "rejection":   ref.RejectionNote,
    }))
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
    return c.JSON(model.SuccessResponse("upload attachment not implemented yet"))
}
