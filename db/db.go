package db

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"github.com/DYankee/resume2/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Conn *sql.DB
}

func New(path string) *DB {
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMaxOpenConns(1)

	db := &DB{Conn: conn}
	db.migrate()
	return db
}

func (db *DB) migrate() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
    		token TEXT PRIMARY KEY,
    		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    		expires_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS skill_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category_id INTEGER NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			icon_url TEXT NOT NULL DEFAULT '',
			proficiency INTEGER NOT NULL DEFAULT 50,
			deleted INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (category_id) REFERENCES skill_categories(id)
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			display_order INTEGER DEFAULT 0,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			long_desc TEXT NOT NULL DEFAULT '',
			image_url TEXT NOT NULL DEFAULT '',
			repo_url TEXT NOT NULL DEFAULT '',
			live_url TEXT NOT NULL DEFAULT '',
			deleted INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS skill_uses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			skill_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,
			FOREIGN KEY (skill_id) REFERENCES skills(id),
			FOREIGN KEY (project_id) REFERENCES projects(id),
			UNIQUE(skill_id, project_id)
		)`,
		`CREATE TABLE IF NOT EXISTS experiences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			company TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			deleted INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS blog_posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			excerpt TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			published INTEGER NOT NULL DEFAULT 0,
			deleted INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS education (
   	 		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		degree TEXT NOT NULL,
    		college TEXT NOT NULL,
    		gpa REAL NOT NULL DEFAULT 0.0,
    		in_progress INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, q := range queries {
		if _, err := db.Conn.Exec(q); err != nil {
			log.Fatalf("migration failed: %v\n%s", err, q)
		}
	}
}

// ==================== Skill Categories ====================

