package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "social-network/models"
	"social-network/pkg/db/sqlite"
	"social-network/util"
)

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the current user from the session
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
			})
			return
	}

	// Get user ID
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	var group struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON format",
		})
		return
	}

	if strings.TrimSpace(group.Description) == "" || strings.TrimSpace(group.Title) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Please provide all required fields",
		})
		return
	}

	// Insert the group
	result, err := sqlite.DB.Exec(
		"INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)",
		creatorID, group.Title, group.Description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create group",
		})
		log.Printf("Error creating group: %v", err)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get group ID",
		})
		return
	}

	// Add creator as a member
	_, err = sqlite.DB.Exec(
		"INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)",
		groupID, creatorID, "creator")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to add creator as member",
		})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      groupID,
		"message": "Group created successfully",
	})
}

func CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	var post m.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	// post will always be public for the group members
	post.Privay = 1

	// check if the passed privacy is within the allowed range
	if post.Privay != 1 && post.Privay != 2 && post.Privay != 3 {
		http.Error(w, "invalid privacy type", http.StatusBadRequest)
		return
	}

	if _, err := sqlite.DB.Exec("INSERT INTO posts (title, content, media, privacy, author, group_id) VALUES (?, ?, ?, ?, ?, ?)", post.Title, post.Content, post.Media, post.Privay, post.Author, groupID); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("create post: %v", err)
		return
	}

	w.Write([]byte("Post created successfully"))
}

func GetGroupPost(w http.ResponseWriter, r *http.Request) {
	var groupPosts []m.Post
	groupIDString := r.PathValue("id")

	// convert the string into a number
	groupID, err := strconv.Atoi(groupIDString)
	if err != nil {
		http.Error(w, "Invalid number", http.StatusBadRequest)
		return
	}

	// the value of the group id can't be less then 1
	if groupID < 1 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rows, err := sqlite.DB.Query("SELECT id, title, content, media, privacy, author, created_at, group_id FROM posts WHERE group_id = ?", groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group does not exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post m.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Media, &post.Privay, &post.Author, &post.CreatedAt, &post.GroupID); err != nil {
			http.Error(w, "Error getting post", http.StatusInternalServerError)
			log.Printf("Error scanning: %v", err)
			return
		}

		groupPosts = append(groupPosts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&groupPosts); err != nil {
		http.Error(w, "Error sending json", http.StatusInternalServerError)
	}
}

