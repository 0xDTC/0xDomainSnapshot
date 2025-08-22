package handlers

import (
	"dns-inventory/internal/database"
	"dns-inventory/internal/services"
	"encoding/json"
	"html/template"
	"net/http"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
	templates   *template.Template
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		templates:   loadTemplates(),
	}
}

// UserPageData represents data for the user page
type UserPageData struct {
	Title  string            `json:"title"`
	Users  []database.User   `json:"users"`
	Groups []string          `json:"groups"`
	Stats  map[string]interface{} `json:"stats"`
}

// HandleUsersPage renders the users management page
func (h *UserHandler) HandleUsersPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all users
	users, err := h.userService.GetUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Get unique groups
	groups, err := h.userService.GetUniqueGroups()
	if err != nil {
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}

	// Get user stats
	stats, err := h.userService.GetUserStats()
	if err != nil {
		http.Error(w, "Failed to get user stats", http.StatusInternalServerError)
		return
	}

	data := UserPageData{
		Title:  "User Management",
		Users:  users,
		Groups: groups,
		Stats:  stats,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleUsersAPI handles user API requests
func (h *UserHandler) HandleUsersAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetUsers(w, r)
	case http.MethodPost:
		h.handleCreateUser(w, r)
	case http.MethodPut:
		h.handleUpdateUser(w, r)
	case http.MethodDelete:
		h.handleDeleteUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetUsers returns users data as JSON
func (h *UserHandler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	group := r.URL.Query().Get("group")

	var users []database.User
	var err error

	if search != "" {
		users, err = h.userService.SearchUsers(search)
	} else if group != "" {
		users, err = h.userService.GetUsersByGroup(group)
	} else {
		users, err = h.userService.GetUsers()
	}

	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users": users,
		"total": len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateUser creates a new user
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Group string `json:"group"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := h.userService.ValidateUser(req.Name, req.Group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(req.Name, req.Group, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// handleUpdateUser updates an existing user
func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Group    string `json:"group"`
		Email    string `json:"email"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := h.userService.ValidateUser(req.Name, req.Group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdateUser(req.ID, req.Name, req.Group, req.Email, req.IsActive); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleDeleteUser deletes a user
func (h *UserHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := parseInt(r.URL.Query().Get("id"), 0)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}