package user

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"path/filepath"
	"sbitnev_back/internal/database/Store"
	"sbitnev_back/internal/database/models"
	"time"
)

type AdminHandler struct {
	Logger  *slog.Logger
	Storage *Store.Storage
}

func (h *AdminHandler) Menu(c *gin.Context) {
	c.HTML(200, "", nil)
}

func (h *AdminHandler) Management(c *gin.Context) {
	c.HTML(200, "", nil)
}

// Registers

func (h *AdminHandler) ScheduleRegister(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	filePath := filepath.Join("/schedule", file.Filename)
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if err := h.Storage.Schedule().ScheduleRegister(filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(202, gin.H{
		"status": "schedule registered",
	})
}

func (h *AdminHandler) UserRegister() gin.HandlerFunc {
	type Request struct {
		Login     string `json:"login"`
		Password  string `json:"password"`
		UserName  string `json:"userName"`
		Role      string `json:"role"`
		GroupName string `json:"groupName"`
	}
	return func(c *gin.Context) {
		const op = "handlers.UserRegister"
		rep := h.Storage.User()

		var request Request
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"error": err,
			})
			return
		}

		user := &models.User{
			Login:    request.Login,
			Password: request.Password,
			FullName: request.UserName,
			Role:     request.Role,
		}
		id, err := rep.CreateUser(user)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"error": err,
			})
			return
		}

		if err := rep.CreateUserLink(id, request.GroupName); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"status": "user registered",
		})
	}
}

func (h *AdminHandler) GroupRegister(c *gin.Context) {

}

func (h *AdminHandler) BackUp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "done",
	})
}

// Grades and Journal
//проверка на query параметры!

func (h *AdminHandler) Journal(c *gin.Context) {
	groups, err := h.Storage.Groups().GetAllGroups()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err,
		})
		return
	}

	c.HTML(200, "", groups)
}

func (h *AdminHandler) GradesRefactor() gin.HandlerFunc {
	type Request struct {
		UserName       string    `json:"userName"`
		DisciplineName string    `json:"disciplineName"`
		OldLevel       int       `json:"oldLevel"`
		OldDate        time.Time `json:"oldDate"`
		OldComment     string    `json:"oldComment,omitempty"`
		NewLevel       int       `json:"newLevel"`
		NewDate        time.Time `json:"newDate"`
		NewComment     string    `json:"newComment,omitempty"`
	}
	return func(c *gin.Context) {
		var req Request
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		user, err := h.Storage.User().GetUserByName(req.UserName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		discipline, err := h.Storage.Disciplines().GetDisciplineByName(req.DisciplineName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		oldGrade := &models.Grade{
			StudentID:    user.UserID,
			DisciplineID: discipline.DisciplineID,
			Level:        req.OldLevel,
			Date:         req.OldDate,
			Comment:      req.OldComment,
		}

		NewGrade := &models.Grade{
			StudentID:    user.UserID,
			DisciplineID: discipline.DisciplineID,
			Level:        req.NewLevel,
			Date:         req.NewDate,
			Comment:      req.NewComment,
		}

		if err := h.Storage.Journal().UpdateGrade(oldGrade, NewGrade); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"Accepted": "Grade is refactored",
		})
	}
}

// Schedules

func (h *AdminHandler) GetSchedule(c *gin.Context) {
	values := c.Request.URL.Query()
	groupName := values.Get("group")
	if groupName != "" {
		h.ScheduleWithQueryGroup(groupName)
		c.Abort()
		return
	}

	teacherName := values.Get("teacher")
	if teacherName != "" {
		h.ScheduleWithQueryTeacher(teacherName)
		c.Abort()
		return
	}

	h.GetPreSchedule()
}

func (h *AdminHandler) ScheduleWithQueryGroup(groupName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		schedule, err := h.Storage.ScheduleMethods.GetScheduleByGroupName(groupName)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
		}

		c.HTML(http.StatusOK, "", gin.H{
			"schedule": schedule,
		})
	}
}

func (h *AdminHandler) ScheduleWithQueryTeacher(teacherName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		schedule, err := h.Storage.ScheduleMethods.GetScheduleByTeacherName(teacherName)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		c.HTML(http.StatusOK, "", gin.H{
			"schedule": schedule,
		})
	}
}

func (h *AdminHandler) GetPreSchedule() gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, err := h.Storage.Groups().GetAllGroups()
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		teachers, err := h.Storage.User().GetAllTeachers()
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		c.HTML(200, "", gin.H{
			"teachers": teachers,
			"groups":   groups,
		})
	}
}