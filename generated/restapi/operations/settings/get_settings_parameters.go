package settings

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/errors"

	"github.com/SecurityBrewery/catalyst/generated/restapi/api"
)

// GetSettingsEndpoint executes the core logic of the related
// route endpoint.
func GetSettingsEndpoint(handler func(ctx context.Context) *api.Response) gin.HandlerFunc {
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

// NewGetSettingsParams creates a new GetSettingsParams object
// with the default values initialized.
func NewGetSettingsParams() *GetSettingsParams {
	var ()
	return &GetSettingsParams{}
}

// GetSettingsParams contains all the bound params for the get settings operation
// typically these are obtained from a http.Request
//
// swagger:parameters getSettings
type GetSettingsParams struct {
}

// ReadRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *GetSettingsParams) ReadRequest(ctx *gin.Context) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}