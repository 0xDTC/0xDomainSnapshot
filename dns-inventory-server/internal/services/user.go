package services

import (
	"dns-inventory/internal/database"
	"fmt"
	"strings"
)

// UserService handles user-related operations
type UserService struct {
	db *database.FileDB
}

// NewUserService creates a new user service
func NewUserService(db *database.FileDB) *UserService {
	return &UserService{
		db: db,
	}
}

// GetUsers retrieves all users
func (s *UserService) GetUsers() ([]database.User, error) {
	return s.db.GetUsers()
}

// CreateUser creates a new user
func (s *UserService) CreateUser(name, group, email string) (*database.User, error) {
	if name == "" {
		return nil, fmt.Errorf("user name is required")
	}
	
	if group == "" {
		return nil, fmt.Errorf("user group is required")
	}
	
	// Check if user already exists
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	for _, user := range users {
		if strings.EqualFold(user.Name, name) {
			return nil, fmt.Errorf("user with name '%s' already exists", name)
		}
	}
	
	user := &database.User{
		Name:     name,
		Group:    group,
		Email:    email,
		IsActive: true,
	}
	
	if err := s.db.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id int, name, group, email string, isActive bool) error {
	if name == "" {
		return fmt.Errorf("user name is required")
	}
	
	if group == "" {
		return fmt.Errorf("user group is required")
	}
	
	// Check if another user with the same name exists
	users, err := s.db.GetUsers()
	if err != nil {
		return err
	}
	
	for _, user := range users {
		if user.ID != id && strings.EqualFold(user.Name, name) {
			return fmt.Errorf("user with name '%s' already exists", name)
		}
	}
	
	user := &database.User{
		ID:       id,
		Name:     name,
		Group:    group,
		Email:    email,
		IsActive: isActive,
	}
	
	return s.db.UpdateUser(user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id int) error {
	// TODO: Consider what to do with assignments when user is deleted
	// For now, we'll just delete the user and leave assignments orphaned
	return s.db.DeleteUser(id)
}

// GetUsersByGroup retrieves users filtered by group
func (s *UserService) GetUsersByGroup(group string) ([]database.User, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	var filteredUsers []database.User
	for _, user := range users {
		if group == "" || strings.EqualFold(user.Group, group) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	
	return filteredUsers, nil
}

// GetUniqueGroups returns all unique groups
func (s *UserService) GetUniqueGroups() ([]string, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	groupSet := make(map[string]bool)
	for _, user := range users {
		if user.Group != "" {
			groupSet[user.Group] = true
		}
	}
	
	var groups []string
	for group := range groupSet {
		groups = append(groups, group)
	}
	
	return groups, nil
}

// GetActiveUsers retrieves only active users
func (s *UserService) GetActiveUsers() ([]database.User, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	var activeUsers []database.User
	for _, user := range users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}
	
	return activeUsers, nil
}

// SearchUsers searches users by name or group
func (s *UserService) SearchUsers(query string) ([]database.User, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	if query == "" {
		return users, nil
	}
	
	query = strings.ToLower(query)
	var filteredUsers []database.User
	
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Name), query) ||
		   strings.Contains(strings.ToLower(user.Group), query) ||
		   strings.Contains(strings.ToLower(user.Email), query) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	
	return filteredUsers, nil
}

// GetUserStats returns user statistics
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"total_users":  len(users),
		"active_users": 0,
		"groups":       make(map[string]int),
	}
	
	groupCounts := make(map[string]int)
	activeCount := 0
	
	for _, user := range users {
		if user.IsActive {
			activeCount++
		}
		groupCounts[user.Group]++
	}
	
	stats["active_users"] = activeCount
	stats["groups"] = groupCounts
	
	return stats, nil
}

// ValidateUser validates user data
func (s *UserService) ValidateUser(name, group string) error {
	if name == "" {
		return fmt.Errorf("user name is required")
	}
	
	if len(name) < 2 || len(name) > 100 {
		return fmt.Errorf("user name must be between 2 and 100 characters")
	}
	
	if group == "" {
		return fmt.Errorf("user group is required")
	}
	
	if len(group) < 2 || len(group) > 50 {
		return fmt.Errorf("user group must be between 2 and 50 characters")
	}
	
	return nil
}