package automations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"

	"github.com/SecurityBrewery/catalyst/generated/models"
	"github.com/SecurityBrewery/catalyst/generated/restapi/api"
)

// CreateAutomationEndpoint executes the core logic of the related
// route endpoint.
func CreateAutomationEndpoint(handler func(ctx context.Context, params *CreateAutomationParams) *api.Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// generate params from request
		params := NewCreateAutomationParams()
		err := params.ReadRequest(ctx)
		if err != nil {
			errObj := err.(*errors.CompositeError)
			ctx.Writer.Header().Set("Content-Type", "application/problem+json")
			ctx.JSON(int(errObj.Code()), gin.H{"error": errObj.Error()})
			return
		}

		resp := handler(ctx, params)

		switch resp.Code {
		case http.StatusNoContent:
			ctx.AbortWithStatus(resp.Code)
		default:
			ctx.JSON(resp.Code, resp.Body)
		}
	}
}

// NewCreateAutomationParams creates a new CreateAutomationParams object
// with the default values initialized.
func NewCreateAutomationParams() *CreateAutomationParams {
	var ()
	return &CreateAutomationParams{}
}

// CreateAutomationParams contains all the bound params for the create automation operation
// typically these are obtained from a http.Request
//
// swagger:parameters createAutomation
type CreateAutomationParams struct {

	/*New automation
	  Required: true
	  In: body
	*/
	Automation *models.AutomationForm
}

// ReadRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *CreateAutomationParams) ReadRequest(ctx *gin.Context) error {
	var res []error

	if runtime.HasBody(ctx.Request) {
		var body models.AutomationForm
		if err := ctx.BindJSON(&body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("automation", "body", ""))
			} else {
				res = append(res, errors.NewParseError("automation", "body", "", err))
			}

		} else {
			o.Automation = &body
		}
	} else {
		res = append(res, errors.Required("automation", "body", ""))
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}