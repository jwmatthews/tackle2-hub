package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	ProxiesRoot = "/proxies"
	ProxyRoot   = ProxiesRoot + "/:" + ID
)

const (
	Kind = "kind"
)

//
// ProxyHandler handles proxy resource routes.
type ProxyHandler struct {
	BaseHandler
}

func (h ProxyHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("proxies"))
	routeGroup.GET(ProxiesRoot, h.List)
	routeGroup.GET(ProxiesRoot+"/", h.List)
	routeGroup.POST(ProxiesRoot, h.Create)
	routeGroup.GET(ProxyRoot, h.Get)
	routeGroup.PUT(ProxyRoot, h.Update)
	routeGroup.DELETE(ProxyRoot, h.Delete)
}

// Get godoc
// @summary Get an proxy by ID.
// @description Get an proxy by ID.
// @tags proxies
// @produce json
// @success 200 {object} Proxy
// @router /proxies/{id} [get]
// @param id path string true "Proxy ID"
func (h ProxyHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	proxy := &model.Proxy{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Proxy{}
	r.With(proxy)

	h.Render(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all proxies.
// @description List all proxies.
// @tags proxies
// @produce json
// @success 200 {object} []Proxy
// @router /proxies [get]
func (h ProxyHandler) List(ctx *gin.Context) {
	var list []model.Proxy
	kind := ctx.Query(Kind)
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Proxy{}
	for i := range list {
		r := Proxy{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create an proxy.
// @description Create an proxy.
// @tags proxies
// @accept json
// @produce json
// @success 201 {object} Proxy
// @router /proxies [post]
// @param proxy body Proxy true "Proxy data"
func (h ProxyHandler) Create(ctx *gin.Context) {
	proxy := &Proxy{}
	err := h.Bind(ctx, proxy)
	if err != nil {
		return
	}
	m := proxy.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	proxy.With(m)

	h.Render(ctx, http.StatusCreated, proxy)
}

// Delete godoc
// @summary Delete an proxy.
// @description Delete an proxy.
// @tags proxies
// @success 204
// @router /proxies/{id} [delete]
// @param id path string true "Proxy ID"
func (h ProxyHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	proxy := &model.Proxy{}
	result := h.DB(ctx).First(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update an proxy.
// @description Update an proxy.
// @tags proxies
// @accept json
// @success 204
// @router /proxies/{id} [put]
// @param id path string true "Proxy ID"
// @param proxy body Proxy true "Proxy data"
func (h ProxyHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Proxy{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Proxy REST resource.
type Proxy struct {
	Resource
	Enabled  bool     `json:"enabled"`
	Kind     string   `json:"kind" binding:"oneof=http https"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Excluded []string `json:"excluded"`
	Identity *Ref     `json:"identity"`
}

//
// With updates the resource with the model.
func (r *Proxy) With(m *model.Proxy) {
	r.Resource.With(&m.Model)
	r.Enabled = m.Enabled
	r.Kind = m.Kind
	r.Host = m.Host
	r.Port = m.Port
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	_ = json.Unmarshal(m.Excluded, &r.Excluded)
	if r.Excluded == nil {
		r.Excluded = []string{}
	}
}

//
// Model builds a model.
func (r *Proxy) Model() (m *model.Proxy) {
	m = &model.Proxy{
		Enabled: r.Enabled,
		Kind:    r.Kind,
		Host:    r.Host,
		Port:    r.Port,
	}
	m.ID = r.ID
	m.IdentityID = r.idPtr(r.Identity)
	if r.Excluded != nil {
		m.Excluded, _ = json.Marshal(r.Excluded)
	}

	return
}
