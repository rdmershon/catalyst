package tickets

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"

	"github.com/SecurityBrewery/catalyst/generated/restapi/api"
)

// RunArtifactEndpoint executes the core logic of the related
// route endpoint.
func RunArtifactEndpoint(handler func(ctx context.Context, params *RunArtifactParams) *api.Response) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// generate params from request
		params := NewRunArtifactParams()
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

// NewRunArtifactParams creates a new RunArtifactParams object
// with the default values initialized.
func NewRunArtifactParams() *RunArtifactParams {
	var ()
	return &RunArtifactParams{}
}

// RunArtifactParams contains all the bound params for the run artifact operation
// typically these are obtained from a http.Request
//
// swagger:parameters runArtifact
type RunArtifactParams struct {

	/*
	  Required: true
	  In: path
	*/
	Automation string
	/*Ticket ID
	  Required: true
	  In: path
	*/
	ID int64
	/*
	  Required: true
	  In: path
	*/
	Name string
}

// ReadRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *RunArtifactParams) ReadRequest(ctx *gin.Context) error {
	var res []error

	rAutomation := []string{ctx.Param("automation")}
	if err := o.bindAutomation(rAutomation, true); err != nil {
		res = append(res, err)
	}

	rID := []string{ctx.Param("id")}
	if err := o.bindID(rID, true); err != nil {
		res = append(res, err)
	}

	rName := []string{ctx.Param("name")}
	if err := o.bindName(rName, true); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *RunArtifactParams) bindAutomation(rawData []string, hasKey bool) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	o.Automation = raw

	return nil
}

func (o *RunArtifactParams) bindID(rawData []string, hasKey bool) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("id", "path", "int64", raw)
	}
	o.ID = value

	return nil
}

func (o *RunArtifactParams) bindName(rawData []string, hasKey bool) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	o.Name = raw

	return nil
}