func VeiwGorups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	var userID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user",
		})
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT 
			g.id, g.title, g.description, g.creator_id, g.created_at,
			u.username as creator_username,
			EXISTS(
				SELECT 1 FROM group_members gm 
				WHERE gm.group_id = g.id AND gm.user_id = ? AND gm.status = 'member'
			) as is_member,
			EXISTS(
				SELECT 1 FROM group_members gm 
				WHERE gm.group_id = g.id AND gm.user_id = ? AND gm.status = 'pending'
			) as has_pending_request
		FROM groups g
		JOIN users u ON g.creator_id = u.id
	`, userID, userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch groups",
		})
		return
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var group struct {
			ID              uint      `json:"id"`
			Title           string    `json:"title"`
			Description     string    `json:"description"`
			CreatorID       uint      `json:"creator_id"`
			CreatorUsername string    `json:"creator_username"`
			CreatedAt       time.Time `json:"created_at"`
			IsMember        bool      `json:"is_member"`
			HasPending      bool      `json:"has_pending_request"`
		}

		err := rows.Scan(
			&group.ID, &group.Title, &group.Description,
			&group.CreatorID, &group.CreatedAt, &group.CreatorUsername,
			&group.IsMember, &group.HasPending)

		if err != nil {
			continue
		}

		groups = append(groups, map[string]interface{}{
			"id":              group.ID,
			"title":          group.Title,
			"description":    group.Description,
			"creator_id":     group.CreatorID,
			"creator_username": group.CreatorUsername,
			"created_at":     group.CreatedAt,
			"isMember":       group.IsMember,
			"hasPendingRequest": group.HasPending,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func GroupInvitation(w http.ResponseWriter, r *http.Request) {
	var inviteRequest m.GroupInvaitation
	var group m.Group
	var groupMembers m.GroupMemebers

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// check if the user exists
	err := sqlite.DB.QueryRow("SELECT * FROM groups WHERE id = ?", inviteRequest.GroupID).Scan(&group.ID, &group.Title, &group.Description, &group.CreatorID, &group.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Group does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	// check if the user is already a member of the group
	err = sqlite.DB.QueryRow("SELECT * FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.InviterID).Scan(&groupMembers.ID, &groupMembers.GroupID, &groupMembers.UserID, &groupMembers.Status, &groupMembers.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User is not a member of the group", http.StatusBadRequest)
			return
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	// insert the new member to the group
	if _, err := sqlite.DB.Exec("INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)", inviteRequest.GroupID, inviteRequest.ReciverID, "pending"); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation sent successfully"})
}

func GroupAccept(w http.ResponseWriter, r *http.Request) {
	var inviteRequest m.GroupInvaitation

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate input
	if inviteRequest.GroupID <= 0 || inviteRequest.ReciverID <= 0 {
		http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
		return
	}

	// Check if the invitation exists and is pending
	var status string
	err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invitation not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error checking invitation status", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
		}
		return
	}

	if status != "pending" {
		http.Error(w, "Invitation is not pending", http.StatusBadRequest)
		return
	}

	// Update the status to "accepted"
	result, err := sqlite.DB.Exec("UPDATE group_members SET status = 'member' WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID)
	if err != nil {
		http.Error(w, "Error updating invitation status", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No invitation updated", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation accepted successfully"})
}

func GroupReject(w http.ResponseWriter, r *http.Request) {
	var inviteRequest m.GroupInvaitation

	if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Validate input
	if inviteRequest.GroupID <= 0 || inviteRequest.ReciverID <= 0 {
		http.Error(w, "Invalid group ID or receiver ID", http.StatusBadRequest)
		return
	}

	// Check if the invitation exists
	var status string
	err := sqlite.DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invitation not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Error checking invitation status", http.StatusInternalServerError)
			log.Printf("Error: %v", err)
			return
		}
	}

	if status != "pending" {
		http.Error(w, "Invitation is not pending", http.StatusBadRequest)
		return
	}

	// Delete the invitation
	result, err := sqlite.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", inviteRequest.GroupID, inviteRequest.ReciverID)
	if err != nil {
		http.Error(w, "Error deleting invitation", http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No invitation deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group invitation rejected successfully"})
}

func GroupLeave(w http.ResponseWriter, r *http.Request) {
	// neded Get the group
	var Leave m.GroupLeave

	if err := json.NewDecoder(r.Body).Decode(&Leave); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var exists bool
	err := sqlite.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)", Leave.GroupID, Leave.UserID).
		Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}

	// leave logic
	result, err := sqlite.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", Leave.GroupID, Leave.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User successfully removed from the group"})

}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid group ID",
		})
		return
	}

	// Get group details with creator username
	var group struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		CreatorID   int       `json:"creator_id"`
		CreatedAt   time.Time `json:"created_at"`
		CreatorUsername string `json:"creator_username"`
	}

	err = sqlite.DB.QueryRow(`
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at, u.username
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		WHERE g.id = ?`, groupID).Scan(
		&group.ID, &group.Title, &group.Description, 
		&group.CreatorID, &group.CreatedAt, &group.CreatorUsername)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Group not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch group",
		})
		log.Printf("Error fetching group: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid group ID",
		})
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT u.id, u.username, u.first_name, u.last_name, u.avatar, gm.status
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY gm.created_at`, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch members",
		})
		log.Printf("Error fetching members: %v", err)
		return
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var member struct {
			ID        int    `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Avatar    string `json:"avatar"`
			Status    string `json:"status"`
		}
		
		if err := rows.Scan(&member.ID, &member.Username, &member.FirstName, 
			&member.LastName, &member.Avatar, &member.Status); err != nil {
			log.Printf("Error scanning member: %v", err)
			continue
		}
		
		members = append(members, map[string]interface{}{
			"id":        member.ID,
			"username":  member.Username,
			"firstName": member.FirstName,
			"lastName":  member.LastName,
			"avatar":    member.Avatar,
			"status":    member.Status,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

func CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user
	username, err := util.GetUsernameFromSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Get user ID
	var creatorID int
	err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&creatorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get user information",
		})
		return
	}

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid group ID",
		})
		return
	}

	var event struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventDate   time.Time `json:"eventDate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request data",
		})
		return
	}

	// Insert the event
	result, err := sqlite.DB.Exec(`
		INSERT INTO group_events (group_id, creator_id, title, description, event_date)
		VALUES (?, ?, ?, ?, ?)`,
		groupID, creatorID, event.Title, event.Description, event.EventDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create event",
		})
		log.Printf("Error creating event: %v", err)
		return
	}

	eventID, _ := result.LastInsertId()

	// Return the created event
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          eventID,
		"groupId":     groupID,
		"creatorId":   creatorID,
		"title":       event.Title,
		"description": event.Description,
		"eventDate":   event.EventDate,
		"goingCount":  0,
		"notGoingCount": 0,
	})
}

func GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groupIDStr := r.PathValue("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid group ID",
		})
		return
	}

	rows, err := sqlite.DB.Query(`
		SELECT e.id, e.creator_id, e.title, e.description, e.event_date, e.created_at,
			   COUNT(CASE WHEN r.rsvp_status = 'going' THEN 1 END) as going_count,
			   COUNT(CASE WHEN r.rsvp_status = 'not going' THEN 1 END) as not_going_count
		FROM group_events e
		LEFT JOIN group_event_RSVP r ON e.id = r.event_id
		WHERE e.group_id = ?
		GROUP BY e.id
		ORDER BY e.event_date`, groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch events",
		})
		log.Printf("Error fetching events: %v", err)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var event struct {
			ID          int       `json:"id"`
			CreatorID   int       `json:"creator_id"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			EventDate   time.Time `json:"event_date"`
			CreatedAt   time.Time `json:"created_at"`
			GoingCount  int       `json:"going_count"`
			NotGoing    int       `json:"not_going_count"`
		}

		if err := rows.Scan(
			&event.ID, &event.CreatorID, &event.Title, &event.Description,
			&event.EventDate, &event.CreatedAt, &event.GoingCount, &event.NotGoing); err != nil {
			log.Printf("Error scanning event: %v", err)
			continue
		}

		events = append(events, map[string]interface{}{
			"id":          event.ID,
			"creatorId":   event.CreatorID,
			"title":      event.Title,
			"description": event.Description,
			"eventDate":   event.EventDate,
			"createdAt":   event.CreatedAt,
			"goingCount":  event.GoingCount,
			"notGoingCount": event.NotGoing,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func RespondToGroupEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	eventIDStr := r.PathValue("eventId")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid event ID",
		})
		return
	}

	var response struct {
		UserID int    `json:"userId"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request data",
		})
		return
	}

	// Delete any existing response
	_, err = sqlite.DB.Exec("DELETE FROM group_event_RSVP WHERE event_id = ? AND user_id = ?",
		eventID, response.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to update response",
		})
		return
	}

	// Insert new response
	_, err = sqlite.DB.Exec(`
		INSERT INTO group_event_RSVP (event_id, user_id, rsvp_status)
		VALUES (?, ?, ?)`,
		eventID, response.UserID, response.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to save response",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Response saved successfully",
	})
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    groupIDStr := r.PathValue("id")
    groupID, err := strconv.Atoi(groupIDStr)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid group ID",
        })
        return
    }

    // Get current user
    username, err := util.GetUsernameFromSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Unauthorized",
        })
        return
    }

    // Check if user is the creator
    var creatorID int
    err = sqlite.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Group not found",
        })
        return
    }

    var userID int
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil || userID != creatorID {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Only the creator can edit the group",
        })
        return
    }

    var updateData struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid request data",
        })
        return
    }

    // Update the group
    _, err = sqlite.DB.Exec(
        "UPDATE groups SET title = ?, description = ? WHERE id = ?",
        updateData.Title, updateData.Description, groupID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to update group",
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Group updated successfully",
    })
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    groupIDStr := r.PathValue("id")
    groupID, err := strconv.Atoi(groupIDStr)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid group ID",
        })
        return
    }

    // Get current user
    username, err := util.GetUsernameFromSession(r)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Unauthorized",
        })
        return
    }

    // Check if user is the creator
    var creatorID int
    err = sqlite.DB.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Group not found",
        })
        return
    }

    var userID int
    err = sqlite.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil || userID != creatorID {
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Only the creator can delete the group",
        })
        return
    }

    // Start a transaction
    tx, err := sqlite.DB.Begin()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to start transaction",
        })
        return
    }

    // Delete related records first
    deleteQueries := []string{
        "DELETE FROM group_event_RSVP WHERE event_id IN (SELECT id FROM group_events WHERE group_id = ?)",
        "DELETE FROM group_events WHERE group_id = ?",
        "DELETE FROM group_members WHERE group_id = ?",
        "DELETE FROM notifications WHERE group_id = ?",
        "DELETE FROM posts WHERE group_id = ?",
        "DELETE FROM groups WHERE id = ?",
    }

    for _, query := range deleteQueries {
        _, err = tx.Exec(query, groupID)
        if err != nil {
            tx.Rollback()
            log.Printf("Error executing delete query: %v", err)
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Failed to delete group",
            })
            return
        }
    }

    // Commit the transaction
    if err = tx.Commit(); err != nil {
        tx.Rollback()
        log.Printf("Error committing transaction: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to delete group",
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Group deleted successfully",
    })
}
