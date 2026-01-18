package service

import (
	"errors"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
)

// MockGroupRepository is a mock implementation for tests
// It implements repository.GroupRepositoryInterface.
type MockGroupRepository struct {
	groups      map[uint]*models.Group
	handles     map[string]*models.Group
	memberships map[uint]map[uint]models.GroupRole
	nextID      uint
}

func NewMockGroupRepository() *MockGroupRepository {
	return &MockGroupRepository{
		groups:      make(map[uint]*models.Group),
		handles:     make(map[string]*models.Group),
		memberships: make(map[uint]map[uint]models.GroupRole),
		nextID:      1,
	}
}

func (m *MockGroupRepository) Create(group *models.Group) error {
	if group.ID == 0 {
		group.ID = m.nextID
		m.nextID++
	}
	m.groups[group.ID] = group
	if group.Handle != nil {
		m.handles[*group.Handle] = group
	}
	return nil
}

func (m *MockGroupRepository) FindByID(id uint) (*models.Group, error) {
	if g, ok := m.groups[id]; ok {
		return g, nil
	}
	return nil, errors.New("record not found")
}

func (m *MockGroupRepository) FindByHandle(handle string) (*models.Group, error) {
	if g, ok := m.handles[handle]; ok {
		return g, nil
	}
	return nil, errors.New("record not found")
}

func (m *MockGroupRepository) SearchPublicGroups(query string, limit int) ([]models.Group, error) {
	var out []models.Group
	for _, g := range m.groups {
		if g.IsPublic {
			out = append(out, *g)
		}
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *MockGroupRepository) AddMember(groupID, userID uint, role models.GroupRole) error {
	if _, ok := m.memberships[groupID]; !ok {
		m.memberships[groupID] = make(map[uint]models.GroupRole)
	}
	m.memberships[groupID][userID] = role
	return nil
}

func (m *MockGroupRepository) RemoveMember(groupID, userID uint) error {
	if gm, ok := m.memberships[groupID]; ok {
		delete(gm, userID)
	}
	return nil
}

func (m *MockGroupRepository) GetMembers(groupID uint) ([]models.User, error) {
	var users []models.User
	if gm, ok := m.memberships[groupID]; ok {
		for uid := range gm {
			users = append(users, models.User{ID: uid})
		}
	}
	return users, nil
}

func (m *MockGroupRepository) IsMember(groupID, userID uint) (bool, error) {
	if gm, ok := m.memberships[groupID]; ok {
		_, ok := gm[userID]
		return ok, nil
	}
	return false, nil
}

func (m *MockGroupRepository) GetMemberRole(groupID, userID uint) (models.GroupRole, error) {
	if gm, ok := m.memberships[groupID]; ok {
		if role, ok := gm[userID]; ok {
			return role, nil
		}
	}
	return "", errors.New("record not found")
}

func (m *MockGroupRepository) GetUserGroups(userID uint) ([]models.Group, error) {
	var out []models.Group
	for gid, gm := range m.memberships {
		if _, ok := gm[userID]; ok {
			if g, ok := m.groups[gid]; ok {
				out = append(out, *g)
			}
		}
	}
	return out, nil
}
