package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) (int, error)
	UpdateUserVerified(ctx context.Context, userID int) (*entity.UserEntity, error)
	UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error
	GetUserByID(ctx context.Context, userID int) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error

	// modul user
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int, int, error)
	GetCustomerByID(ctx context.Context, customerID int) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) (int, error)
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int) error
	GetUsersByIDs(ctx context.Context, userIDs []int) ([]entity.UserEntity, error)
}

type UserRepository struct {
	db *gorm.DB
}

func (u *UserRepository) GetUsersByIDs(ctx context.Context, userIDs []int) ([]entity.UserEntity, error) {
	if len(userIDs) == 0 {
		return []entity.UserEntity{}, nil
	}

	modelUsers := []models.User{}
	// Gunakan GORM untuk mengambil multiple users berdasarkan ID dalam satu query
	if err := u.db.WithContext(ctx).Where("id IN (?)", userIDs).Preload("Roles").Find(&modelUsers).Error; err != nil {
		log.Errorf("[UserRepository-1] GetUsersByIDs: %v", err)
		return nil, err
	}

	if len(modelUsers) == 0 {
		// Jika tidak ada user ditemukan, kembalikan slice kosong tanpa error, atau error 404 jika memang tidak boleh kosong
		// Untuk kasus bulk fetch, biasanya lebih baik mengembalikan slice kosong jika tidak ada yang ditemukan
		return []entity.UserEntity{}, nil
	}

	respEntities := make([]entity.UserEntity, 0, len(modelUsers))
	for _, userMdl := range modelUsers {
		roleName := ""
		roleId := 0
		if len(userMdl.Roles) > 0 {
			roleName = userMdl.Roles[0].Name
			roleId = userMdl.Roles[0].ID
		}
		respEntities = append(respEntities, entity.UserEntity{
			ID:         userMdl.ID,
			Name:       userMdl.Name,
			Email:      userMdl.Email,
			Address:    userMdl.Address,
			Photo:      userMdl.Photo,
			Lat:        userMdl.Lat,
			Lng:        userMdl.Lng,
			Phone:      userMdl.Phone,
			RoleID:     roleId,
			RoleName:   roleName,
			IsVerified: userMdl.IsVerified,
		})
	}
	return respEntities, nil
}

// DeleteCustomer implements IUserRepository.
func (u *UserRepository) DeleteCustomer(ctx context.Context, customerID int) error {
	modelUser := models.User{}

	if err := u.db.Where("id =?", customerID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[UserRepository-1] DeleteCustomer: User not found")
			return err
		}
		log.Errorf("[UserRepository-2] DeleteCustomer: %v", err)
		return err
	}

	if err := u.db.Delete(&modelUser).Error; err != nil {
		log.Errorf("[UserRepository-3] DeleteCustomer: %v", err)
		return err
	}
	return nil
}

// UpdateCustomer implements IUserRepository.
func (u *UserRepository) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	modelRole := models.Role{}

	if err := u.db.Where("id =?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Fatalf("[UserRepository-1] UpdateCustomer: %v", err)
		return err
	}

	modelUser := models.User{}

	if err := u.db.Where("id =?", req.ID).First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[UserRepository-2] UpdateCustomer: User not found")
			return err
		}
		log.Errorf("[UserRepository-3] UpdateCustomer: %v", err)
		return err
	}

	modelUser.Name = req.Name
	modelUser.Email = req.Email
	modelUser.Phone = req.Phone
	modelUser.Roles = []models.Role{modelRole}

	if req.Address != "" {
		modelUser.Address = req.Address
	}

	if req.Lat != "" {
		modelUser.Lat = req.Lat
	}

	if req.Lng != "" {
		modelUser.Lng = req.Lng
	}
	if req.Photo != "" {
		modelUser.Photo = req.Photo
	}

	if req.Password != "" {
		modelUser.Password = req.Password
	}

	if err := u.db.Save(&modelUser).Error; err != nil {
		log.Errorf("[UserRepository-4] UpdateCustomer: %v", err)
		return err
	}

	return nil

}

