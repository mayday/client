package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	DefaultPort = 8080
)

type File struct {
	Id      int
	Path    string
	Created string `orm:"auto_now_add;type(datetime)"`
	Case    *Case  `orm:"rel(fk)"`
}

type Case struct {
	Id           int       `orm:"auto"`
	Created      time.Time `orm:"auto_now_add;type(datetime)"`
	Updated      time.Time `orm:"auto_now_add;type(datetime)"`
	IsSigned     bool      `orm:"default(false)"`
	IsPrivate    bool      `orm:"default(false)"`
	Token        string
	Config       string `orm:"type(text)"`
	SignedConfig string `orm:"type(text)"`
}

type CaseHandler struct{}

func (handler *CaseHandler) List(request *restful.Request, response *restful.Response) {
	response.WriteEntity(Case{})
}

func (handler *CaseHandler) Create(request *restful.Request, response *restful.Response) {
	c := new(Case)
	err := request.ReadEntity(c)

	o := orm.NewOrm()
	o.Insert(c)

	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(c)
}

func (r *CaseHandler) CreateConfig(request *restful.Request, response *restful.Response) {
	usr := new(Case)
	err := request.ReadEntity(usr)

	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(usr)
}

func init() {
	orm.RegisterModel(new(Case))
	orm.RegisterDataBase("default", "sqlite3", "/tmp/database.db", 30)
	orm.RunCommand()
	// Database alias.
	name := "default"

	// Drop table and re-create.
	force := false

	// Print log.
	verbose := true

	// Error.
	err := orm.RunSyncdb(name, force, verbose)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	handler := &CaseHandler{}

	ws := new(restful.WebService)
	ws.Path("/1/case").
		Doc("Manage support reports").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/config/{report-id}").To(handler.List).
		Doc("get a specific report").
		Operation("findCase").
		Param(ws.PathParameter("report-id", "report identifier").DataType("string")).
		Writes(Case{}))

	ws.Route(ws.POST("").To(handler.Create).
		Doc("create a case").
		Operation("createCase").
		Reads(Case{})) // from the request

	container := restful.NewContainer()
	container.Add(ws)

	config := swagger.Config{
		WebServices:    container.RegisteredWebServices(),
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",
		SwaggerPath:    "/apidocs/",
	}

	swagger.RegisterSwaggerService(config, container)
	log.Printf("start listening on localhost:8080")
	server := &http.Server{Addr: ":8080", Handler: container}
	log.Fatal(server.ListenAndServe())
}
