package models

import "time"

type Skill struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	Proficiency int8      `json:"proficiency"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type Skill_category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID           int64     `json:"id"`
	DisplayOrder int32     `json:"display_order"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	LongDesc     string    `json:"long_desc"`
	ImageURL     string    `json:"image_url"`
	RepoURL      string    `json:"repo_url"`
	LiveURL      string    `json:"live_url"`
	Deleted      bool      `json:"deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`
}

type SkillUse struct {
	ID         int64 `json:"id"`
	Skill_ID   int64 `json:"skill_id"`
	Project_ID int64 `json:"project_id"`
}

type Experience struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"` // empty = "Present"
	Description string    `json:"description"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type Education struct {
	ID          int64   `json:"id"`
	Degree      string  `json:"degree"`
	College     string  `json:"college"`
	Gpa         float64 `json:"gpa"`
	In_progress bool    `json:"in_progress"`
}

type BlogPost struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Excerpt   string    `json:"excerpt"`
	Content   string    `json:"content"`
	Tags      string    `json:"tags"`
	Published bool      `json:"published"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