// CreateCustomer implements IUserRepository.
func (u *UserRepository) CreateCustomer(ctx context.Context, req entity.UserEntity) (int, error) {
	modelRole := models.Role{}

	if err := u.db.WithContext(ctx).Where("id = ?", req.RoleID).First(&modelRole).Error; err != nil {
		log.Errorf("[UserRepository-1] CreateCustomer: %v", err)
		return 0, err
	}

	userMdl := models.User{
		Name:       req.Name,
		Email:      req.Email,
		Address:    req.Address,
		Photo:      req.Photo,
		Lat:        req.Lat,
		Lng:        req.Lng,
		Phone:      req.Phone,
		Password:   req.Password,
		Roles:      []models.Role{modelRole},
		IsVerified: true,
	}

	if err := u.db.WithContext(ctx).Create(&userMdl).Error; err != nil {
		log.Errorf("[UserRepository-2] CreateCustomer: %v", err)
		return 0, err
	}

	return userMdl.ID, nil
}

// GetCustomerByID implements IUserRepository.
func (u *UserRepository) GetCustomerByID(ctx context.Context, customerID int) (*entity.UserEntity, error) {
	userMdl := models.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", customerID).Preload("Roles").First(&userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] GetCustomerByID: %v", err)
			return nil, err
		}
		log.Errorf("[UserRepository-2] GetCustomerByID: %v", err)
		return nil, err
	}

	var roleMdl models.Role // Deklarasi roleMdl di sini
	for _, role := range userMdl.Roles {
		roleMdl = role // Assign role pertama ke roleMdl
		break         // Ambil hanya role pertama
	}

	return &entity.UserEntity{
		ID:      userMdl.ID,
		Name:    userMdl.Name,
		Email:   userMdl.Email,
		Address: userMdl.Address,
		Photo:   userMdl.Photo,
		Lat:     userMdl.Lat,
		Lng:     userMdl.Lng,
		Phone:   userMdl.Phone,
		RoleID:  roleMdl.ID, // Gunakan ID dari roleMdl yang ditemukan
	}, nil
}

// GetCustomerAll implements IUserRepository.
func (u *UserRepository) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int, int, error) {
	modelUsers := []models.User{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	// Base query with JOINs to filter by "Customer" role first.
	// This is the most critical optimization.
	queryBuilder := u.db.Model(&models.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", "user").
		Where("(users.name || ' ' || users.phone || ' ' || users.email) ILIKE ?", "%"+query.Search+"%")

	// 1. Execute the COUNT query on the filtered dataset.
	if err := queryBuilder.Count(&countData).Error; err != nil {
		log.Errorf("[UserRepository-1] GetCustomerAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := 0
	if countData > 0 {
		totalPage = int(math.Ceil(float64(countData) / float64(query.Limit)))
	}

	// 2. Execute the FIND query WITH Preload, Order, Limit, and Offset.
	// Preload is still useful to fetch the Role details for the response.
	if err := queryBuilder.Preload("Roles").Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelUsers).Error; err != nil {
		log.Errorf("[UserRepository-2] GetCustomerAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelUsers) < 1 {
		err := errors.New("404")
		log.Errorf("[UserRepository-3] GetCustomerAll: No Customer found")
		return nil, 0, 0, err
	}

	respEntities := []entity.UserEntity{}

	for _, v := range modelUsers {
		roleName := ""
		for _, x := range v.Roles {
			roleName = x.Name
		}

		respEntities = append(respEntities, entity.UserEntity{
			ID:       v.ID,
			Name:     v.Name,
			Email:    v.Email,
			RoleName: roleName,
			Phone:    v.Phone,
			Photo:    v.Photo,
		})
	}

	return respEntities, int(countData), totalPage, nil
}

// UpdateDataUser implements IUserRepository.
func (u *UserRepository) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	userMdl := models.User{
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	if err := u.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", req.ID).Updates(userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] UpdateDataUser: %v", err)
			return err
		}
		log.Errorf("[UserRepository-1] UpdateDataUser: %v", err)
		return err
	}

	userMdl.Lat = req.Lat
	userMdl.Lng = req.Lng
	userMdl.Address = req.Address
	userMdl.Phone = req.Phone

	if err := u.db.UpdateColumns(&userMdl).Error; err != nil {
		log.Errorf("[UserRepository-3] UpdateDataUser: %v", err)
		return err
	}

	return nil

}

// GetUserByID implements IUserRepository.
func (u *UserRepository) GetUserByID(ctx context.Context, userID int) (*entity.UserEntity, error) {
	modelUser := models.User{}
	if err := u.db.Where("id =? AND is_verified = true", userID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] GetUserByID: %v", err)
			return nil, err
		}
		log.Errorf("[UserRepository-2] GetUserByID: %v", err)
		return nil, err
	}

	return &entity.UserEntity{
		ID:       modelUser.ID,
		Email:    modelUser.Email,
		Name:     modelUser.Name,
		RoleName: modelUser.Roles[0].Name,
		Lat:      modelUser.Lat,
		Lng:      modelUser.Lng,
		Address:  modelUser.Address,
		Phone:    modelUser.Phone,
		Photo:    modelUser.Photo,
	}, nil
}

