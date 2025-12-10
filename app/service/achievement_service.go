package service

import (
	"context"
	"time"
    "fmt"

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

	// ✅ Ambil JWT claims dari middleware
	claims, ok := c.Locals("user").(*model.JWTClaims)
	if !ok {
		return c.Status(401).JSON(model.ErrorResponse("unauthorized", nil))
	}

	// ✅ Ambil student berdasarkan user login
	student, err := s.studentRepo.FindByUserID(claims.UserID)
	if err != nil || student == nil {
		return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
	}

	// ✅ Paksa studentId berasal dari login user
	req.StudentID = student.ID

	// ✅ Create ke Mongo terlebih dulu
	ctx := context.Background()
	objID, err := s.mongoRepo.CreateAchievement(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse("failed create achievement (mongo)", err.Error()))
	}

	// ✅ Simpan reference ke Postgre
	ref := &model.AchievementReference{
		ID:                 uuid.New().String(),
		StudentID:          student.ID,
		MongoAchievementID: objID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.postgresRepo.CreateReference(ref); err != nil {
		// rollback mongo insert
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

	// Ambil reference
	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	// ⛔ Ownership: mahasiswa hanya boleh update miliknya sendiri
	if err := s.checkOwnership(c, ref.StudentID); err != nil {
		return err
	}

	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
	}

	// Update Mongo
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

	// Ambil reference
	ref, err := s.postgresRepo.FindReferenceByID(id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
	}

	// ⛔ Ownership: mahasiswa hanya boleh delete miliknya sendiri
	if err := s.checkOwnership(c, ref.StudentID); err != nil {
		return err
	}

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

    // Ambil reference
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    // ⛔ Mahasiswa hanya boleh submit achievement miliknya sendiri
    if err := s.checkOwnership(c, ref.StudentID); err != nil {
        return err
    }

    // Update status menjadi submitted
    now := time.Now()
    if err := s.postgresRepo.UpdateReferenceStatus(
        id,
        "submitted",
        &now, // submittedAt diisi sekarang
        nil,  // verifiedAt
        nil,  // verifiedBy
        nil,  // rejectionNote
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

    return c.JSON(model.SuccessResponse("achievement verified"))
}



// POST /achievements/:id/reject
func (s *AchievementService) Reject(c *fiber.Ctx) error {
    id := c.Params("id")

    body := struct{ Note string `json:"note"` }{}
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid body", err.Error()))
    }

    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    // ⛔ Cek apakah mahasiswa ini dibimbing dosen tersebut
    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }

    if err := s.postgresRepo.UpdateReferenceStatus(
        id, "rejected", nil, nil, nil, &body.Note,
    ); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to reject", err.Error()))
    }

    return c.JSON(model.SuccessResponse("achievement rejected"))
}

func (s *AchievementService) List(c *fiber.Ctx) error {
    claims := c.Locals("user").(*model.JWTClaims)

    // 🧑‍🎓 = Mahasiswa → list achievement miliknya sendiri
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

    // 🎓 = Dosen Wali → list semua mahasiswa bimbingannya saja
    if claims.Role == "Dosen Wali" {
        // ambil mahasiswa yang dibimbing dosen ini
        students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
        if err != nil {
            return c.Status(500).JSON(model.ErrorResponse("failed to fetch advisory students", nil))
        }

        var allRefs []model.AchievementReference

        // Ambil achievement untuk masing-masing mahasiswa
        for _, stu := range students {
            refs, _ := s.postgresRepo.FindByStudentID(stu.ID)
            allRefs = append(allRefs, refs...)
        }

        return c.JSON(model.SuccessResponse(allRefs))
    }

    // 🛠️ = Admin → ambil semua achievement
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

    // ⛔ Advisor check untuk dosen wali
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

    // ⛔ Dosen wali hanya boleh lihat history mahasiswa bimbingan
    if err := s.checkAdvisor(c, ref.StudentID); err != nil {
        return err
    }

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

    // 1. Ambil reference dari PostgreSQL
    ref, err := s.postgresRepo.FindReferenceByID(id)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("reference not found", nil))
    }

    // 2. Ambil file dari request
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("file is required", err.Error()))
    }

    // 3. Buat nama file unik
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    savePath := fmt.Sprintf("./uploads/%s", filename)

    // 4. Simpan file ke folder lokal
    if err := c.SaveFile(file, savePath); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to save file", err.Error()))
    }

    // 5. Buat meta attachment
    attachment := model.Attachment{
        FileName:   file.Filename,
        FileURL:    "/uploads/" + filename,
        FileType:   file.Header.Get("Content-Type"),
        UploadedAt: time.Now(),
    }

    // 6. Konversi Mongo ID
    objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
    if err != nil {
        return c.Status(400).JSON(model.ErrorResponse("invalid mongo id", nil))
    }

    // 7. Simpan attachment ke MongoDB
    if err := s.mongoRepo.AddAttachment(context.Background(), objID, attachment); err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to update mongo", err.Error()))
    }

    // 8. Response sukses
    return c.JSON(model.SuccessResponse(fiber.Map{
        "message": "file uploaded successfully",
        "file":    attachment,
    }))
}


func (s *AchievementService) checkOwnership(c *fiber.Ctx, refStudentID string) error {
    claims := c.Locals("user").(*model.JWTClaims)

    // Mahasiswa saja yang butuh ownership check
    if claims.Role != "Mahasiswa" {
        return nil // dosen/admin lolos (ini khusus untuk step 1)
    }

    // Cari student milik user login
    student, err := s.studentRepo.FindByUserID(claims.UserID)
    if err != nil {
        return c.Status(404).JSON(model.ErrorResponse("student not found", nil))
    }

    // Cocokkan StudentID pemilik achievement dengan mahasiswa login
    if student.ID != refStudentID {
        return c.Status(403).JSON(model.ErrorResponse("forbidden: not owner of the achievement", nil))
    }

    // Lolos → lanjut
    return nil
}

func (s *AchievementService) checkAdvisor(c *fiber.Ctx, refStudentID string) error {
    claims := c.Locals("user").(*model.JWTClaims)

    // Beri access penuh untuk Admin
    if claims.Role == "Admin" {
        return nil
    }

    // Hanya berlaku untuk Dosen Wali
    if claims.Role != "Dosen Wali" {
        return nil
    }

    // Cari semua mahasiswa bimbingan dosen wali
    students, err := s.studentRepo.FindByAdvisorID(claims.UserID)
    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to fetch advisory students", nil))
    }

    // Cek apakah studentID pemilik achievement ada di dalam list bimbingan
    for _, s := range students {
        if s.ID == refStudentID {
            return nil // berarti pemilik achievement adalah mahasiswa bimbingannya
        }
    }

    return c.Status(403).JSON(model.ErrorResponse("forbidden: student is not under your supervision", nil))
}

func (s *AchievementService) GetByStudentID(studentID string) ([]fiber.Map, error) {

    // Ambil semua reference dari Postgre
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

        // Ambil data achievement dari Mongo
        ach, err := s.mongoRepo.FindByID(ctx, objID)
        if err != nil {
            continue
        }

        // Gabungkan
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

    // Ambil semua reference milik student
    refs, err := s.postgresRepo.FindByStudentID(studentID)
    if err != nil {
        return c.Status(500).JSON(model.ErrorResponse("failed to fetch references", err.Error()))
    }

    ctx := context.Background()
    results := []fiber.Map{}

    // Loop setiap reference → ambil achievement dari Mongo
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
