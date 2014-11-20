package main

import (
	"code.google.com/p/go-uuid/uuid"
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
	Id      int `orm:"auto"`
	Path    string
	Created string `orm:"auto_now_add;type(datetime)"`
	Case    *Case  `orm:"rel(fk)"`
}

type Case struct {
	Id          int       `orm:"auto"`
	Description string    `orm:"default(""), type(text)"`
	Created     time.Time `orm:"auto_now_add;type(datetime)"`
	Updated     time.Time `orm:"auto_now_add;type(datetime)"`
	IsSigned    bool      `orm:"default(false)"`
	IsPrivate   bool      `orm:"default(false)"`
	Token       string
	Config      string  `orm:"default(""), type(text)"`
	Signed      string  `orm:"default(""), type(text)"`
	Files       []*File `orm:"reverse(many)"`
}

type CaseHandler struct{}

func (handler *CaseHandler) Create(request *restful.Request, response *restful.Response) {
	c := new(Case)
	err := request.ReadEntity(c)

	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if c.IsPrivate {
		c.Token = uuid.New()
	}

	if c.Signed != "" {
		c.IsSigned = true
	}

	o := orm.NewOrm()
	o.Insert(c)

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(c)
}

func (handler *CaseHandler) Get(request *restful.Request, response *restful.Response) {
	// id := request.PathParameter("case-id")
	// c := Case{Id: id}

	// o := orm.NewOrm()
	// err := o.Read(&c)
	// if err {
	// 	response.WriteErrorString(http.StatusInternalServerError, err.Error())
	// }

	// response.WriteEntity(Case{})
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

	ws.Route(ws.GET("/{case-id}").To(handler.Get).
		Doc("get a specific report").
		Operation("findCase").
		Param(ws.PathParameter("case-id", "case identifier").DataType("int")).
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
