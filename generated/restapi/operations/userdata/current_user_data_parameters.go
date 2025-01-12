package userdata

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/errors"

	"github.com/SecurityBrewery/catalyst/generated/restapi/api"
)

// CurrentUserDataEndpoint executes the core logic of the related
// route endpoint.
func CurrentUserDataEndpoint(handler func(ctx context.Context) *api.Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		resp := handler(ctx)

		switch resp.Code {
		case http.StatusNoContent:
			ctx.AbortWithStatus(resp.Code)
		default:
			ctx.JSON(resp.Code, resp.Body)
		}
	}
}

// NewCurrentUserDataParams creates a new CurrentUserDataParams object
// with the default values initialized.
func NewCurrentUserDataParams() *CurrentUserDataParams {
	var ()
	return &CurrentUserDataParams{}
}

// CurrentUserDataParams contains all the bound params for the current user data operation
// typically these are obtained from a http.Request
//
// swagger:parameters currentUserData
type CurrentUserDataParams struct {
}

// ReadRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *CurrentUserDataParams) ReadRequest(ctx *gin.Context) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
