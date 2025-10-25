package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	dbmodels "github.com/mohammedrefaat/hamber/DB_models"
	"github.com/mohammedrefaat/hamber/utils"
)

// ========== CALENDAR EVENT CONTROLLERS ==========

type CreateEventRequest struct {
	Title          string               `json:"title" binding:"required"`
	Description    string               `json:"description"`
	Location       string               `json:"location"`
	StartTime      time.Time            `json:"start_time" binding:"required"`
	EndTime        time.Time            `json:"end_time" binding:"required"`
	AllDay         bool                 `json:"all_day"`
	EventType      string               `json:"event_type"`
	Color          string               `json:"color"`
	IsPublic       bool                 `json:"is_public"`
	Recurring      bool                 `json:"recurring"`
	RecurrenceRule string               `json:"recurrence_rule"`
	RemindBefore   int                  `json:"remind_before"`
	Attendees      []EventAttendeeInput `json:"attendees"`
}

type EventAttendeeInput struct {
	UserID *uint  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func CreateCalendarEvent(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}
	if req.IsPublic && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create public events"})
		return
	}

	event := &dbmodels.CalendarEvent{
		UserID:         claims.UserID,
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		AllDay:         req.AllDay,
		EventType:      req.EventType,
		Color:          req.Color,
		IsPublic:       req.IsPublic,
		Recurring:      req.Recurring,
		RecurrenceRule: req.RecurrenceRule,
		RemindBefore:   req.RemindBefore,
		Status:         dbmodels.EventStatus_SCHEDULED,
	}

	if err := globalStore.StStore.CreateCalendarEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}
	//  Schedule reminder notification
	if event.RemindBefore > 0 && globalStore.NotifService != nil {
		reminderTime := event.StartTime.Add(-time.Duration(event.RemindBefore) * time.Minute)
		if reminderTime.After(time.Now()) {
			// You can implement a scheduler here to send notification at specific time
			// For now, we'll just log it
			log.Printf("Reminder scheduled for event %d at %s", event.ID, reminderTime)
		}
	}
	// Add attendees if provided
	for _, attendee := range req.Attendees {
		eventAttendee := &dbmodels.EventAttendee{
			EventID:        event.ID,
			UserID:         attendee.UserID,
			Email:          attendee.Email,
			Name:           attendee.Name,
			ResponseStatus: "pending",
		}
		globalStore.StStore.AddEventAttendee(eventAttendee)
	}

	c.JSON(http.StatusCreated, gin.H{
		"event":   event,
		"message": "Event created successfully",
	})
}

func GetUserEvents(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Parse month and year from query params
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
	includePublic := c.DefaultQuery("include_public", "true") == "true"

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	// Calculate start and end of month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	events, err := globalStore.StStore.GetUserCalendarEvents(claims.UserID, startDate, endDate, includePublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"period": gin.H{
			"year":  year,
			"month": month,
			"start": startDate,
			"end":   endDate,
		},
	})
}

// GetPublicEvents godoc
// @Summary      Get public calendar events
// @Description  Retrieves all public events for a specific month/year
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        year           query   int     false  "Year (default: current year)"
// @Param        month          query   int     false  "Month 1-12 (default: current month)"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/public-events [get]
// @Security     Bearer

// GetPublicEvents returns all public events for a given month
func GetPublicEvents(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	events, err := globalStore.StStore.GetPublicEvents(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"period": gin.H{
			"year":  year,
			"month": month,
		},
	})
}

// GetCalendarEvent godoc
// @Summary      Get a specific calendar event
// @Description  Retrieves details of a single calendar event by ID
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id             path      int     true  "Event ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/events/{id} [get]
// @Security     Bearer

// GetCalendarEvent returns a specific event
func GetCalendarEvent(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := globalStore.StStore.GetCalendarEvent(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Get attendees
	attendees, _ := globalStore.StStore.GetEventAttendees(event.ID)

	c.JSON(http.StatusOK, gin.H{
		"event":     event,
		"attendees": attendees,
	})
}

// UpdateCalendarEvent godoc
// @Summary      Update an existing calendar event
// @Description  Updates details of a specific calendar event by ID
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string              true  "Bearer token"
// @Param        id             path      int                 true  "Event ID"
// @Param        event          body      CreateEventRequest  true  "Updated event details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/events/{id} [put]
// @Security     Bearer

// UpdateCalendarEvent updates an existing event
func UpdateCalendarEvent(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := globalStore.StStore.GetCalendarEvent(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Verify ownership
	if event.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	event.Title = req.Title
	event.Description = req.Description
	event.Location = req.Location
	event.StartTime = req.StartTime
	event.EndTime = req.EndTime
	event.AllDay = req.AllDay
	event.EventType = req.EventType
	event.Color = req.Color
	event.IsPublic = req.IsPublic
	event.Recurring = req.Recurring
	event.RecurrenceRule = req.RecurrenceRule
	event.RemindBefore = req.RemindBefore

	if err := globalStore.StStore.UpdateCalendarEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event":   event,
		"message": "Event updated successfully",
	})
}

// DeleteCalendarEvent godoc
// @Summary      Delete a calendar event
// @Description  Deletes a specific calendar event by ID
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id             path      int     true  "Event ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/events/{id} [delete]
// @Security     Bearer

// DeleteCalendarEvent deletes an event
func DeleteCalendarEvent(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := globalStore.StStore.DeleteCalendarEvent(uint(id), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event deleted successfully",
	})
}

// UpdateEventStatus godoc
// @Summary      Update the status of a calendar event
// @Description  Updates the status (e.g., SCHEDULED, ONGOING) of a specific calendar event by ID
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id             path      int     true  "Event ID"
// @Param        status         body      object  true  "Event status update"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/events/{id}/status [put]
// @Security     Bearer

// UpdateEventStatus updates the status of an event
func UpdateEventStatus(c *gin.Context) {
	claims, err := utils.GetclamsFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=SCHEDULED ONGOING COMPLETED CANCELLED POSTPONED"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := globalStore.StStore.GetCalendarEvent(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.UserID != claims.UserID && claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	statusMap := map[string]dbmodels.EventStatus{
		"SCHEDULED": dbmodels.EventStatus_SCHEDULED,
		"ONGOING":   dbmodels.EventStatus_ONGOING,
		"COMPLETED": dbmodels.EventStatus_COMPLETED,
		"CANCELLED": dbmodels.EventStatus_CANCELLED,
		"POSTPONED": dbmodels.EventStatus_POSTPONED,
	}

	event.Status = statusMap[req.Status]

	if err := globalStore.StStore.UpdateCalendarEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event":   event,
		"message": "Event status updated successfully",
	})
}

// RespondToInvitation godoc
// @Summary      Respond to an event invitation
// @Description  Allows an attendee to accept, decline, or tentatively accept an event invitation
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        attendee_id    path      int     true  "Attendee ID"
// @Param        response       body      object  true  "Attendee response (accepted, declined, tentative)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /calendar/attendees/{attendee_id}/response [put]
// @Security     Bearer

// RespondToInvitation allows attendees to respond to event invitations
func RespondToInvitation(c *gin.Context) {

	attendeeID, err := strconv.ParseUint(c.Param("attendee_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendee ID"})
		return
	}

	var req struct {
		Response string `json:"response" binding:"required,oneof=accepted declined tentative"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := globalStore.StStore.UpdateAttendeeResponse(uint(attendeeID), req.Response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Response updated successfully",
	})
}