func (db *DB) GetAllSkillCategories() ([]models.Skill_category, error) {
	rows, err := db.Conn.Query(`SELECT id, name FROM skill_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Skill_category
	for rows.Next() {
		var c models.Skill_category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (db *DB) CreateSkillCategory(name string) (int64, error) {
	res, err := db.Conn.Exec(`INSERT INTO skill_categories (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ==================== Skills ====================

func (db *DB) GetAllSkills() ([]models.Skill, error) {
	rows, err := db.Conn.Query(`
		SELECT s.id, s.name, sc.name, s.description, s.icon_url,
		       s.proficiency, s.deleted, s.created_at, s.updated_at,
		       COALESCE(s.deleted_at, '')
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		WHERE s.deleted = 0
		ORDER BY s.proficiency DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var deletedAt string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Category, &s.Description, &s.IconURL,
			&s.Proficiency, &s.Deleted, &s.CreatedAt, &s.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			s.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (db *DB) GetSkillByID(id int64) (*models.Skill, error) {
	var s models.Skill
	var deletedAt string
	err := db.Conn.QueryRow(`
		SELECT s.id, s.name, sc.name, s.description, s.icon_url,
		       s.proficiency, s.deleted, s.created_at, s.updated_at,
		       COALESCE(s.deleted_at, '')
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		WHERE s.id = ? AND s.deleted = 0`, id,
	).Scan(
		&s.ID, &s.Name, &s.Category, &s.Description, &s.IconURL,
		&s.Proficiency, &s.Deleted, &s.CreatedAt, &s.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}
	if deletedAt != "" {
		s.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
	}
	return &s, nil
}

func (db *DB) GetSkillsByCategoryID(categoryID int64) ([]models.Skill, error) {
	rows, err := db.Conn.Query(`
		SELECT s.id, s.name, sc.name, s.description, s.icon_url,
		       s.proficiency, s.deleted, s.created_at, s.updated_at,
		       COALESCE(s.deleted_at, '')
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		WHERE s.category_id = ? AND s.deleted = 0
		ORDER BY s.name`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var deletedAt string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Category, &s.Description, &s.IconURL,
			&s.Proficiency, &s.Deleted, &s.CreatedAt, &s.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			s.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (db *DB) GetRandomProjectForSkill(skillID int64) (*models.Project, error) {
	projects, err := db.GetProjectsForSkill(skillID)
	if err != nil || len(projects) == 0 {
		return nil, err
	}
	return &projects[rand.Intn(len(projects))], nil
}

func (db *DB) CreateSkill(name string, categoryID int64, description, iconURL string, proficiency int8) (int64, error) {
	res, err := db.Conn.Exec(`
		INSERT INTO skills (name, category_id, description, icon_url, proficiency)
		VALUES (?, ?, ?, ?, ?)`,
		name, categoryID, description, iconURL, proficiency,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateSkill(id int64, name string, categoryID int64, description, iconURL string, proficiency int8) error {
	_, err := db.Conn.Exec(`
		UPDATE skills
		SET name = ?, category_id = ?, description = ?, icon_url = ?,
		    proficiency = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted = 0`,
		name, categoryID, description, iconURL, proficiency, id,
	)
	return err
}

func (db *DB) SoftDeleteSkill(id int64) error {
	_, err := db.Conn.Exec(`
		UPDATE skills
		SET deleted = 1, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id,
	)
	return err
}

// ==================== Projects ====================

func (db *DB) GetAllProjects() ([]models.Project, error) {
	rows, err := db.Conn.Query(`
		SELECT id, title, description, long_desc, image_url, repo_url, live_url,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM projects
		WHERE deleted = 0
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var deletedAt string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.LongDesc, &p.ImageURL,
			&p.RepoURL, &p.LiveURL, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (db *DB) GetProjectByID(id int64) (*models.Project, error) {
	var p models.Project
	var deletedAt string
	err := db.Conn.QueryRow(`
		SELECT id, title, description, long_desc, image_url, repo_url, live_url,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM projects
		WHERE id = ? AND deleted = 0`, id,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.LongDesc, &p.ImageURL,
		&p.RepoURL, &p.LiveURL, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}
	if deletedAt != "" {
		p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
	}
	return &p, nil
}

func (db *DB) CreateProject(title, description, longDesc, imageURL, repoURL, liveURL string) (int64, error) {
	res, err := db.Conn.Exec(`
		INSERT INTO projects (display_order, title, description, long_desc, image_url, repo_url, live_url)
		VALUES (MAX(display_order + 1), ?, ?, ?, ?, ?, ?)`,
		title, description, longDesc, imageURL, repoURL, liveURL,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateProject(id int64, title, description, longDesc, imageURL, repoURL, liveURL string) error {
	_, err := db.Conn.Exec(`
		UPDATE projects
		SET title = ?, description = ?, long_desc = ?, image_url = ?,
		    repo_url = ?, live_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted = 0`,
		title, description, longDesc, imageURL, repoURL, liveURL, id,
	)
	return err
}

func (db *DB) SoftDeleteProject(id int64) error {
	_, err := db.Conn.Exec(`
		UPDATE projects
		SET deleted = 1, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id,
	)
	return err
}

// ==================== Skill Uses (Project <-> Skill) ====================

func (db *DB) AddSkillToProject(skillID, projectID int64) error {
	_, err := db.Conn.Exec(`
		INSERT OR IGNORE INTO skill_uses (skill_id, project_id)
		VALUES (?, ?)`, skillID, projectID,
	)
	return err
}

func (db *DB) RemoveSkillFromProject(skillID, projectID int64) error {
	_, err := db.Conn.Exec(`
		DELETE FROM skill_uses WHERE skill_id = ? AND project_id = ?`,
		skillID, projectID,
	)
	return err
}

func (db *DB) GetSkillsForProject(projectID int64) ([]models.Skill, error) {
	rows, err := db.Conn.Query(`
		SELECT s.id, s.name, sc.name, s.description, s.icon_url,
		       s.proficiency, s.deleted, s.created_at, s.updated_at,
		       COALESCE(s.deleted_at, '')
		FROM skills s
		JOIN skill_categories sc ON s.category_id = sc.id
		JOIN skill_uses su ON su.skill_id = s.id
		WHERE su.project_id = ? AND s.deleted = 0
		ORDER BY s.name`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var deletedAt string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Category, &s.Description, &s.IconURL,
			&s.Proficiency, &s.Deleted, &s.CreatedAt, &s.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			s.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (db *DB) GetProjectsForSkill(skillID int64) ([]models.Project, error) {
	rows, err := db.Conn.Query(`
		SELECT p.id, p.title, p.description, p.long_desc, p.image_url,
		       p.repo_url, p.live_url, p.deleted, p.created_at, p.updated_at,
		       COALESCE(p.deleted_at, '')
		FROM projects p
		JOIN skill_uses su ON su.project_id = p.id
		WHERE su.skill_id = ? AND p.deleted = 0
		ORDER BY p.created_at DESC`, skillID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var deletedAt string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.LongDesc, &p.ImageURL,
			&p.RepoURL, &p.LiveURL, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// ==================== Experiences ====================

func (db *DB) GetAllExperiences() ([]models.Experience, error) {
	rows, err := db.Conn.Query(`
		SELECT id, title, company, start_date, end_date, description,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM experiences
		WHERE deleted = 0
		ORDER BY start_date DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []models.Experience
	for rows.Next() {
		var e models.Experience
		var deletedAt string
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Company, &e.StartDate, &e.EndDate,
			&e.Description, &e.Deleted, &e.CreatedAt, &e.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			e.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		exps = append(exps, e)
	}
	return exps, nil
}

func (db *DB) GetExperienceByID(experienceID int64) (*models.Experience, error) {
	var e models.Experience
	var deletedAt string
	err := db.Conn.QueryRow(`
		SELECT id, title, company, start_date, end_date, description,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM experiences
		WHERE id = ? AND deleted = 0`, experienceID,
	).Scan(
		&e.ID, &e.Title, &e.Company, &e.StartDate, &e.EndDate,
		&e.Description, &e.Deleted, &e.CreatedAt, &e.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (db *DB) CreateExperience(title, company, startDate, endDate, description string) (int64, error) {
	res, err := db.Conn.Exec(`
		INSERT INTO experiences (title, company, start_date, end_date, description)
		VALUES (?, ?, ?, ?, ?)`,
		title, company, startDate, endDate, description,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateExperience(id int64, title, company, startDate, endDate, description string) error {
	_, err := db.Conn.Exec(`
		UPDATE experiences
		SET title = ?, company = ?, start_date = ?, end_date = ?,
		    description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted = 0`,
		title, company, startDate, endDate, description, id,
	)
	return err
}

func (db *DB) SoftDeleteExperience(id int64) error {
	_, err := db.Conn.Exec(`
		UPDATE experiences
		SET deleted = 1, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id,
	)
	return err
}

// ==================== Education ====================

func (db *DB) GetAllEducation() ([]models.Education, error) {
	rows, err := db.Conn.Query(`
		SELECT id, degree, college, gpa, in_progress
		FROM education
		ORDER BY in_progress DESC, degree`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edus []models.Education
	for rows.Next() {
		var e models.Education
		if err := rows.Scan(&e.ID, &e.Degree, &e.College, &e.Gpa, &e.In_progress); err != nil {
			return nil, err
		}
		edus = append(edus, e)
	}
	return edus, nil
}

func (db *DB) CreateEducation(degree, college string, gpa float64, inProgress bool) error {
	ip := 0
	if inProgress {
		ip = 1
	}
	_, err := db.Conn.Exec(`
		INSERT INTO education (degree, college, gpa, in_progress)
		VALUES (?, ?, ?, ?)`,
		degree, college, gpa, ip,
	)
	return err
}

func (db *DB) GetEducationByID(id int64) (*models.Education, error) {
	var e models.Education
	err := db.Conn.QueryRow(`
		SELECT id, degree, college, gpa, in_progress
		FROM education
		WHERE id = ?`, id,
	).Scan(&e.ID, &e.Degree, &e.College, &e.Gpa, &e.In_progress)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (db *DB) UpdateEducation(
	id int64, degree, college string, gpa float64, inProgress bool,
) error {
	ip := 0
	if inProgress {
		ip = 1
	}
	_, err := db.Conn.Exec(`
		UPDATE education
		SET degree = ?, college = ?, gpa = ?, in_progress = ?
		WHERE id = ?`,
		degree, college, gpa, ip, id,
	)
	return err
}

func (db *DB) SoftDeleteEducation(id int64) error {
	_, err := db.Conn.Exec(`
		UPDATE education
		SET deleted = 1, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id,
	)
	return err
}

// ==================== Blog Posts ====================

func (db *DB) GetPublishedPosts() ([]models.BlogPost, error) {
	rows, err := db.Conn.Query(`
		SELECT id, title, slug, excerpt, content, tags, published,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM blog_posts
		WHERE published = 1 AND deleted = 0
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.BlogPost
	for rows.Next() {
		var p models.BlogPost
		var deletedAt string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.Tags,
			&p.Published, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (db *DB) GetAllPosts() ([]models.BlogPost, error) {
	rows, err := db.Conn.Query(`
		SELECT id, title, slug, excerpt, content, tags, published,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM blog_posts
		WHERE deleted = 0
		ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.BlogPost
	for rows.Next() {
		var p models.BlogPost
		var deletedAt string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.Tags,
			&p.Published, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, err
		}
		if deletedAt != "" {
			p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (db *DB) GetPostBySlug(slug string) (*models.BlogPost, error) {
	var p models.BlogPost
	var deletedAt string
	err := db.Conn.QueryRow(`
		SELECT id, title, slug, excerpt, content, tags, published,
		       deleted, created_at, updated_at, COALESCE(deleted_at, '')
		FROM blog_posts
		WHERE slug = ? AND published = 1 AND deleted = 0`, slug,
	).Scan(
		&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.Tags,
		&p.Published, &p.Deleted, &p.CreatedAt, &p.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}
	if deletedAt != "" {
		p.DeletedAt, _ = time.Parse("2006-01-02 15:04:05", deletedAt)
	}
	return &p, nil
}

func (db *DB) CreateBlogPost(title, slug, excerpt, content, tags string, published bool) (int64, error) {
	pub := 0
	if published {
		pub = 1
	}
	res, err := db.Conn.Exec(`
		INSERT INTO blog_posts (title, slug, excerpt, content, tags, published)
		VALUES (?, ?, ?, ?, ?, ?)`,
		title, slug, excerpt, content, tags, pub,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) UpdateBlogPost(id int64, title, slug, excerpt, content, tags string, published bool) error {
	pub := 0
	if published {
		pub = 1
	}
	_, err := db.Conn.Exec(`
		UPDATE blog_posts
		SET title = ?, slug = ?, excerpt = ?, content = ?, tags = ?,
		    published = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted = 0`,
		title, slug, excerpt, content, tags, pub, id,
	)
	return err
}

func (db *DB) SoftDeleteBlogPost(id int64) error {
	_, err := db.Conn.Exec(`
		UPDATE blog_posts
		SET deleted = 1, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, id,
	)
	return err
}
