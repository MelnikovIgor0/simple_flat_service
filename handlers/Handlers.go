package handlers

import (
	"bootcamp_task/cache"
	"bootcamp_task/storage/entities"
	"bootcamp_task/storage/storages"
	"database/sql"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type Handlers struct {
	cache     *cache.Cache
	storage   *storages.Storage
	validator *validator.Validate
}

func NewHandlers(cache *cache.Cache, storage *storages.Storage) *Handlers {
	h := Handlers{
		cache,
		storage,
		validator.New(),
	}
	return &h
}

func (h *Handlers) getUserType(value string) (bool, error) {
	switch value {
	case "client":
		return false, nil
	case "moderator":
		return true, nil
	default:
		return false, errors.New("invalid user type")
	}
}

func (h *Handlers) DummyLogin(c *fiber.Ctx) error {
	value := c.Query("user_type")
	admin, err := h.getUserType(value)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	token, err := h.cache.CreateSession("", admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=6,max=50"`
	UserType string `json:"user_type" validate:"required,max=9"`
}

func (h *Handlers) Register(c *fiber.Ctx) error {
	var req registerRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if _, err := h.storage.GetUser(req.Email); errors.Is(err, nil) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user with same email already exists"})
	}
	admin, err := h.getUserType(req.UserType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	uid, err := h.storage.CreateUser(req.Email, req.Password, admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user_id": uid})
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,max=50"`
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	var req loginRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	user, err := h.storage.GetUser(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if user.Password != req.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "wrong password"})
	}
	token, err := h.cache.CreateSession(user.Id, user.IsAdmin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
}

func (h *Handlers) validateSession(id string) (userId string, admin bool, valid bool, err error) {
	userId, admin, err = h.cache.GetSession(id)
	if errors.Is(err, redis.Nil) {
		return "", false, false, nil
	}
	if err != nil {
		return "", false, false, err
	}
	return userId, admin, true, nil
}

type createHomeRequest struct {
	Address   string `json:"address" validate:"required,max=120"`
	Year      int    `json:"year" validate:"required,min=1"`
	Developer string `json:"developer" validate:"max=30"`
}

func (h *Handlers) CreateHome(c *fiber.Ctx) error {
	sessionId := c.Get("auth")
	userId, admin, valid, err := h.validateSession(sessionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if !admin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you have no permission to create house"})
	}
	var req createHomeRequest
	err = c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	home, errCreation := h.storage.CreateHome(req.Address, req.Year, req.Developer, userId)
	if errCreation != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"house": map[string]interface{}{
		"id":         home.Id,
		"year":       home.Year,
		"address":    home.Address,
		"developer":  home.Developer,
		"created_at": home.CreatedAt,
		"updated_at": home.UpdatedAt,
	}})
}

type createFlatRequest struct {
	HouseId int `json:"house_id" validate:"required,min=1"`
	FlatId  int `json:"id" validate:"required,min=1"`
	Price   int `json:"price" validate:"required,min=1"`
	Rooms   int `json:"rooms" validate:"required,min=1"`
}

func (h *Handlers) CreateFlat(c *fiber.Ctx) error {
	sessionId := c.Get("auth")
	_, _, valid, err := h.validateSession(sessionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req createFlatRequest
	err = c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if _, err := h.storage.GetLastHomeUpdate(req.HouseId); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "house with specified id not found"})
	}
	flat, err := h.storage.CreateFlat(
		req.FlatId,
		req.HouseId,
		req.Price,
		req.Rooms,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"flat": flat})
}

type updateFlatRequest struct {
	HouseId int    `json:"house_id" validate:"required,min=1"`
	FlatId  int    `json:"id" validate:"required,min=1"`
	Price   int    `json:"price" validate:"required,min=1"`
	Rooms   int    `json:"rooms" validate:"required,min=1"`
	Status  string `json:"status" validate:"required,min=1"`
}

func (h *Handlers) setStatus(s string) (entities.ModerationStatus, error) {
	switch s {
	case "created":
		return entities.CREATED, nil
	case "declined":
		return entities.DECLINED, nil
	case "approved":
		return entities.APPROVED, nil
	case "on_moderation":
		return entities.ON_MODERATION, nil
	default:
		return entities.CREATED, errors.New("invalid moderation status")
	}
}

func (h *Handlers) UpdateFlat(c *fiber.Ctx) error {
	sessionId := c.Get("auth")
	userId, admin, valid, err := h.validateSession(sessionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if !admin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you have no permission to update this house"})
	}
	var req updateFlatRequest
	err = c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	status, err := h.setStatus(req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	reviewer, err := h.storage.GetHomeReviewer(req.HouseId)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "house with specified id not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "internal server error"})
	}
	if reviewer != userId {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only house creator able to review flats placed in this house"})
	}
	flat, err := h.storage.UpdateFlat(
		req.FlatId,
		req.HouseId,
		req.Price,
		req.Rooms,
		status,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"flat": flat})
}

func (h *Handlers) getHouseFlats(houseId int, admin bool) ([]entities.Flat, error) {
	if admin {
		return h.storage.FilterFlats(houseId, true)
	}
	lastUpdate, err := h.storage.GetLastHomeUpdate(houseId)
	if err != nil {
		return nil, err
	}
	cacheName := strconv.FormatInt(lastUpdate.Unix(), 10) + "-" + strconv.Itoa(houseId)
	var flats []entities.Flat
	flats, err = h.cache.GetFlatsCache(cacheName)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	} else if errors.Is(err, redis.Nil) {
		r, err2 := h.storage.FilterFlats(houseId, false)
		if err2 != nil {
			return nil, err2
		}
		if err3 := h.cache.PutFlatsCache(cacheName, r); err3 != nil {
			return nil, err3
		}
		return r, nil
	}
	return flats, nil
}

func (h *Handlers) GetHouseFlats(c *fiber.Ctx) error {
	sessionId := c.Get("auth")
	_, admin, valid, err := h.validateSession(sessionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	houseIdStr := c.Params("id")
	houseId, err := strconv.Atoi(houseIdStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
	}
	flats, err := h.getHouseFlats(houseId, admin)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"flats": flats})
}
