package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ReportRepository handles SQL aggregations (achievement_references)
type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// CountByStatus returns map[status]count, optionally filtered by student IDs.
// If studentIDs is empty or nil, it counts across all students.
func (r *ReportRepository) CountByStatus(studentIDs []string) (map[string]int, error) {
	result := map[string]int{
		"draft":     0,
		"submitted": 0,
		"verified":  0,
		"rejected":  0,
	}

	var rows *sql.Rows
	var err error

	if len(studentIDs) == 0 {
		rows, err = r.db.Query(`
			SELECT status, COUNT(*) as cnt
			FROM achievement_references
			GROUP BY status
		`)
	} else {
		// build IN clause
		args := make([]interface{}, len(studentIDs))
		placeholders := make([]string, len(studentIDs))
		for i, v := range studentIDs {
			args[i] = v
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(`
			SELECT status, COUNT(*) as cnt
			FROM achievement_references
			WHERE student_id IN (%s)
			GROUP BY status
		`, strings.Join(placeholders, ","))
		rows, err = r.db.Query(query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var cnt int
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, err
		}
		result[status] = cnt
	}

	return result, nil
}

// CountTotalPerStudent returns map[studentId]count for given studentIDs.
// If studentIDs empty -> returns for all students (may be large).
func (r *ReportRepository) CountTotalPerStudent(studentIDs []string, limit int) (map[string]int, error) {
	result := map[string]int{}

	var rows *sql.Rows
	var err error

	if len(studentIDs) == 0 {
		// get top by count
		query := `
			SELECT student_id, COUNT(*) as cnt
			FROM achievement_references
			GROUP BY student_id
			ORDER BY cnt DESC
			LIMIT $1
		`
		rows, err = r.db.Query(query, limit)
	} else {
		args := make([]interface{}, len(studentIDs)+1)
		placeholders := make([]string, len(studentIDs))
		for i, v := range studentIDs {
			args[i] = v
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		args[len(studentIDs)] = limit

		query := fmt.Sprintf(`
			SELECT student_id, COUNT(*) as cnt
			FROM achievement_references
			WHERE student_id IN (%s)
			GROUP BY student_id
			ORDER BY cnt DESC
			LIMIT $%d
		`, strings.Join(placeholders, ","), len(studentIDs)+1)

		rows, err = r.db.Query(query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sid string
		var cnt int
		if err := rows.Scan(&sid, &cnt); err != nil {
			return nil, err
		}
		result[sid] = cnt
	}

	return result, nil
}

// CountByPeriod returns counts grouped by month for given monthsBack (e.g., 6 months)
// optionally filtered by studentIDs. Returns map["YYYY-MM"] = count
func (r *ReportRepository) CountByPeriod(studentIDs []string, monthsBack int) (map[string]int, error) {
	res := map[string]int{}

	// compute from date
	from := time.Now().AddDate(0, -monthsBack+1, 0).Format("2006-01-02") // inclusive

	var rows *sql.Rows
	var err error

	if len(studentIDs) == 0 {
		rows, err = r.db.Query(`
			SELECT to_char(created_at, 'YYYY-MM') as month, COUNT(*) as cnt
			FROM achievement_references
			WHERE created_at >= $1
			GROUP BY month
			ORDER BY month
		`, from)
	} else {
		args := make([]interface{}, len(studentIDs)+1)
		placeholders := make([]string, len(studentIDs))
		for i, v := range studentIDs {
			args[i] = v
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		args[len(studentIDs)] = from

		query := fmt.Sprintf(`
			SELECT to_char(created_at, 'YYYY-MM') as month, COUNT(*) as cnt
			FROM achievement_references
			WHERE student_id IN (%s) AND created_at >= $%d
			GROUP BY month
			ORDER BY month
		`, strings.Join(placeholders, ","), len(studentIDs)+1)

		rows, err = r.db.Query(query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var month string
		var cnt int
		if err := rows.Scan(&month, &cnt); err != nil {
			return nil, err
		}
		res[month] = cnt
	}

	return res, nil
}
