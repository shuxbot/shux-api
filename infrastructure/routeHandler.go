package infrastructure

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/shuxbot/shux-api/application"
	"github.com/shuxbot/shux-api/domain"
)


type routeHandler struct {
	userApp *application.UserApp
}

func result(success bool, err error) map[string]interface{} {
	status := make(map[string]interface{})
	status["error"] = err
	status["success"] = success

	return status
}

func (h *routeHandler) GetUser(c *fiber.Ctx) error {
	u, err := h.userApp.Get(c.Params("user_id"), c.Params("server_id"))
	if err != nil {
		return c.Status(404).JSON(result(false, err))
	}
	return c.JSON(u)
}

func (h *routeHandler) DeleteUser(c *fiber.Ctx) error {
	err := h.userApp.Delete(c.Params("user_id"), c.Params("server_id"))
	if err != nil {
		return c.Status(404).JSON(result(false, err))
	}
	return c.JSON(result(true, nil))
}

func (h *routeHandler) CreateUser(c *fiber.Ctx) error {
	var u domain.User
	json.Unmarshal(c.Body(), &u)
	u.UserId = c.Params("user_id")

	err := h.userApp.Create(&u, c.Params("server_id"))

	if err != nil {
		return c.Status(404).JSON(result(false, err))
	}

	return c.JSON(result(true, nil))
}

func NewRouteHandler(userApp *application.UserApp) *routeHandler {
	return &routeHandler{userApp: userApp}
}