// UpdatePasswordByID implements IUserRepository.
func (u *UserRepository) UpdatePasswordByID(ctx context.Context, req entity.UserEntity) error {
	userMdl := models.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", req.ID).First(&userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] UpdatePasswordByID: %v", err)
			return err
		}
		log.Errorf("[UserRepository-2] UpdatePasswordByID: %v", err)
		return err
	}

	userMdl.Password = req.Password

	if err := u.db.WithContext(ctx).Save(&userMdl).Error; err != nil {
		log.Errorf("[UserRepository-3] UpdatePasswordByID: %v", err)
		return err
	}

	return nil
}

// UpdateUserVerified implements IUserRepository.
func (u *UserRepository) UpdateUserVerified(ctx context.Context, userID int) (*entity.UserEntity, error) {
	modelUser := models.User{}

	if err := u.db.WithContext(ctx).Where("id = ?", userID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] UpdateUserVerified: %v", err)
			return nil, err
		}
		log.Errorf("[UserRepository-2] UpdateUserVerified: %v", err)
		return nil, err
	}

	modelUser.IsVerified = true

	if err := u.db.WithContext(ctx).Save(&modelUser).Error; err != nil {
		log.Errorf("[UserRepository-3] UpdateUserVerified: %v", err)
		return nil, err
	}

	return &entity.UserEntity{
		ID:         userID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		RoleName:   modelUser.Roles[0].Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

// CreateUserAccount implements IUserRepository.
func (u *UserRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) (int, error) {
	var roleId int

	// Check if email already exists
	existingUser := models.User{}
	errCheck := u.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error
	if errCheck == nil {
		return 0, errors.New("Email sudah terdaftar")
	}
	if !errors.Is(errCheck, gorm.ErrRecordNotFound) {
		log.Errorf("[UserRepository-0] CreateUserAccount: Failed to check existing email: %v", errCheck)
		return 0, errCheck
	}

	if err := u.db.WithContext(ctx).Select("id").
		Where("name = ?", "user").
		Model(&models.Role{}).
		Scan(&roleId).
		Error; err != nil {
		log.Errorf("[UserRepository-1] CreateUserAccount : %v", err)
		return 0, err
	}

	userMdl := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Roles:    []models.Role{{ID: roleId}},
	}

	if err := u.db.WithContext(ctx).Create(&userMdl).Error; err != nil {
		log.Errorf("[UserRepository-2] CreateUserAccount : %v", err)
		return 0, err
	}

	verifyMdl := models.VerificationToken{
		UserID:    userMdl.ID,
		Token:     req.Token,
		TokenType: "email_verification",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := u.db.WithContext(ctx).Create(&verifyMdl).Error; err != nil {
		log.Errorf("[UserRepository-3] CreateUserAccount : %v", err)
		return 0, err
	}

	return userMdl.ID, nil

}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	userMdl := models.User{}

	if err := u.db.WithContext(ctx).Where("email = ? and is_verified = ?", email, true).Preload("Roles").First(&userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[UserRepository-1] GetUserByEmail : %v", err)
			return nil, err
		}

		log.Errorf("[UserRepository-1] GetUserByEmail : %v", err)
		return nil, err
	}

	userE := entity.UserEntity{
		ID:         userMdl.ID,
		Name:       userMdl.Name,
		Email:      userMdl.Email,
		Password:   userMdl.Password,
		Phone:      userMdl.Phone,
		Photo:      userMdl.Photo,
		Address:    userMdl.Address,
		Lat:        userMdl.Lat,
		Lng:        userMdl.Lng,
		IsVerified: userMdl.IsVerified,
		RoleName:   userMdl.Roles[0].Name,
	}

	return &userE, nil
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}
