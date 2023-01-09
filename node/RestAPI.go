package node

import (
	"net/http"

	"github.com/gansoi/gansoi/database"
	"github.com/gin-gonic/gin"
)

type (
	// RestAPI is a generic way to build a REST-based API for the cluster
	// database.
	RestAPI[T any] struct {
		db database.ReadWriter
	}
)

// NewRestAPI will instantiate a new RestAPI.
func NewRestAPI[T any](db database.ReadWriter) *RestAPI[T] {
	return &RestAPI[T]{
		db: db,
	}
}

func (r *RestAPI[T]) new() *T {
	var zero T

	return &zero
}

func reply(c *gin.Context, code int, data string) {
	c.Data(code, "text/plain", []byte(data))
}

func (r *RestAPI[T]) list(c *gin.Context) {
	list := make([]T, 0)

	err := r.db.All(&list, -1, 0, false)
	if err != nil {
		reply(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, list)
}

func (r *RestAPI[T]) create(c *gin.Context) {
	record := r.new()

	err := c.BindJSON(record)
	if err != nil {
		reply(c, http.StatusBadRequest, err.Error())
		return
	}

	validator, ok := any(record).(database.Validator)
	if ok {
		err = validator.Validate(r.db)
		if err != nil {
			reply(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	err = r.db.Save(record)
	if err != nil {
		reply(c, http.StatusInternalServerError, err.Error())
		return
	}

	data := "got it, thanks"

	// If the type is an IDSetter, we should return the new ID.
	obj, valid := any(record).(database.IDSetter)
	if valid {
		data = obj.GetID()
	}

	reply(c, http.StatusAccepted, data)
}

func (r *RestAPI[T]) replace(c *gin.Context) {
	// For now we simply treat this as a create. Should be safe.
	r.create(c)
}

func (r *RestAPI[T]) delete(c *gin.Context) {
	record := r.new()

	err := r.db.One("ID", c.Param("id"), record)
	if err != nil {
		reply(c, http.StatusNotFound, err.Error())
		return
	}

	err = r.db.Delete(record)
	if err != nil {
		reply(c, http.StatusInternalServerError, err.Error())
		return
	}

	reply(c, http.StatusAccepted, "deleted")
}

// Router can be used to assign a Gin routergroup.
func (r *RestAPI[T]) Router(router *gin.RouterGroup) {
	router.GET("/", r.list)
	router.POST("", r.create)
	router.PUT("/:id", r.replace)
	router.DELETE("/:id", r.delete)
}
