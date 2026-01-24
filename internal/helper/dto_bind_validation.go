package helper

import (
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/labstack/echo/v4"
)

func BindAndValidate(ctx echo.Context, i interface{}) error {

	if err := (&echo.DefaultBinder{}).BindHeaders(ctx, i); err != nil {
		return err
	}

	if err := ctx.Bind(i); err != nil {
		return err
	}

	if err := modifiers.New().Struct(ctx.Request().Context(), i); err != nil {
		return err
	}

	if err := ctx.Validate(i); err != nil {
		return err
	}

	return nil
}
