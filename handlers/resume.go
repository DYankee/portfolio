package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DYankee/resume2/db"
	"github.com/DYankee/resume2/models"
	"github.com/DYankee/resume2/templates/pages"
	"github.com/go-pdf/fpdf"
	"github.com/labstack/echo/v4"
)

type ResumeHandler struct {
	DB *db.DB
}

func (h *ResumeHandler) HandleResumePage(c echo.Context) error {
	skills, _ := h.DB.GetAllSkills()
	experiences, _ := h.DB.GetAllExperiences()
	education, _ := h.DB.GetAllEducation()

	if c.Request().Header.Get("HX-Request") == "true" {
		return pages.ResumeContent(skills, experiences, education).Render(c.Request().Context(), c.Response())
	}

	return pages.ResumePage(skills, experiences, education).Render(c.Request().Context(), c.Response())
}

func (h *ResumeHandler) HandleResumePDF(c echo.Context) error {
	skills, err := h.DB.GetAllSkills()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load skills")
	}
	experiences, err := h.DB.GetAllExperiences()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load experiences")
	}
	education, err := h.DB.GetAllEducation()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load education")
	}

	pdf := buildResumePDF(skills, experiences, education)

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set(
		"Content-Disposition",
		`attachment; filename="Zachary_Geary_Resume.pdf"`,
	)

	return pdf.Output(c.Response().Writer)
}

func buildResumePDF(
	skills []models.Skill,
	experiences []models.Experience,
	education []models.Education,
) *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// ── Header ───────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 22)
	pdf.CellFormat(0, 10, "Zachary Geary", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 7, "Software Developer", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(
		0, 5,
		"zpgeary@gmail.com  |  linkedin.com/in/zacharygeary  |  github.com/DYankee",
		"", 1, "C", false, 0, "",
	)
	pdf.SetTextColor(0, 0, 0)

	drawLine(pdf)

	// ── About ────────────────────────────────────────────────
	sectionHeading(pdf, "About")
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(0, 5,
		"I'm a highly motivated college student looking to gain experience "+
			"in the professional world. I have experience developing small "+
			"websites and programs, working with other developers in an agile "+
			"like environment and interfacing with customers to ensure product "+
			"satisfaction. Currently open to work opportunities or internships.",
		"", "L", false,
	)
	pdf.Ln(2)
	pdf.MultiCell(0, 5,
		"I currently attend SUNY Polytechnic full time pursuing a CS degree. "+
			"In my free time I enjoy playing guitar, making games, coding tools "+
			"for myself, and going hiking.",
		"", "L", false,
	)

	// ── Education ────────────────────────────────────────────
	if len(education) > 0 {
		sectionHeading(pdf, "Education")
		for _, edu := range education {
			pdf.SetFont("Helvetica", "B", 11)
			degreeW := pdf.GetStringWidth(edu.Degree) + 2

			rightText := ""
			if edu.Gpa > 0 {
				rightText = fmt.Sprintf("GPA: %.2f", edu.Gpa)
			}
			if edu.In_progress {
				if rightText != "" {
					rightText += "  |  "
				}
				rightText += "In Progress"
			}

			pdf.CellFormat(degreeW, 6, edu.Degree, "", 0, "L", false, 0, "")

			if rightText != "" {
				pdf.SetFont("Helvetica", "", 9)
				pdf.CellFormat(0, 6, rightText, "", 0, "R", false, 0, "")
			}
			pdf.Ln(6)

			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(80, 80, 80)
			pdf.CellFormat(0, 5, edu.College, "", 1, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(2)
		}
	}

	// ── Experience ───────────────────────────────────────────
	if len(experiences) > 0 {
		sectionHeading(pdf, "Experience")
		for _, exp := range experiences {
			pdf.SetFont("Helvetica", "B", 11)
			titleW := pdf.GetStringWidth(exp.Title) + 2
			pdf.CellFormat(titleW, 6, exp.Title, "", 0, "L", false, 0, "")

			endDate := exp.EndDate
			if endDate == "" {
				endDate = "Present"
			} else if len(endDate) >= 7 {
				endDate = endDate[:7]
			}
			startDate := exp.StartDate
			if len(startDate) >= 7 {
				startDate = startDate[:7]
			}
			dateStr := startDate + "  -  " + endDate

			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(100, 100, 100)
			pdf.CellFormat(0, 6, dateStr, "", 0, "R", false, 0, "")
			pdf.Ln(6)

			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(80, 80, 80)
			pdf.CellFormat(0, 5, exp.Company, "", 1, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)

			if exp.Description != "" {
				pdf.SetFont("Helvetica", "", 9)
				pdf.MultiCell(0, 4.5, exp.Description, "", "L", false)
			}
			pdf.Ln(3)
		}
	}

	// ── Skills ───────────────────────────────────────────────
	if len(skills) > 0 {
		sectionHeading(pdf, "Skills")
		pdf.SetFont("Helvetica", "", 10)

		names := make([]string, len(skills))
		for i, s := range skills {
			names[i] = s.Name
		}
		pdf.MultiCell(0, 5, strings.Join(names, "  |  "), "", "L", false)
	}

	return pdf
}

func sectionHeading(pdf *fpdf.Fpdf, title string) {
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	drawLine(pdf)
}

func drawLine(pdf *fpdf.Fpdf) {
	w, _ := pdf.GetPageSize()
	ml, _, mr, _ := pdf.GetMargins()
	y := pdf.GetY()
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(ml, y, w-mr, y)
	pdf.Ln(3)
}
